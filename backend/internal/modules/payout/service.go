package payout

import (
	"context"
	"fmt"
	"strings"

	"bona-backend/internal/modules/admin"
	"bona-backend/internal/modules/notification"
	"bona-backend/internal/repository"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Service struct {
	repo     *repository.Queries
	db       *pgxpool.Pool
	notifier *notification.Service
	auditor  *admin.Service
}

func NewService(db *pgxpool.Pool, notifier *notification.Service, auditor *admin.Service) *Service {
	return &Service{
		repo:     repository.New(db),
		db:       db,
		notifier: notifier,
		auditor:  auditor,
	}
}

// EnsureForTournament computes the winner prize and creates the payout row
// when a tournament transitions to completed. Idempotent — calling twice is safe.
func (s *Service) EnsureForTournament(ctx context.Context, tournamentID string) (*repository.Payout, error) {
	t, err := s.repo.GetTournament(ctx, tournamentID)
	if err != nil {
		return nil, fmt.Errorf("tournament not found")
	}
	if t.Status != "completed" {
		return nil, fmt.Errorf("tournament is not completed")
	}

	winnerID, err := s.findWinner(ctx, tournamentID)
	if err != nil {
		return nil, err
	}
	if winnerID == "" {
		return nil, fmt.Errorf("no winner yet")
	}

	paidCount, err := s.repo.CountPaidRegistrations(ctx, tournamentID)
	if err != nil {
		return nil, fmt.Errorf("count paid: %w", err)
	}
	collected := int64(t.EntryFee) * paidCount
	platformCut := collected * int64(t.PlatformFeePct) / 100
	organizerCut := collected * int64(t.OrganizerFeePct) / 100
	winnerPrize := collected - platformCut - organizerCut
	if winnerPrize < 0 {
		winnerPrize = 0
	}
	// If the organizer set a prize_pool higher than what entry fees collected,
	// honour the larger number — that's the advertised pool.
	if int64(t.PrizePool) > winnerPrize {
		winnerPrize = int64(t.PrizePool)
	}

	payout, err := s.repo.CreatePayout(ctx, repository.CreatePayoutParams{
		TournamentID: tournamentID,
		WinnerID:     winnerID,
		Amount:       int32(winnerPrize),
		Currency:     t.Currency,
		Status:       "pending",
	})
	if err != nil {
		return nil, fmt.Errorf("create payout: %w", err)
	}

	if s.notifier != nil {
		link := fmt.Sprintf("/me/payouts")
		s.notifier.Emit(ctx, repository.CreateNotificationParams{
			UserID:  winnerID,
			Type:    "tournament_won",
			Title:   "You won " + t.Title + "!",
			Message: fmt.Sprintf("Prize: %d %s. Submit payout details so the organizer can pay you.", payout.Amount, payout.Currency),
			Link:    &link,
		})
		s.notifier.Emit(ctx, repository.CreateNotificationParams{
			UserID:  t.OrganizerID,
			Type:    "payout_due",
			Title:   "Payout due for " + t.Title,
			Message: fmt.Sprintf("Winner is owed %d %s.", payout.Amount, payout.Currency),
			Link:    &link,
		})
	}

	return &payout, nil
}

// findWinner returns the winner_id of the highest-round completed match.
func (s *Service) findWinner(ctx context.Context, tournamentID string) (string, error) {
	maxRound, err := s.repo.GetMaxRound(ctx, tournamentID)
	if err != nil {
		return "", err
	}
	if maxRound == 0 {
		return "", nil
	}
	finals, err := s.repo.GetMatchesByTournamentAndRound(ctx, repository.GetMatchesByTournamentAndRoundParams{
		TournamentID: tournamentID,
		Round:        maxRound,
	})
	if err != nil {
		return "", err
	}
	for _, m := range finals {
		if m.WinnerID != nil && m.Status == "completed" {
			return *m.WinnerID, nil
		}
	}
	return "", nil
}

func (s *Service) MarkPaid(ctx context.Context, payoutID, organizerID, note string) (*repository.Payout, error) {
	p, err := s.repo.GetPayout(ctx, payoutID)
	if err != nil {
		return nil, fmt.Errorf("payout not found")
	}
	t, err := s.repo.GetTournament(ctx, p.TournamentID)
	if err != nil {
		return nil, fmt.Errorf("tournament not found")
	}
	if t.OrganizerID != organizerID {
		return nil, fmt.Errorf("only the organizer can mark payouts paid")
	}
	if p.Status == "paid" {
		return nil, fmt.Errorf("payout already paid")
	}
	if p.PayoutDetailsSubmittedAt == nil {
		return nil, fmt.Errorf("winner has not submitted payout details")
	}

	var notePtr *string
	if note != "" {
		notePtr = &note
	}

	updated, err := s.repo.MarkPayoutPaid(ctx, repository.MarkPayoutPaidParams{
		ID:     payoutID,
		PaidBy: organizerID,
		Note:   notePtr,
	})
	if err != nil {
		return nil, fmt.Errorf("mark paid: %w", err)
	}

	if s.notifier != nil {
		link := "/me/payouts"
		s.notifier.Emit(ctx, repository.CreateNotificationParams{
			UserID:  updated.WinnerID,
			Type:    "payout_paid",
			Title:   "Prize paid",
			Message: fmt.Sprintf("Your prize of %d %s has been paid.", updated.Amount, updated.Currency),
			Link:    &link,
		})
	}

	if s.auditor != nil {
		pid := updated.ID
		s.auditor.Log(ctx, organizerID, "organizer", "payout.marked_paid", "payout", &pid, map[string]any{
			"amount":   updated.Amount,
			"currency": updated.Currency,
			"winner":   updated.WinnerID,
		})
	}

	return &updated, nil
}

type SubmitDetailsParams struct {
	PayoutMethod       string  `json:"payout_method"`
	PhoneNumber        *string `json:"phone_number,omitempty"`
	TelebirrNumber     *string `json:"telebirr_number,omitempty"`
	BankName           *string `json:"bank_name,omitempty"`
	BankAccountName    *string `json:"bank_account_name,omitempty"`
	BankAccountNumber  *string `json:"bank_account_number,omitempty"`
	PayoutInstructions *string `json:"payout_instructions,omitempty"`
}

func (s *Service) SubmitDetails(ctx context.Context, payoutID, winnerID string, params SubmitDetailsParams) (*repository.Payout, error) {
	p, err := s.repo.GetPayout(ctx, payoutID)
	if err != nil {
		return nil, fmt.Errorf("payout not found")
	}
	if p.WinnerID != winnerID {
		return nil, fmt.Errorf("only the winner can submit payout details")
	}
	if p.Status == "paid" {
		return nil, fmt.Errorf("payout is already paid")
	}

	method := strings.ToLower(strings.TrimSpace(params.PayoutMethod))
	phone := cleanString(params.PhoneNumber)
	telebirr := cleanString(params.TelebirrNumber)
	bankName := cleanString(params.BankName)
	bankAccountName := cleanString(params.BankAccountName)
	bankAccountNumber := cleanString(params.BankAccountNumber)
	instructions := cleanString(params.PayoutInstructions)

	switch method {
	case "telebirr":
		if telebirr == nil {
			return nil, fmt.Errorf("telebirr number is required")
		}
		bankName = nil
		bankAccountName = nil
		bankAccountNumber = nil
		instructions = nil
	case "bank":
		if bankName == nil || bankAccountName == nil || bankAccountNumber == nil {
			return nil, fmt.Errorf("bank name, account name, and account number are required")
		}
		telebirr = nil
		instructions = nil
	case "other":
		if instructions == nil {
			return nil, fmt.Errorf("payment instructions are required")
		}
		telebirr = nil
		bankName = nil
		bankAccountName = nil
		bankAccountNumber = nil
	default:
		return nil, fmt.Errorf("payout method must be telebirr, bank, or other")
	}

	updated, err := s.repo.UpdatePayoutDetails(ctx, repository.UpdatePayoutDetailsParams{
		ID:                 payoutID,
		PayoutMethod:       method,
		PhoneNumber:        phone,
		TelebirrNumber:     telebirr,
		BankName:           bankName,
		BankAccountName:    bankAccountName,
		BankAccountNumber:  bankAccountNumber,
		PayoutInstructions: instructions,
	})
	if err != nil {
		return nil, fmt.Errorf("save payout details: %w", err)
	}

	if s.notifier != nil {
		t, err := s.repo.GetTournament(ctx, updated.TournamentID)
		if err == nil {
			link := "/dashboard"
			s.notifier.Emit(ctx, repository.CreateNotificationParams{
				UserID:  t.OrganizerID,
				Type:    "payout_details_submitted",
				Title:   "Payout details submitted",
				Message: fmt.Sprintf("The winner submitted payout details for %s.", t.Title),
				Link:    &link,
			})
		}
	}

	return &updated, nil
}

func cleanString(value *string) *string {
	if value == nil {
		return nil
	}
	trimmed := strings.TrimSpace(*value)
	if trimmed == "" {
		return nil
	}
	return &trimmed
}

func (s *Service) ListByTournament(ctx context.Context, tournamentID, organizerID string) ([]repository.Payout, error) {
	t, err := s.repo.GetTournament(ctx, tournamentID)
	if err != nil {
		return nil, fmt.Errorf("tournament not found")
	}
	if t.OrganizerID != organizerID {
		return nil, fmt.Errorf("only the organizer can list tournament payouts")
	}
	return s.repo.ListPayoutsByTournament(ctx, tournamentID)
}

func (s *Service) ListByWinner(ctx context.Context, winnerID string) ([]repository.Payout, error) {
	if err := s.ensureMissingWinnerPayouts(ctx, winnerID); err != nil {
		return nil, err
	}
	return s.repo.ListPayoutsByWinner(ctx, winnerID)
}

func (s *Service) ListByOrganizer(ctx context.Context, organizerID string) ([]repository.Payout, error) {
	return s.repo.ListPayoutsByOrganizer(ctx, organizerID)
}

func (s *Service) ensureMissingWinnerPayouts(ctx context.Context, winnerID string) error {
	tournamentIDs, err := s.repo.ListCompletedWonTournamentIDsMissingPayout(ctx, winnerID)
	if err != nil {
		return fmt.Errorf("find missing winner payouts: %w", err)
	}
	for _, tournamentID := range tournamentIDs {
		if _, err := s.EnsureForTournament(ctx, tournamentID); err != nil {
			return fmt.Errorf("create missing payout for tournament %s: %w", tournamentID, err)
		}
	}
	return nil
}
