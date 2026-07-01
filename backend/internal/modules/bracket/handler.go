package bracket

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

func (h *Handler) Generate(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())
	tournamentID := chi.URLParam(r, "id")

	matches, err := h.service.Generate(r.Context(), tournamentID, userID)
	if err != nil {
		utils.ErrorJSON(w, http.StatusBadRequest, err.Error())
		return
	}

	utils.JSON(w, http.StatusCreated, matches)
}

func (h *Handler) GetBracket(w http.ResponseWriter, r *http.Request) {
	tournamentID := chi.URLParam(r, "id")

	bracket, err := h.service.GetBracket(r.Context(), tournamentID)
	if err != nil {
		utils.ErrorJSON(w, http.StatusNotFound, "Bracket not found")
		return
	}

	utils.JSON(w, http.StatusOK, bracket)
}
