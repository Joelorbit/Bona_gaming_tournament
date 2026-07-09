package addispay

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"testing"
)

func TestHostedCheckoutURL(t *testing.T) {
	resp := &PaymentResponse{
		CheckoutURL: "https://checkout.addispay.et/pay/",
		UUID:        "order-uuid",
	}

	got := resp.HostedCheckoutURL()
	want := "https://checkout.addispay.et/pay/order-uuid"
	if got != want {
		t.Fatalf("HostedCheckoutURL() = %q, want %q", got, want)
	}
}

func TestHostedCheckoutURLEmptyWhenIncomplete(t *testing.T) {
	resp := &PaymentResponse{CheckoutURL: "not-a-url"}

	if got := resp.HostedCheckoutURL(); got != "" {
		t.Fatalf("HostedCheckoutURL() = %q, want empty string", got)
	}
}

func TestHostedCheckoutURLUsesDirectCheckoutURL(t *testing.T) {
	resp := &PaymentResponse{CheckoutURL: "https://checkout.addispay.et/pay/order-uuid"}

	got := resp.HostedCheckoutURL()
	want := "https://checkout.addispay.et/pay/order-uuid"
	if got != want {
		t.Fatalf("HostedCheckoutURL() = %q, want %q", got, want)
	}
}

func TestHostedCheckoutURLResolvesRelativeCheckoutURLAgainstBaseURL(t *testing.T) {
	resp := &PaymentResponse{
		CheckoutURL: "/checkout/order",
		UUID:        "order-uuid",
		baseURL:     "https://uat.api.addispay.et",
	}

	got := resp.HostedCheckoutURL()
	want := "https://uat.api.addispay.et/checkout/order/order-uuid"
	if got != want {
		t.Fatalf("HostedCheckoutURL() = %q, want %q", got, want)
	}
}

func TestPaymentResponseUnmarshalWrappedData(t *testing.T) {
	var resp PaymentResponse
	body := []byte(`{"status":"success","data":{"checkout_url":"https://checkout.addispay.et/pay","uuid":"order-uuid"}}`)
	if err := json.Unmarshal(body, &resp); err != nil {
		t.Fatalf("Unmarshal() error = %v", err)
	}

	got := resp.HostedCheckoutURL()
	want := "https://checkout.addispay.et/pay/order-uuid"
	if got != want {
		t.Fatalf("HostedCheckoutURL() = %q, want %q", got, want)
	}
}

func TestPaymentResponseUnmarshalHostedCheckoutURL(t *testing.T) {
	var resp PaymentResponse
	body := []byte(`{"data":{"hosted_checkout_url":"https://checkout.addispay.et/pay/order-uuid"}}`)
	if err := json.Unmarshal(body, &resp); err != nil {
		t.Fatalf("Unmarshal() error = %v", err)
	}

	got := resp.HostedCheckoutURL()
	want := "https://checkout.addispay.et/pay/order-uuid"
	if got != want {
		t.Fatalf("HostedCheckoutURL() = %q, want %q", got, want)
	}
}

func TestWebhookPayloadUnmarshalWrappedFields(t *testing.T) {
	var payload WebhookPayload
	body := []byte(`{
		"data": {
			"tx_ref": "local-payment-id",
			"payment_status": "PAID",
			"total_amount": "10",
			"currency": "ETB",
			"signature": "abc123",
			"uuid": "gateway-payment-id"
		}
	}`)
	if err := json.Unmarshal(body, &payload); err != nil {
		t.Fatalf("Unmarshal() error = %v", err)
	}

	if payload.Reference != "local-payment-id" {
		t.Fatalf("Reference = %q, want local-payment-id", payload.Reference)
	}
	if payload.Status != "PAID" {
		t.Fatalf("Status = %q, want PAID", payload.Status)
	}
	if payload.Amount != 10 {
		t.Fatalf("Amount = %d, want 10", payload.Amount)
	}
	if payload.Currency != "ETB" {
		t.Fatalf("Currency = %q, want ETB", payload.Currency)
	}
	if payload.Signature != "abc123" {
		t.Fatalf("Signature = %q, want abc123", payload.Signature)
	}
	if payload.PaymentID != "gateway-payment-id" {
		t.Fatalf("PaymentID = %q, want gateway-payment-id", payload.PaymentID)
	}
}

func TestWebhookPayloadUnmarshalTopLevelAliases(t *testing.T) {
	var payload WebhookPayload
	body := []byte(`{
		"transactionReference": "local-payment-id",
		"state": "captured",
		"totalAmount": 42.9,
		"currency": "ETB",
		"paymentId": "provider-payment-id"
	}`)
	if err := json.Unmarshal(body, &payload); err != nil {
		t.Fatalf("Unmarshal() error = %v", err)
	}

	if payload.Reference != "local-payment-id" {
		t.Fatalf("Reference = %q, want local-payment-id", payload.Reference)
	}
	if payload.Status != "captured" {
		t.Fatalf("Status = %q, want captured", payload.Status)
	}
	if payload.Amount != 42 {
		t.Fatalf("Amount = %d, want 42", payload.Amount)
	}
	if payload.PaymentID != "provider-payment-id" {
		t.Fatalf("PaymentID = %q, want provider-payment-id", payload.PaymentID)
	}
}

func TestVerifyWebhookSignature(t *testing.T) {
	payload := WebhookPayload{
		Reference: "payment-123",
		Status:    "paid",
		Amount:    150,
	}
	payload.Signature = signWebhookPayload(payload, "shared-secret")

	if !(&Client{}).VerifyWebhookSignature(payload, "shared-secret") {
		t.Fatal("VerifyWebhookSignature() = false, want true")
	}
}

func TestVerifyWebhookSignatureRejectsInvalidInputs(t *testing.T) {
	payload := WebhookPayload{
		Reference: "payment-123",
		Status:    "paid",
		Amount:    150,
	}
	payload.Signature = signWebhookPayload(payload, "shared-secret")

	tests := []struct {
		name    string
		payload WebhookPayload
		secret  string
	}{
		{name: "empty secret", payload: payload, secret: ""},
		{name: "wrong secret", payload: payload, secret: "wrong-secret"},
		{name: "tampered amount", payload: WebhookPayload{
			Reference: payload.Reference,
			Status:    payload.Status,
			Amount:    payload.Amount + 1,
			Signature: payload.Signature,
		}, secret: "shared-secret"},
		{name: "non hex signature", payload: WebhookPayload{
			Reference: payload.Reference,
			Status:    payload.Status,
			Amount:    payload.Amount,
			Signature: "not-hex",
		}, secret: "shared-secret"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if (&Client{}).VerifyWebhookSignature(tt.payload, tt.secret) {
				t.Fatal("VerifyWebhookSignature() = true, want false")
			}
		})
	}
}

func signWebhookPayload(payload WebhookPayload, secret string) string {
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(fmt.Sprintf("%s|%s|%d", payload.Reference, payload.Status, payload.Amount)))
	return hex.EncodeToString(mac.Sum(nil))
}
