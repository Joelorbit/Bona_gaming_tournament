package payment

import (
	"encoding/json"
	"net/url"
	"testing"

	"bona-backend/internal/repository"
	addispayclient "bona-backend/pkg/addispay"
)

func TestCustomerNameUsesDisplayName(t *testing.T) {
	displayName := "Abebe Kebede"
	first, last := customerName(repository.Profile{
		Username:    "user_12345678",
		DisplayName: &displayName,
	})

	if first != "Abebe" || last != "Kebede" {
		t.Fatalf("customerName() = %q %q, want Abebe Kebede", first, last)
	}
}

func TestCustomerNameSanitizesInvalidUsername(t *testing.T) {
	first, last := customerName(repository.Profile{
		Username: "user_12345678",
	})

	if first != "User" || last != "Player" {
		t.Fatalf("customerName() = %q %q, want User Player", first, last)
	}
}

func TestCustomerNameFallsBackToEmailTokens(t *testing.T) {
	email := "abebe.kebede92@example.com"
	first, last := customerName(repository.Profile{
		Username: "x_1",
		Email:    &email,
	})

	if first != "Abebe" || last != "Kebede" {
		t.Fatalf("customerName() = %q %q, want Abebe Kebede", first, last)
	}
}

func TestCustomerNameDropsShortAndNumericTokens(t *testing.T) {
	displayName := "A 123 B"
	first, last := customerName(repository.Profile{
		Username:    "p9",
		DisplayName: &displayName,
	})

	if first != "Bona" || last != "Player" {
		t.Fatalf("customerName() = %q %q, want Bona Player", first, last)
	}
}

func TestPaymentReturnTokenValidatesOriginalPaymentDetails(t *testing.T) {
	payment := repository.Payment{
		ID:           "payment-123",
		UserID:       "user-123",
		TournamentID: "tournament-123",
		Amount:       250,
	}
	token := paymentReturnToken(payment, "return-secret", "success")

	if token == "" {
		t.Fatal("paymentReturnToken() returned empty token")
	}
	if !validPaymentReturnToken(payment, "return-secret", token, "success") {
		t.Fatal("validPaymentReturnToken() = false, want true")
	}
	if !validPaymentReturnToken(payment, "return-secret", token, " SUCCESS ") {
		t.Fatal("validPaymentReturnToken() with trimmed/case-changed status = false, want true")
	}

	tampered := payment
	tampered.Amount++
	if validPaymentReturnToken(tampered, "return-secret", token, "success") {
		t.Fatal("validPaymentReturnToken() accepted token after amount tamper")
	}
	if validPaymentReturnToken(payment, "wrong-secret", token, "success") {
		t.Fatal("validPaymentReturnToken() accepted token with wrong secret")
	}
	if validPaymentReturnToken(payment, "return-secret", token, "failed") {
		t.Fatal("validPaymentReturnToken() accepted token with wrong status")
	}
}

func TestPaymentReturnURLBuildsSignedFrontendReturn(t *testing.T) {
	service := &Service{
		checkout: CheckoutConfig{WebhookSecret: "return-secret"},
	}
	payment := repository.Payment{
		ID:           "payment-123",
		UserID:       "user-123",
		TournamentID: "tournament-123",
		Amount:       250,
	}

	got := service.paymentReturnURL("https://app.example.com/old/path?keep=yes", payment, "success")
	parsed, err := url.Parse(got)
	if err != nil {
		t.Fatalf("url.Parse() error = %v", err)
	}

	if parsed.Scheme != "https" || parsed.Host != "app.example.com" {
		t.Fatalf("return URL host = %s://%s, want https://app.example.com", parsed.Scheme, parsed.Host)
	}
	if parsed.Path != "/payments/return" {
		t.Fatalf("return URL path = %q, want /payments/return", parsed.Path)
	}

	q := parsed.Query()
	if q.Get("payment_id") != payment.ID {
		t.Fatalf("payment_id = %q, want %q", q.Get("payment_id"), payment.ID)
	}
	if q.Get("tournament_id") != payment.TournamentID {
		t.Fatalf("tournament_id = %q, want %q", q.Get("tournament_id"), payment.TournamentID)
	}
	if q.Get("status") != "success" {
		t.Fatalf("status = %q, want success", q.Get("status"))
	}
	if !validPaymentReturnToken(payment, "return-secret", q.Get("token"), "success") {
		t.Fatal("return URL token is not valid for payment")
	}
	if q.Get("keep") != "yes" {
		t.Fatalf("existing query keep = %q, want yes", q.Get("keep"))
	}
}

func TestPaymentReturnURLWithoutSecretOmitsToken(t *testing.T) {
	service := &Service{}
	payment := repository.Payment{
		ID:           "payment-123",
		UserID:       "user-123",
		TournamentID: "tournament-123",
		Amount:       250,
	}

	got := service.paymentReturnURL("https://app.example.com", payment, "success")
	parsed, err := url.Parse(got)
	if err != nil {
		t.Fatalf("url.Parse() error = %v", err)
	}
	if parsed.Query().Get("token") != "" {
		t.Fatalf("token = %q, want empty", parsed.Query().Get("token"))
	}
}

func TestGatewayPaymentStatus(t *testing.T) {
	tests := []struct {
		status string
		want   string
	}{
		{status: "completed", want: "paid"},
		{status: " SUCCESS ", want: "paid"},
		{status: "approved", want: "paid"},
		{status: "captured", want: "paid"},
		{status: "failed", want: "failed"},
		{status: "cancelled", want: "failed"},
		{status: "expired", want: "failed"},
		{status: "processing", want: ""},
		{status: "", want: ""},
	}

	for _, tt := range tests {
		t.Run(tt.status, func(t *testing.T) {
			if got := gatewayPaymentStatus(tt.status); got != tt.want {
				t.Fatalf("gatewayPaymentStatus(%q) = %q, want %q", tt.status, got, tt.want)
			}
		})
	}
}

func TestWebhookMetadataCapturesProviderFields(t *testing.T) {
	got := webhookMetadata(addispayclient.WebhookPayload{
		Reference: "payment-123",
		Status:    "paid",
		Amount:    250,
		Currency:  "ETB",
		PaymentID: "provider-123",
	})

	var metadata map[string]any
	if err := json.Unmarshal(got, &metadata); err != nil {
		t.Fatalf("json.Unmarshal() error = %v", err)
	}
	if metadata["reference"] != "payment-123" {
		t.Fatalf("metadata reference = %v, want payment-123", metadata["reference"])
	}
	if metadata["status"] != "paid" {
		t.Fatalf("metadata status = %v, want paid", metadata["status"])
	}
	if metadata["currency"] != "ETB" {
		t.Fatalf("metadata currency = %v, want ETB", metadata["currency"])
	}
	if metadata["payment_id"] != "provider-123" {
		t.Fatalf("metadata payment_id = %v, want provider-123", metadata["payment_id"])
	}
}
