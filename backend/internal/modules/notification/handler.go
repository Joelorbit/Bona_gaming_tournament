package notification

import (
	"net/http"
	"strconv"

	"bona-backend/internal/middleware"
	"bona-backend/internal/repository"
	"bona-backend/internal/utils"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Handler struct {
	service *Service
}

func NewHandler(db *pgxpool.Pool) *Handler {
	return &Handler{service: NewService(db)}
}

func NewHandlerWithService(svc *Service) *Handler {
	return &Handler{service: svc}
}

func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())
	if userID == "" {
		utils.ErrorJSON(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	limit := int32(50)
	offset := int32(0)
	if l := r.URL.Query().Get("limit"); l != "" {
		if v, err := strconv.Atoi(l); err == nil && v > 0 && v <= 100 {
			limit = int32(v)
		}
	}
	if o := r.URL.Query().Get("offset"); o != "" {
		if v, err := strconv.Atoi(o); err == nil && v >= 0 {
			offset = int32(v)
		}
	}

	items, err := h.service.List(r.Context(), userID, limit, offset)
	if err != nil {
		utils.ErrorJSON(w, http.StatusInternalServerError, "Failed to list notifications: "+err.Error())
		return
	}
	if items == nil {
		items = []repository.Notification{}
	}
	utils.JSON(w, http.StatusOK, items)
}

func (h *Handler) UnreadCount(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())
	if userID == "" {
		utils.ErrorJSON(w, http.StatusUnauthorized, "Unauthorized")
		return
	}
	count, err := h.service.CountUnread(r.Context(), userID)
	if err != nil {
		utils.ErrorJSON(w, http.StatusInternalServerError, "Failed to count notifications: "+err.Error())
		return
	}
	utils.JSON(w, http.StatusOK, map[string]int64{"count": count})
}

func (h *Handler) MarkRead(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())
	if userID == "" {
		utils.ErrorJSON(w, http.StatusUnauthorized, "Unauthorized")
		return
	}
	id := chi.URLParam(r, "id")
	n, err := h.service.MarkRead(r.Context(), id, userID)
	if err != nil {
		utils.ErrorJSON(w, http.StatusInternalServerError, "Failed to mark notification read")
		return
	}
	utils.JSON(w, http.StatusOK, n)
}

func (h *Handler) MarkAllRead(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())
	if userID == "" {
		utils.ErrorJSON(w, http.StatusUnauthorized, "Unauthorized")
		return
	}
	if err := h.service.MarkAllRead(r.Context(), userID); err != nil {
		utils.ErrorJSON(w, http.StatusInternalServerError, "Failed to mark all read")
		return
	}
	utils.JSON(w, http.StatusOK, map[string]bool{"ok": true})
}

func (h *Handler) Delete(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())
	if userID == "" {
		utils.ErrorJSON(w, http.StatusUnauthorized, "Unauthorized")
		return
	}
	id := chi.URLParam(r, "id")
	if err := h.service.Delete(r.Context(), id, userID); err != nil {
		utils.ErrorJSON(w, http.StatusInternalServerError, "Failed to delete notification")
		return
	}
	utils.JSON(w, http.StatusOK, map[string]bool{"ok": true})
}
