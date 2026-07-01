package payout

import (
	"encoding/json"
	"net/http"

	"bona-backend/internal/middleware"
	"bona-backend/internal/utils"

	"github.com/go-chi/chi/v5"
)

type Handler struct {
	service *Service
}

func NewHandler(svc *Service) *Handler { return &Handler{service: svc} }

func (h *Handler) Service() *Service { return h.service }

func (h *Handler) ListByTournament(w http.ResponseWriter, r *http.Request) {
	tournamentID := chi.URLParam(r, "id")
	userID := middleware.GetUserID(r.Context())
	items, err := h.service.ListByTournament(r.Context(), tournamentID, userID)
	if err != nil {
		utils.ErrorJSON(w, http.StatusBadRequest, err.Error())
		return
	}
	utils.JSON(w, http.StatusOK, items)
}

func (h *Handler) ListMine(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())
	items, err := h.service.ListByWinner(r.Context(), userID)
	if err != nil {
		utils.ErrorJSON(w, http.StatusInternalServerError, "Failed to list payouts: "+err.Error())
		return
	}
	utils.JSON(w, http.StatusOK, items)
}

func (h *Handler) ListByOrganizer(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())
	items, err := h.service.ListByOrganizer(r.Context(), userID)
	if err != nil {
		utils.ErrorJSON(w, http.StatusInternalServerError, "Failed to list payouts: "+err.Error())
		return
	}
	utils.JSON(w, http.StatusOK, items)
}

func (h *Handler) MarkPaid(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	userID := middleware.GetUserID(r.Context())

	var req struct {
		Note string `json:"note"`
	}
	if r.ContentLength > 0 {
		json.NewDecoder(r.Body).Decode(&req)
	}

	p, err := h.service.MarkPaid(r.Context(), id, userID, req.Note)
	if err != nil {
		utils.ErrorJSON(w, http.StatusBadRequest, err.Error())
		return
	}
	utils.JSON(w, http.StatusOK, p)
}

func (h *Handler) SubmitDetails(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	userID := middleware.GetUserID(r.Context())

	var req SubmitDetailsParams
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.ErrorJSON(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	p, err := h.service.SubmitDetails(r.Context(), id, userID, req)
	if err != nil {
		utils.ErrorJSON(w, http.StatusBadRequest, err.Error())
		return
	}
	utils.JSON(w, http.StatusOK, p)
}
