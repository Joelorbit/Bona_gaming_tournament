package middleware

import (
	"context"
	"log"
	"net/http"
	"strings"
	"time"

	"bona-backend/internal/utils"
	"bona-backend/pkg/supabase"
)

type contextKey string

const UserIDKey contextKey = "user_id"
const UserRoleKey contextKey = "user_role"
const UserEmailKey contextKey = "user_email"

func AuthMiddleware(supabaseClient *supabase.Client) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			token := extractBearerToken(r)
			if token == "" {
				utils.ErrorJSON(w, http.StatusUnauthorized, "Missing authorization token")
				return
			}

			user, err := supabaseClient.VerifyJWT(r.Context(), token)
			if err != nil {
				utils.ErrorJSON(w, http.StatusUnauthorized, "Invalid or expired token")
				return
			}

			ctx := context.WithValue(r.Context(), UserIDKey, user.ID)
			ctx = context.WithValue(ctx, UserRoleKey, user.Role)
			ctx = context.WithValue(ctx, UserEmailKey, user.Email)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func extractBearerToken(r *http.Request) string {
	auth := r.Header.Get("Authorization")
	if strings.HasPrefix(auth, "Bearer ") {
		return strings.TrimPrefix(auth, "Bearer ")
	}
	return ""
}

func GetUserID(ctx context.Context) string {
	if id, ok := ctx.Value(UserIDKey).(string); ok {
		return id
	}
	return ""
}

func GetUserEmail(ctx context.Context) string {
	if e, ok := ctx.Value(UserEmailKey).(string); ok {
		return e
	}
	return ""
}

func GetUserRole(ctx context.Context) string {
	if r, ok := ctx.Value(UserRoleKey).(string); ok {
		return r
	}
	return ""
}

func LoggerMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		wrapped := &responseWriter{ResponseWriter: w, status: http.StatusOK}
		next.ServeHTTP(wrapped, r)
		log.Printf("%s %s %d %s", r.Method, r.URL.Path, wrapped.status, time.Since(start))
	})
}

type responseWriter struct {
	http.ResponseWriter
	status int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.status = code
	rw.ResponseWriter.WriteHeader(code)
}
