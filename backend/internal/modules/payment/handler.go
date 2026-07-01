package payment

import (
	"encoding/json"
	"log"
	"net/http"

	"bona-backend/internal/middleware"
	"bona-backend/internal/modules/notification"
	"bona-backend/internal/utils"
	"bona-backend/pkg/addispay"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Handler struct {
	service       *Service
	webhookSecret string
}

func NewHandler(db *pgxpool.Pool, addispayClient *addispay.Client, checkout CheckoutConfig, notifier *notification.Service) *Handler {
	return &Handler{
		service:       NewService(db, addispayClient, checkout, notifier),
		webhookSecret: checkout.WebhookSecret,
	}
}

func (h *Handler) CreatePayment(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())

	var req struct {
		TournamentID string `json:"tournament_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.ErrorJSON(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	result, err := h.service.CreatePayment(r.Context(), userID, req.TournamentID)
	if err != nil {
		utils.ErrorJSON(w, http.StatusBadRequest, err.Error())
		return
	}

	utils.JSON(w, http.StatusCreated, result)
}

func (h *Handler) Webhook(w http.ResponseWriter, r *http.Request) {
	var payload addispay.WebhookPayload
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		log.Printf("AddisPay webhook decode failed: %v", err)
		utils.ErrorJSON(w, http.StatusBadRequest, "Invalid webhook payload")
		return
	}
	if payload.Signature == "" {
		payload.Signature = firstHeader(r,
			"X-AddisPay-Signature",
			"X-Addispay-Signature",
			"X-Signature",
			"Signature",
		)
	}

	if !(&addispay.Client{}).VerifyWebhookSignature(payload, h.webhookSecret) {
		log.Printf("AddisPay webhook signature failed: reference=%q status=%q amount=%d payment_id=%q", payload.Reference, payload.Status, payload.Amount, payload.PaymentID)
		utils.ErrorJSON(w, http.StatusUnauthorized, "Invalid signature")
		return
	}

	if err := h.service.HandleWebhook(r.Context(), payload); err != nil {
		log.Printf("AddisPay webhook handling failed: reference=%q status=%q amount=%d payment_id=%q error=%v", payload.Reference, payload.Status, payload.Amount, payload.PaymentID, err)
		utils.ErrorJSON(w, http.StatusBadRequest, err.Error())
		return
	}

	utils.JSON(w, http.StatusOK, map[string]string{"status": "received"})
}

func firstHeader(r *http.Request, keys ...string) string {
	for _, key := range keys {
		if v := r.Header.Get(key); v != "" {
			return v
		}
	}
	return ""
}

func (h *Handler) GetPaymentStatus(w http.ResponseWriter, r *http.Request) {
	paymentID := chi.URLParam(r, "id")

	payment, err := h.service.GetPayment(r.Context(), paymentID)
	if err != nil {
		utils.ErrorJSON(w, http.StatusNotFound, "Payment not found")
		return
	}

	utils.JSON(w, http.StatusOK, payment)
}

func (h *Handler) ConfirmReturn(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())

	var req struct {
		PaymentID string `json:"payment_id"`
		Status    string `json:"status"`
		Token     string `json:"token"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.ErrorJSON(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	result, err := h.service.ConfirmReturn(r.Context(), userID, req.PaymentID, req.Status, req.Token)
	if err != nil {
		utils.ErrorJSON(w, http.StatusBadRequest, err.Error())
		return
	}

	utils.JSON(w, http.StatusOK, result)
}
