package tournament

import (
	"encoding/json"
	"net/http"
	"strconv"

	"bona-backend/internal/middleware"
	"bona-backend/internal/modules/admin"
	"bona-backend/internal/modules/notification"
	"bona-backend/internal/modules/payout"
	"bona-backend/internal/repository"
	"bona-backend/internal/utils"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Handler struct {
	service *Service
}

func NewHandler(db *pgxpool.Pool, payouts *payout.Service, auditor *admin.Service, notifier *notification.Service) *Handler {
	return &Handler{
		service: NewService(db, payouts, auditor, notifier),
	}
}

func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())
	if userID == "" {
		utils.ErrorJSON(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	var req repository.CreateTournamentParams
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.ErrorJSON(w, http.StatusBadRequest, "Invalid request body")
		return
	}
	req.OrganizerID = userID

	tournament, err := h.service.Create(r.Context(), req)
	if err != nil {
		utils.ErrorJSON(w, http.StatusBadRequest, err.Error())
		return
	}

	utils.JSON(w, http.StatusCreated, tournament)
}

func (h *Handler) GetByID(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if id == "" {
		utils.ErrorJSON(w, http.StatusBadRequest, "Missing tournament ID")
		return
	}

	tournament, err := h.service.GetByID(r.Context(), id)
	if err != nil {
		utils.ErrorJSON(w, http.StatusNotFound, "Tournament not found")
		return
	}

	utils.JSON(w, http.StatusOK, tournament)
}

func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))
	if limit <= 0 || limit > 100 {
		limit = 20
	}

	tournaments, err := h.service.List(r.Context(), ListFilters{
		Status: r.URL.Query().Get("status"),
		Game:   r.URL.Query().Get("game"),
		Query:  r.URL.Query().Get("q"),
		Paid:   r.URL.Query().Get("paid"),
		Sort:   r.URL.Query().Get("sort"),
		Limit:  int32(limit),
		Offset: int32(offset),
	})
	if err != nil {
		utils.ErrorJSON(w, http.StatusInternalServerError, "Failed to fetch tournaments: "+err.Error())
		return
	}
	if tournaments == nil {
		tournaments = []repository.Tournament{}
	}

	utils.JSON(w, http.StatusOK, tournaments)
}

func (h *Handler) Update(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())
	id := chi.URLParam(r, "id")

	var req repository.UpdateTournamentParams
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.ErrorJSON(w, http.StatusBadRequest, "Invalid request body")
		return
	}
	req.ID = id

	tournament, err := h.service.Update(r.Context(), req, userID)
	if err != nil {
		utils.ErrorJSON(w, http.StatusForbidden, err.Error())
		return
	}

	utils.JSON(w, http.StatusOK, tournament)
}

func (h *Handler) Delete(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())
	id := chi.URLParam(r, "id")

	if err := h.service.Delete(r.Context(), id, userID); err != nil {
		utils.ErrorJSON(w, http.StatusForbidden, err.Error())
		return
	}

	utils.JSON(w, http.StatusOK, map[string]string{"message": "Tournament deleted"})
}

func (h *Handler) UpdateStatus(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())
	id := chi.URLParam(r, "id")

	var req struct {
		Status string `json:"status"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.ErrorJSON(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	tournament, err := h.service.UpdateStatus(r.Context(), id, req.Status, userID)
	if err != nil {
		utils.ErrorJSON(w, http.StatusBadRequest, err.Error())
		return
	}

	utils.JSON(w, http.StatusOK, tournament)
}

func (h *Handler) ListByOrganizer(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())

	tournaments, err := h.service.ListByOrganizer(r.Context(), userID)
	if err != nil {
		utils.ErrorJSON(w, http.StatusInternalServerError, "Failed to fetch tournaments")
		return
	}

	utils.JSON(w, http.StatusOK, tournaments)
}
