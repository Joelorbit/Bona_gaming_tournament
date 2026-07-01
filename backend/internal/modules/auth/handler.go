package auth

import (
	"net/http"

	"bona-backend/internal/utils"
)

type Handler struct{}

func NewHandler() *Handler {
	return &Handler{}
}

func (h *Handler) RegisterRoutes(mux *http.ServeMux) {}

func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	utils.ErrorJSON(w, http.StatusNotImplemented, "Auth is handled by Supabase client-side")
}

func (h *Handler) Callback(w http.ResponseWriter, r *http.Request) {
	utils.JSON(w, http.StatusOK, map[string]string{"message": "Auth callback received"})
}
