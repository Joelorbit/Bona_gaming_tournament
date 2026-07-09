package payment

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/url"
	"strconv"
	"strings"

	"bona-backend/internal/modules/notification"
	"bona-backend/internal/repository"
	addispayclient "bona-backend/pkg/addispay"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Service struct {
	repo           *repository.Queries
	db             *pgxpool.Pool
	addispayClient *addispayclient.Client
	checkout       CheckoutConfig
	notifier       *notification.Service
}

type CheckoutConfig struct {
	WebhookSecret string
	RedirectURL   string
	CancelURL     string
	SuccessURL    string
	ErrorURL      string
}

func NewService(db *pgxpool.Pool, client *addispayclient.Client, checkout CheckoutConfig, notifier *notification.Service) *Service {
	return &Service{
		repo:           repository.New(db),
		db:             db,
		addispayClient: client,
		checkout:       checkout,
		notifier:       notifier,
	}
}

type PaymentResult struct {
	Payment    *repository.Payment `json:"payment"`
	PaymentURL string              `json:"payment_url,omitempty"`
}

func (s *Service) CreatePayment(ctx context.Context, userID, tournamentID string) (*PaymentResult, error) {
	tournament, err := s.repo.GetTournament(ctx, tournamentID)
	if err != nil {
		return nil, fmt.Errorf("tournament not found")
	}

	if tournament.EntryFee <= 0 {
		return nil, fmt.Errorf("tournament is free, no payment required")
	}

	existingReg, err := s.repo.GetRegistrationByUserAndTournament(ctx, repository.GetRegistrationByUserAndTournamentParams{
		UserID:       userID,
		TournamentID: tournamentID,
	})
	if err != nil {
		return nil, fmt.Errorf("you must register before paying")
	}
	if existingReg.PaymentStatus == "paid" {
		return nil, fmt.Errorf("payment already completed")
	}
	if s.addispayClient == nil {
		return nil, fmt.Errorf("payment gateway is not configured")
	}

	profile, err := s.repo.GetProfile(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("profile not found")
	}

	payment, err := s.repo.CreatePayment(ctx, repository.CreatePaymentParams{
		UserID:       userID,
		TournamentID: tournamentID,
		Amount:       tournament.EntryFee,
		Currency:     "ETB",
		Status:       "pending",
	})
	if err != nil {
		return nil, fmt.Errorf("create payment: %w", err)
	}

	result := &PaymentResult{
		Payment: &payment,
	}

	description := fmt.Sprintf("Tournament entry: %s", tournament.Title)
	customerFirstName, customerLastName := customerName(profile)
	successURL := s.paymentReturnURL(firstNonEmpty(s.checkout.SuccessURL, s.checkout.RedirectURL), payment, "success")
	errorURL := s.paymentReturnURL(firstNonEmpty(s.checkout.ErrorURL, s.checkout.RedirectURL), payment, "failed")
	cancelURL := s.paymentReturnURL(firstNonEmpty(s.checkout.CancelURL, s.checkout.RedirectURL), payment, "cancelled")
	addisReq := addispayclient.PaymentRequest{
		Message: "Bona tournament entry payment",
		Data: addispayclient.OrderData{
			RedirectURL:    successURL,
			CancelURL:      cancelURL,
			SuccessURL:     successURL,
			ErrorURL:       errorURL,
			OrderReason:    description,
			Currency:       payment.Currency,
			Email:          stringValue(profile.Email, "player@bona.local"),
			FirstName:      customerFirstName,
			LastName:       customerLastName,
			Nonce:          payment.ID,
			PhoneNumber:    "",
			SessionExpired: "5000",
			TotalAmount:    strconv.Itoa(int(payment.Amount)),
			TxRef:          payment.ID,
			OrderDetail: addispayclient.OrderDetail{
				Amount:      int(payment.Amount),
				Description: description,
			},
		},
	}

	addisResp, err := s.addispayClient.CreatePayment(addisReq)
	if err != nil {
		reason := err.Error()
		_, _ = s.repo.MarkPaymentGatewayFailed(ctx, repository.MarkPaymentGatewayFailedParams{
			ID:            payment.ID,
			FailureReason: &reason,
		})
		return nil, fmt.Errorf("start AddisPay checkout: %w", err)
	}

	result.PaymentURL = addisResp.HostedCheckoutURL()
	if result.PaymentURL == "" {
		reason := "AddisPay did not return a checkout URL"
		_, _ = s.repo.MarkPaymentGatewayFailed(ctx, repository.MarkPaymentGatewayFailedParams{
			ID:            payment.ID,
			FailureReason: &reason,
		})
		return nil, fmt.Errorf("%s", reason)
	}

	ref := payment.ID
	updated, err := s.repo.UpdatePaymentGatewayFields(ctx, repository.UpdatePaymentGatewayFieldsParams{
		ID:          payment.ID,
		AddispayRef: ref,
		PaymentURL:  result.PaymentURL,
	})
	if err == nil {
		result.Payment = &updated
	}

	return result, nil
}

