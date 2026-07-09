package repository

import (
	"strings"
	"testing"
)

func TestCompletePaymentFromWebhookSQLIsIdempotentForPaidPayments(t *testing.T) {
	requiredFragments := []string{
		"WHEN status = 'paid' THEN status",
		"WHEN $2 = 'paid' THEN COALESCE(verified_at, NOW())",
		"webhook_received_at = NOW()",
	}

	for _, fragment := range requiredFragments {
		if !strings.Contains(completePaymentFromWebhook, fragment) {
			t.Fatalf("completePaymentFromWebhook SQL missing fragment %q", fragment)
		}
	}
}
