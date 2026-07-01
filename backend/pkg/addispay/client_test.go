package addispay

import (
	"encoding/json"
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
