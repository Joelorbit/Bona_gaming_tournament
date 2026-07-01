package registration

import (
	"net/http"

	"bona-backend/internal/middleware"
	"bona-backend/internal/modules/notification"
	"bona-backend/internal/utils"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Handler struct {
	service *Service
}

func NewHandler(db *pgxpool.Pool, notifier *notification.Service) *Handler {
	return &Handler{
		service: NewService(db, notifier),
	}
}

func (h *Handler) Join(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())
	tournamentID := chi.URLParam(r, "id")

	registration, err := h.service.Join(r.Context(), userID, tournamentID)
	if err != nil {
		utils.ErrorJSON(w, http.StatusBadRequest, err.Error())
		return
	}

	utils.JSON(w, http.StatusCreated, registration)
}

func (h *Handler) Leave(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())
	tournamentID := chi.URLParam(r, "id")

	if err := h.service.Leave(r.Context(), userID, tournamentID); err != nil {
		utils.ErrorJSON(w, http.StatusBadRequest, err.Error())
		return
	}

	utils.JSON(w, http.StatusOK, map[string]string{"message": "Left tournament"})
}

func (h *Handler) ListByTournament(w http.ResponseWriter, r *http.Request) {
	tournamentID := chi.URLParam(r, "id")

	registrations, err := h.service.ListByTournament(r.Context(), tournamentID)
	if err != nil {
		utils.ErrorJSON(w, http.StatusInternalServerError, "Failed to fetch registrations")
		return
	}

	utils.JSON(w, http.StatusOK, registrations)
}

func (h *Handler) ListByUser(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())

	registrations, err := h.service.ListByUser(r.Context(), userID)
	if err != nil {
		utils.ErrorJSON(w, http.StatusInternalServerError, "Failed to fetch registrations")
		return
	}

	utils.JSON(w, http.StatusOK, registrations)
}
