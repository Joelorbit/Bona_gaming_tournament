package user

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"bona-backend/internal/middleware"
	"bona-backend/internal/repository"
	"bona-backend/internal/utils"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Handler struct {
	service *Service
}

func NewHandler(db *pgxpool.Pool) *Handler {
	return &Handler{
		service: NewService(db),
	}
}

func (h *Handler) GetProfile(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())
	if userID == "" {
		utils.ErrorJSON(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	profile, err := h.service.GetProfileWithStats(r.Context(), userID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			utils.ErrorJSON(w, http.StatusNotFound, "Profile not found")
			return
		}
		utils.ErrorJSON(w, http.StatusInternalServerError, "Failed to load profile: "+err.Error())
		return
	}

	utils.JSON(w, http.StatusOK, profile)
}

// EnsureProfile is the bootstrap endpoint (POST /users/me). It is idempotent:
// if the caller already has a profile, the existing one is returned; otherwise
// a profile is created with the supplied fields (username/display_name/etc.)
// or sensible defaults.
func (h *Handler) EnsureProfile(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())
	if userID == "" {
		utils.ErrorJSON(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	var req repository.CreateProfileParams
	if r.ContentLength > 0 {
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			utils.ErrorJSON(w, http.StatusBadRequest, "Invalid request body")
			return
		}
	}

	if req.Email == nil {
		email := middleware.GetUserEmail(r.Context())
		if email != "" {
			req.Email = &email
		}
	}

	profile, err := h.service.EnsureProfile(r.Context(), userID, req)
	if err != nil {
		utils.ErrorJSON(w, http.StatusInternalServerError, "Failed to ensure profile: "+err.Error())
		return
	}

	utils.JSON(w, http.StatusOK, profile)
}

func (h *Handler) UpdateProfile(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())
	if userID == "" {
		utils.ErrorJSON(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	var req repository.UpdateProfileParams
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.ErrorJSON(w, http.StatusBadRequest, "Invalid request body")
		return
	}
	req.ID = userID

	profile, err := h.service.UpdateProfile(r.Context(), req)
	if err != nil {
		utils.ErrorJSON(w, http.StatusInternalServerError, "Failed to update profile: "+err.Error())
		return
	}

	utils.JSON(w, http.StatusOK, profile)
}

func (h *Handler) GetByUsername(w http.ResponseWriter, r *http.Request) {
	username := chi.URLParam(r, "username")
	if username == "" {
		utils.ErrorJSON(w, http.StatusBadRequest, "Username required")
		return
	}
	p, err := h.service.GetByUsername(r.Context(), username)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			utils.ErrorJSON(w, http.StatusNotFound, "User not found")
			return
		}
		utils.ErrorJSON(w, http.StatusInternalServerError, "Failed to load profile: "+err.Error())
		return
	}
	utils.JSON(w, http.StatusOK, p)
}

func (h *Handler) GetByID(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if id == "" {
		utils.ErrorJSON(w, http.StatusBadRequest, "User ID required")
		return
	}
	p, err := h.service.GetProfileWithStats(r.Context(), id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			utils.ErrorJSON(w, http.StatusNotFound, "User not found")
			return
		}
		utils.ErrorJSON(w, http.StatusInternalServerError, "Failed to load profile: "+err.Error())
		return
	}
	utils.JSON(w, http.StatusOK, p)
}

func (h *Handler) Search(w http.ResponseWriter, r *http.Request) {
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))

	profiles, err := h.service.Search(r.Context(), r.URL.Query().Get("q"), int32(limit), int32(offset))
	if err != nil {
		utils.ErrorJSON(w, http.StatusInternalServerError, "Failed to search users: "+err.Error())
		return
	}
	if profiles == nil {
		profiles = []repository.ProfileSearchResult{}
	}

	utils.JSON(w, http.StatusOK, profiles)
}