type PaymentReturnResult struct {
	Payment       *repository.Payment `json:"payment"`
	TournamentID  string              `json:"tournament_id"`
	PaymentStatus string              `json:"payment_status"`
}

func (s *Service) ConfirmReturn(ctx context.Context, userID, paymentID, status, token string) (*PaymentReturnResult, error) {
	if strings.TrimSpace(paymentID) == "" {
		return nil, fmt.Errorf("payment_id is required")
	}

	payment, err := s.repo.GetPayment(ctx, paymentID)
	if err != nil || payment.UserID != userID {
		return nil, fmt.Errorf("payment not found")
	}
	if !validPaymentReturnToken(payment, s.checkout.WebhookSecret, token, status) {
		return nil, fmt.Errorf("invalid payment return token")
	}

	newStatus := gatewayPaymentStatus(status)
	if newStatus == "" {
		return nil, fmt.Errorf("payment return status unsupported")
	}

	var failureReason *string
	if newStatus == "failed" {
		reason := fmt.Sprintf("AddisPay returned %s", strings.ToLower(strings.TrimSpace(status)))
		failureReason = &reason
	}
	payload := addispayclient.WebhookPayload{
		Reference: payment.ID,
		Status:    status,
		Amount:    int(payment.Amount),
		Currency:  payment.Currency,
	}
	updated, err := s.completePaymentFromGateway(ctx, payment, payload, newStatus, failureReason, false)
	if err != nil {
		return nil, err
	}

	return &PaymentReturnResult{
		Payment:       &updated,
		TournamentID:  updated.TournamentID,
		PaymentStatus: updated.Status,
	}, nil
}

func (s *Service) HandleWebhook(ctx context.Context, payload addispayclient.WebhookPayload) error {
	if payload.Reference == "" {
		return fmt.Errorf("missing reference")
	}

	payment, err := s.repo.GetPaymentByAddispayRef(ctx, payload.Reference)
	if err != nil {
		payment, err = s.repo.GetPayment(ctx, payload.Reference)
		if err != nil {
			return fmt.Errorf("payment not found")
		}
	}

	if int32(payload.Amount) != payment.Amount {
		s.recordWebhookMismatch(ctx, payment.ID, payload, "payment amount mismatch")
		return fmt.Errorf("payment amount mismatch")
	}
	if payload.Currency != "" && !strings.EqualFold(payload.Currency, payment.Currency) {
		s.recordWebhookMismatch(ctx, payment.ID, payload, "payment currency mismatch")
		return fmt.Errorf("payment currency mismatch")
	}

	newStatus := gatewayPaymentStatus(payload.Status)
	if newStatus == "" {
		s.recordWebhookMismatch(ctx, payment.ID, payload, "payment status unsupported")
		return fmt.Errorf("payment status unsupported")
	}
	_, err = s.completePaymentFromGateway(ctx, payment, payload, newStatus, nil, true)
	return err
}

