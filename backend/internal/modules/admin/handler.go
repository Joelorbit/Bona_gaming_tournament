package admin

import (
	"net/http"
	"strconv"

	"bona-backend/internal/middleware"
	"bona-backend/internal/repository"
	"bona-backend/internal/utils"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Handler struct {
	service *Service
	repo    *repository.Queries
}

func NewHandler(db *pgxpool.Pool) *Handler {
	return &Handler{
		service: NewService(db),
		repo:    repository.New(db),
	}
}

func NewHandlerWithService(svc *Service, db *pgxpool.Pool) *Handler {
	return &Handler{service: svc, repo: repository.New(db)}
}

// requireAdmin checks that the caller has the admin role. Because the role
// from Supabase JWT may be 'authenticated' (Supabase's own role, not our app
// role), we look up the profile to read the application role.
func (h *Handler) requireAdmin(w http.ResponseWriter, r *http.Request) (string, bool) {
	userID := middleware.GetUserID(r.Context())
	if userID == "" {
		utils.ErrorJSON(w, http.StatusUnauthorized, "Unauthorized")
		return "", false
	}
	profile, err := h.repo.GetProfile(r.Context(), userID)
	if err != nil || profile.Role != "admin" {
		utils.ErrorJSON(w, http.StatusForbidden, "Admin role required")
		return "", false
	}
	return userID, true
}

func (h *Handler) Stats(w http.ResponseWriter, r *http.Request) {
	if _, ok := h.requireAdmin(w, r); !ok {
		return
	}
	stats, err := h.service.Stats(r.Context())
	if err != nil {
		utils.ErrorJSON(w, http.StatusInternalServerError, "Failed to load stats: "+err.Error())
		return
	}
	utils.JSON(w, http.StatusOK, stats)
}

func (h *Handler) AuditLog(w http.ResponseWriter, r *http.Request) {
	if _, ok := h.requireAdmin(w, r); !ok {
		return
	}
	limit := int32(100)
	offset := int32(0)
	if l := r.URL.Query().Get("limit"); l != "" {
		if v, err := strconv.Atoi(l); err == nil && v > 0 && v <= 500 {
			limit = int32(v)
		}
	}
	if o := r.URL.Query().Get("offset"); o != "" {
		if v, err := strconv.Atoi(o); err == nil && v >= 0 {
			offset = int32(v)
		}
	}
	entries, err := h.service.ListAudit(r.Context(), limit, offset)
	if err != nil {
		utils.ErrorJSON(w, http.StatusInternalServerError, "Failed to load audit log: "+err.Error())
		return
	}
	if entries == nil {
		entries = []repository.AuditLogEntry{}
	}
	utils.JSON(w, http.StatusOK, entries)
}

func (h *Handler) ListAllTournaments(w http.ResponseWriter, r *http.Request) {
	if _, ok := h.requireAdmin(w, r); !ok {
		return
	}
	tournaments, err := h.repo.ListTournaments(r.Context(), repository.ListTournamentsParams{Limit: 500, Offset: 0})
	if err != nil {
		utils.ErrorJSON(w, http.StatusInternalServerError, "Failed to load tournaments: "+err.Error())
		return
	}
	if tournaments == nil {
		tournaments = []repository.Tournament{}
	}
	utils.JSON(w, http.StatusOK, tournaments)
}

func (h *Handler) ListAllPayouts(w http.ResponseWriter, r *http.Request) {
	if _, ok := h.requireAdmin(w, r); !ok {
		return
	}
	limit := int32(200)
	offset := int32(0)
	if l := r.URL.Query().Get("limit"); l != "" {
		if v, err := strconv.Atoi(l); err == nil && v > 0 && v <= 500 {
			limit = int32(v)
		}
	}
	if o := r.URL.Query().Get("offset"); o != "" {
		if v, err := strconv.Atoi(o); err == nil && v >= 0 {
			offset = int32(v)
		}
	}
	payouts, err := h.repo.ListAllPayouts(r.Context(), limit, offset)
	if err != nil {
		utils.ErrorJSON(w, http.StatusInternalServerError, "Failed to load payouts: "+err.Error())
		return
	}
	if payouts == nil {
		payouts = []repository.Payout{}
	}
	for i := range payouts {
		payouts[i].PhoneNumber = nil
		payouts[i].TelebirrNumber = nil
		payouts[i].BankName = nil
		payouts[i].BankAccountName = nil
		payouts[i].BankAccountNumber = nil
	}
	utils.JSON(w, http.StatusOK, payouts)
}

func (h *Handler) ListAllPayments(w http.ResponseWriter, r *http.Request) {
	if _, ok := h.requireAdmin(w, r); !ok {
		return
	}
	limit := int32(200)
	offset := int32(0)
	if l := r.URL.Query().Get("limit"); l != "" {
		if v, err := strconv.Atoi(l); err == nil && v > 0 && v <= 500 {
			limit = int32(v)
		}
	}
	if o := r.URL.Query().Get("offset"); o != "" {
		if v, err := strconv.Atoi(o); err == nil && v >= 0 {
			offset = int32(v)
		}
	}
	payments, err := h.repo.ListAllPayments(r.Context(), limit, offset)
	if err != nil {
		utils.ErrorJSON(w, http.StatusInternalServerError, "Failed to load payments: "+err.Error())
		return
	}
	if payments == nil {
		payments = []repository.Payment{}
	}
	utils.JSON(w, http.StatusOK, payments)
}
