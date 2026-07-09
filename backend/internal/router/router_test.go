package router

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestFrontendProtectedPathsAreMounted(t *testing.T) {
	r := New(nil, &RouterConfig{})

	tests := []struct {
		method string
		path   string
		body   string
	}{
		{http.MethodGet, "/api/v1/users/me", ""},
		{http.MethodPost, "/api/v1/users/me", "{}"},
		{http.MethodPatch, "/api/v1/users/me", "{}"},
		{http.MethodGet, "/api/v1/users/test-user", ""},
		{http.MethodGet, "/api/v1/tournaments", ""},
		{http.MethodPost, "/api/v1/tournaments", "{}"},
		{http.MethodGet, "/api/v1/tournaments/my", ""},
		{http.MethodGet, "/api/v1/tournaments/tournament-id", ""},
		{http.MethodPatch, "/api/v1/tournaments/tournament-id/status", "{}"},
		{http.MethodPost, "/api/v1/tournaments/tournament-id/join", ""},
		{http.MethodDelete, "/api/v1/tournaments/tournament-id/leave", ""},
		{http.MethodGet, "/api/v1/tournaments/tournament-id/players", ""},
		{http.MethodPost, "/api/v1/tournaments/tournament-id/bracket/generate", ""},
		{http.MethodGet, "/api/v1/tournaments/tournament-id/bracket", ""},
		{http.MethodGet, "/api/v1/me/registrations", ""},
		{http.MethodGet, "/api/v1/me/matches", ""},
		{http.MethodGet, "/api/v1/me/disputes", ""},
		{http.MethodGet, "/api/v1/me/payouts", ""},
		{http.MethodGet, "/api/v1/me/organizer-payouts", ""},
		{http.MethodGet, "/api/v1/me/organizer-payments", ""},
		{http.MethodGet, "/api/v1/matches/match-id", ""},
		{http.MethodPost, "/api/v1/matches/match-id/result", "{}"},
		{http.MethodPost, "/api/v1/matches/match-id/confirm", ""},
		{http.MethodPost, "/api/v1/matches/match-id/dispute", "{}"},
		{http.MethodPost, "/api/v1/matches/match-id/resolve", "{}"},
		{http.MethodPost, "/api/v1/payouts/payout-id/details", "{}"},
		{http.MethodPost, "/api/v1/payouts/payout-id/mark-paid", "{}"},
		{http.MethodGet, "/api/v1/notifications", ""},
		{http.MethodGet, "/api/v1/notifications/unread-count", ""},
		{http.MethodPost, "/api/v1/notifications/read-all", ""},
		{http.MethodPost, "/api/v1/notifications/notification-id/read", ""},
		{http.MethodDelete, "/api/v1/notifications/notification-id", ""},
		{http.MethodPost, "/api/v1/payments/create", "{}"},
		{http.MethodGet, "/api/v1/payments/status/payment-id", ""},
		{http.MethodPost, "/api/v1/payments/payment-id/mark-refunded", "{}"},
		{http.MethodGet, "/api/v1/admin/stats", ""},
		{http.MethodGet, "/api/v1/admin/audit", ""},
		{http.MethodGet, "/api/v1/admin/tournaments", ""},
		{http.MethodGet, "/api/v1/admin/payments", ""},
		{http.MethodGet, "/api/v1/admin/payouts", ""},
	}

	for _, tt := range tests {
		t.Run(tt.method+" "+tt.path, func(t *testing.T) {
			req := httptest.NewRequest(tt.method, tt.path, strings.NewReader(tt.body))
			if tt.body != "" {
				req.Header.Set("Content-Type", "application/json")
			}
			rec := httptest.NewRecorder()

			r.Router.ServeHTTP(rec, req)

			if rec.Code != http.StatusUnauthorized {
				t.Fatalf("status = %d, want %d; body: %s", rec.Code, http.StatusUnauthorized, rec.Body.String())
			}
		})
	}
}

func TestPublicPathsAreMounted(t *testing.T) {
	r := New(nil, &RouterConfig{})

	tests := []struct {
		method     string
		path       string
		wantStatus int
	}{
		{http.MethodGet, "/health", http.StatusOK},
		{http.MethodGet, "/api/v1/auth/callback", http.StatusOK},
		{http.MethodPost, "/api/v1/payments/webhook", http.StatusBadRequest},
	}

	for _, tt := range tests {
		t.Run(tt.method+" "+tt.path, func(t *testing.T) {
			req := httptest.NewRequest(tt.method, tt.path, nil)
			rec := httptest.NewRecorder()

			r.Router.ServeHTTP(rec, req)

			if rec.Code != tt.wantStatus {
				t.Fatalf("status = %d, want %d; body: %s", rec.Code, tt.wantStatus, rec.Body.String())
			}
		})
	}
}

func TestWebhookRouteRejectsInvalidSignatureBeforeDatabase(t *testing.T) {
	r := New(nil, &RouterConfig{
		AddisPayWebhookSecret: "shared-secret",
	})

	req := httptest.NewRequest(http.MethodPost, "/api/v1/payments/webhook", strings.NewReader(`{
		"reference": "payment-123",
		"status": "paid",
		"amount": 100,
		"currency": "ETB",
		"signature": "00"
	}`))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	r.Router.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want %d; body: %s", rec.Code, http.StatusUnauthorized, rec.Body.String())
	}
}