func (s *Service) completePaymentFromGateway(ctx context.Context, payment repository.Payment, payload addispayclient.WebhookPayload, status string, failureReason *string, fromWebhook bool) (repository.Payment, error) {
	alreadyPaid := payment.Status == "paid"
	alreadyFailed := payment.Status == "failed"
	metadata := webhookMetadata(payload)
	providerPaymentID := cleanString(payload.PaymentID)
	providerStatus := strings.TrimSpace(payload.Status)
	if providerStatus == "" {
		providerStatus = status
	}

	tx, err := s.db.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return repository.Payment{}, fmt.Errorf("begin payment transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	q := repository.New(tx)
	var updated repository.Payment
	if fromWebhook {
		updated, err = q.CompletePaymentFromWebhook(ctx, repository.CompletePaymentFromWebhookParams{
			ID:                payment.ID,
			Status:            status,
			ProviderStatus:    providerStatus,
			ProviderPaymentID: providerPaymentID,
			FailureReason:     failureReason,
			Metadata:          metadata,
		})
	} else {
		updated, err = q.CompletePaymentFromReturn(ctx, repository.CompletePaymentFromReturnParams{
			ID:                payment.ID,
			Status:            status,
			ProviderStatus:    providerStatus,
			ProviderPaymentID: providerPaymentID,
			FailureReason:     failureReason,
			Metadata:          metadata,
		})
	}
	if err != nil {
		return updated, fmt.Errorf("complete payment: %w", err)
	}

	if updated.Status == "paid" || updated.Status == "failed" {
		_, err = q.UpdateRegistrationPaymentStatus(ctx, repository.UpdateRegistrationPaymentStatusParams{
			UserID:        payment.UserID,
			TournamentID:  payment.TournamentID,
			PaymentStatus: updated.Status,
		})
		if err != nil {
			return updated, fmt.Errorf("update registration: %w", err)
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return updated, fmt.Errorf("commit payment transaction: %w", err)
	}

	if updated.Status == "paid" && !alreadyPaid && s.notifier != nil {
		tournament, terr := s.repo.GetTournament(ctx, payment.TournamentID)
		title := "Tournament"
		if terr == nil {
			title = tournament.Title
		}
		link := fmt.Sprintf("/tournaments/%s", payment.TournamentID)
		s.notifier.Emit(ctx, repository.CreateNotificationParams{
			UserID:  payment.UserID,
			Type:    "payment_confirmed",
			Title:   "Payment confirmed",
			Message: fmt.Sprintf("Your entry to %s is paid and confirmed.", title),
			Link:    &link,
		})
	}

	if updated.Status == "failed" && !alreadyFailed && !alreadyPaid && s.notifier != nil {
		link := fmt.Sprintf("/tournaments/%s", payment.TournamentID)
		s.notifier.Emit(ctx, repository.CreateNotificationParams{
			UserID:  payment.UserID,
			Type:    "payment_failed",
			Title:   "Payment failed",
			Message: "Your payment could not be processed. Please try again.",
			Link:    &link,
		})
	}

	return updated, nil
}

func (s *Service) recordWebhookMismatch(ctx context.Context, paymentID string, payload addispayclient.WebhookPayload, reason string) {
	providerStatus := strings.TrimSpace(payload.Status)
	if providerStatus == "" {
		providerStatus = "invalid"
	}
	_, _ = s.repo.CompletePaymentFromWebhook(ctx, repository.CompletePaymentFromWebhookParams{
		ID:                paymentID,
		Status:            "failed",
		ProviderStatus:    providerStatus,
		ProviderPaymentID: cleanString(payload.PaymentID),
		FailureReason:     &reason,
		Metadata:          webhookMetadata(payload),
	})
}

func (s *Service) GetPayment(ctx context.Context, id string) (*repository.Payment, error) {
	payment, err := s.repo.GetPayment(ctx, id)
	if err != nil {
		return nil, err
	}
	return &payment, nil
}

func (s *Service) ListByOrganizer(ctx context.Context, organizerID string) ([]repository.Payment, error) {
	return s.repo.ListPaymentsByOrganizer(ctx, organizerID)
}

func (s *Service) MarkRefunded(ctx context.Context, paymentID, organizerID string) (*repository.Payment, error) {
	payment, err := s.repo.GetPayment(ctx, paymentID)
	if err != nil {
		return nil, fmt.Errorf("payment not found")
	}
	tournament, err := s.repo.GetTournament(ctx, payment.TournamentID)
	if err != nil {
		return nil, fmt.Errorf("tournament not found")
	}
	if tournament.OrganizerID != organizerID {
		return nil, fmt.Errorf("only the tournament organizer can mark refunds")
	}
	if tournament.Status != "cancelled" {
		return nil, fmt.Errorf("refunds can only be marked after tournament cancellation")
	}
	if payment.RefundStatus != "pending" {
		return nil, fmt.Errorf("payment is not pending refund")
	}

	tx, err := s.db.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return nil, fmt.Errorf("begin refund transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	q := repository.New(tx)
	updated, err := q.MarkPaymentRefunded(ctx, repository.MarkPaymentRefundedParams{
		ID:         paymentID,
		RefundedBy: organizerID,
	})
	if err != nil {
		return nil, fmt.Errorf("mark refunded: %w", err)
	}
	_, err = q.UpdateRegistrationPaymentStatus(ctx, repository.UpdateRegistrationPaymentStatusParams{
		UserID:        updated.UserID,
		TournamentID:  updated.TournamentID,
		PaymentStatus: "refunded",
	})
	if err != nil {
		return nil, fmt.Errorf("update registration refund status: %w", err)
	}
	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("commit refund transaction: %w", err)
	}

	if s.notifier != nil {
		link := fmt.Sprintf("/tournaments/%s", updated.TournamentID)
		s.notifier.Emit(ctx, repository.CreateNotificationParams{
			UserID:  updated.UserID,
			Type:    "refund_paid",
			Title:   "Refund marked paid",
			Message: fmt.Sprintf("Your %d %s entry refund for %s was marked paid.", updated.Amount, updated.Currency, tournament.Title),
			Link:    &link,
		})
	}

	return &updated, nil
}

func stringValue(v *string, fallback string) string {
	if v != nil && strings.TrimSpace(*v) != "" {
		return strings.TrimSpace(*v)
	}
	return fallback
}

func cleanString(value string) *string {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return nil
	}
	return &trimmed
}

func webhookMetadata(payload addispayclient.WebhookPayload) json.RawMessage {
	b, err := json.Marshal(map[string]any{
		"reference":  payload.Reference,
		"status":     payload.Status,
		"amount":     payload.Amount,
		"currency":   payload.Currency,
		"payment_id": payload.PaymentID,
	})
	if err != nil {
		return nil
	}
	return b
}

func customerName(profile repository.Profile) (string, string) {
	tokens := customerNameTokens(stringValue(profile.DisplayName, ""))
	if len(tokens) < 2 {
		tokens = append(tokens, customerNameTokens(profile.Username)...)
	}
	if len(tokens) < 2 && profile.Email != nil {
		localPart := strings.SplitN(*profile.Email, "@", 2)[0]
		tokens = append(tokens, customerNameTokens(localPart)...)
	}

	switch len(tokens) {
	case 0:
		return "Bona", "Player"
	case 1:
		return tokens[0], "Player"
	default:
		return tokens[0], tokens[1]
	}
}

func customerNameTokens(raw string) []string {
	var tokens []string
	var b strings.Builder
	flush := func() {
		if b.Len() == 0 {
			return
		}
		token := titleASCII(b.String())
		b.Reset()
		if len(token) >= 2 {
			tokens = append(tokens, token)
		}
	}

	for _, r := range raw {
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') {
			b.WriteRune(r)
			continue
		}
		flush()
	}
	flush()

	return tokens
}

