package match

import (
	"encoding/json"
	"net/http"

	"bona-backend/internal/middleware"
	"bona-backend/internal/modules/admin"
	"bona-backend/internal/modules/notification"
	"bona-backend/internal/modules/payout"
	"bona-backend/internal/utils"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Handler struct {
	service *Service
}

func NewHandler(db *pgxpool.Pool, notifier *notification.Service, payouts *payout.Service, auditor *admin.Service) *Handler {
	return &Handler{service: NewService(db, notifier, payouts, auditor)}
}

func (h *Handler) SubmitResult(w http.ResponseWriter, r *http.Request) {
	matchID := chi.URLParam(r, "id")
	userID := middleware.GetUserID(r.Context())

	var req struct {
		WinnerID              string  `json:"winner_id"`
		Score                 string  `json:"score"`
		EvidenceScreenshotURL *string `json:"evidence_screenshot_url"`
		EvidenceVideoURL      *string `json:"evidence_video_url"`
		EvidenceNotes         *string `json:"evidence_notes"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.ErrorJSON(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	m, err := h.service.SubmitResult(r.Context(), SubmitInput{
		MatchID:               matchID,
		UserID:                userID,
		WinnerID:              req.WinnerID,
		Score:                 req.Score,
		EvidenceScreenshotURL: req.EvidenceScreenshotURL,
		EvidenceVideoURL:      req.EvidenceVideoURL,
		EvidenceNotes:         req.EvidenceNotes,
	})
	if err != nil {
		utils.ErrorJSON(w, http.StatusBadRequest, err.Error())
		return
	}
	utils.JSON(w, http.StatusOK, m)
}

func (h *Handler) ConfirmResult(w http.ResponseWriter, r *http.Request) {
	matchID := chi.URLParam(r, "id")
	userID := middleware.GetUserID(r.Context())

	m, err := h.service.ConfirmResult(r.Context(), matchID, userID)
	if err != nil {
		utils.ErrorJSON(w, http.StatusBadRequest, err.Error())
		return
	}
	utils.JSON(w, http.StatusOK, m)
}

func (h *Handler) Dispute(w http.ResponseWriter, r *http.Request) {
	matchID := chi.URLParam(r, "id")
	userID := middleware.GetUserID(r.Context())

	var req struct {
		Reason string `json:"reason"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.ErrorJSON(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	m, err := h.service.Dispute(r.Context(), matchID, userID, req.Reason)
	if err != nil {
		utils.ErrorJSON(w, http.StatusBadRequest, err.Error())
		return
	}
	utils.JSON(w, http.StatusOK, m)
}

func (h *Handler) Resolve(w http.ResponseWriter, r *http.Request) {
	matchID := chi.URLParam(r, "id")
	userID := middleware.GetUserID(r.Context())

	var req struct {
		WinnerID string `json:"winner_id"`
		Score    string `json:"score"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.ErrorJSON(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	m, err := h.service.ResolveDispute(r.Context(), matchID, userID, req.WinnerID, req.Score)
	if err != nil {
		utils.ErrorJSON(w, http.StatusBadRequest, err.Error())
		return
	}
	utils.JSON(w, http.StatusOK, m)
}

func (h *Handler) GetByID(w http.ResponseWriter, r *http.Request) {
	matchID := chi.URLParam(r, "id")
	m, err := h.service.GetByID(r.Context(), matchID)
	if err != nil {
		utils.ErrorJSON(w, http.StatusNotFound, "Match not found")
		return
	}
	utils.JSON(w, http.StatusOK, m)
}

func (h *Handler) MyMatches(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())
	matches, err := h.service.ListMyMatches(r.Context(), userID)
	if err != nil {
		utils.ErrorJSON(w, http.StatusInternalServerError, "Failed to list matches: "+err.Error())
		return
	}
	utils.JSON(w, http.StatusOK, matches)
}

func (h *Handler) MyDisputes(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())
	matches, err := h.service.ListDisputesByOrganizer(r.Context(), userID)
	if err != nil {
		utils.ErrorJSON(w, http.StatusInternalServerError, "Failed to list disputes: "+err.Error())
		return
	}
	utils.JSON(w, http.StatusOK, matches)
}