func titleASCII(s string) string {
	s = strings.ToLower(s)
	if s == "" {
		return s
	}
	return strings.ToUpper(s[:1]) + s[1:]
}

func (s *Service) paymentReturnURL(base string, payment repository.Payment, status string) string {
	base = strings.TrimSpace(base)
	if base == "" {
		return ""
	}
	u, err := url.Parse(base)
	if err != nil || u.Scheme == "" || u.Host == "" {
		return base
	}
	u.Path = "/payments/return"
	q := u.Query()
	q.Set("payment_id", payment.ID)
	q.Set("tournament_id", payment.TournamentID)
	q.Set("status", status)
	if token := paymentReturnToken(payment, s.checkout.WebhookSecret, status); token != "" {
		q.Set("token", token)
	}
	u.RawQuery = q.Encode()
	return u.String()
}

func paymentReturnToken(payment repository.Payment, secret, status string) string {
	secret = strings.TrimSpace(secret)
	if secret == "" {
		return ""
	}
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(payment.ID))
	mac.Write([]byte("|"))
	mac.Write([]byte(payment.UserID))
	mac.Write([]byte("|"))
	mac.Write([]byte(payment.TournamentID))
	mac.Write([]byte("|"))
	mac.Write([]byte(strconv.Itoa(int(payment.Amount))))
	mac.Write([]byte("|"))
	mac.Write([]byte(strings.ToLower(strings.TrimSpace(status))))
	return hex.EncodeToString(mac.Sum(nil))
}

func validPaymentReturnToken(payment repository.Payment, secret, token, status string) bool {
	expected := paymentReturnToken(payment, secret, status)
	if expected == "" || strings.TrimSpace(token) == "" {
		return false
	}
	return hmac.Equal([]byte(expected), []byte(strings.TrimSpace(token)))
}

func gatewayPaymentStatus(status string) string {
	switch strings.ToLower(strings.TrimSpace(status)) {
	case "completed", "complete", "success", "successful", "paid", "approved", "captured":
		return "paid"
	case "failed", "failure", "declined", "cancelled", "canceled", "expired", "error":
		return "failed"
	default:
		return ""
	}
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return value
		}
	}
	return ""
}
