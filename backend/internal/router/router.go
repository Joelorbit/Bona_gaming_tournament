package router

import (
	"net/http"

	"bona-backend/internal/middleware"
	"bona-backend/internal/modules/admin"
	"bona-backend/internal/modules/auth"
	"bona-backend/internal/modules/bracket"
	"bona-backend/internal/modules/match"
	"bona-backend/internal/modules/notification"
	"bona-backend/internal/modules/payment"
	"bona-backend/internal/modules/payout"
	"bona-backend/internal/modules/registration"
	"bona-backend/internal/modules/tournament"
	"bona-backend/internal/modules/user"
	"bona-backend/pkg/addispay"
	"bona-backend/pkg/supabase"

	"github.com/go-chi/chi/v5"
	chimw "github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Router struct {
	Router *chi.Mux
	DB     *pgxpool.Pool
	Config *RouterConfig
}

type RouterConfig struct {
	SupabaseClient        *supabase.Client
	AddisPayClient        *addispay.Client
	AddisPayWebhookSecret string
	AddisPayRedirectURL   string
	AddisPayCancelURL     string
	AddisPaySuccessURL    string
	AddisPayErrorURL      string
	Environment           string
	AllowedOrigins        []string
}

func New(db *pgxpool.Pool, cfg *RouterConfig) *Router {
	r := &Router{
		Router: chi.NewRouter(),
		DB:     db,
		Config: cfg,
	}

	r.setupMiddleware(cfg)
	r.setupRoutes(cfg)

	return r
}

func (r *Router) setupMiddleware(cfg *RouterConfig) {
	r.Router.Use(chimw.RequestID)
	r.Router.Use(chimw.RealIP)
	r.Router.Use(middleware.LoggerMiddleware)
	r.Router.Use(chimw.Recoverer)
	r.Router.Use(cors.Handler(middleware.CORSMiddleware(cfg.AllowedOrigins)))
}

func (r *Router) setupRoutes(cfg *RouterConfig) {
	r.Router.Get("/health", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	notifier := notification.NewService(r.DB)
	auditor := admin.NewService(r.DB)
	payouts := payout.NewService(r.DB, notifier, auditor)
	adminHandler := admin.NewHandlerWithService(auditor, r.DB)

	userHandler := user.NewHandler(r.DB)
	tournamentHandler := tournament.NewHandler(r.DB, payouts, auditor, notifier)
	registrationHandler := registration.NewHandler(r.DB, notifier)
	bracketHandler := bracket.NewHandler(r.DB, notifier)
	matchHandler := match.NewHandler(r.DB, notifier, payouts, auditor)
	paymentHandler := payment.NewHandler(r.DB, cfg.AddisPayClient, payment.CheckoutConfig{
		WebhookSecret: cfg.AddisPayWebhookSecret,
		RedirectURL:   cfg.AddisPayRedirectURL,
		CancelURL:     cfg.AddisPayCancelURL,
		SuccessURL:    cfg.AddisPaySuccessURL,
		ErrorURL:      cfg.AddisPayErrorURL,
	}, notifier)
	notificationHandler := notification.NewHandlerWithService(notifier)
	payoutHandler := payout.NewHandler(payouts)

	authMiddleware := middleware.AuthMiddleware(cfg.SupabaseClient)

	r.Router.Route("/api/v1", func(r chi.Router) {
		r.Get("/auth/callback", (&auth.Handler{}).Callback)

		r.Group(func(r chi.Router) {
			r.Use(authMiddleware)

			r.Route("/users", func(r chi.Router) {
				r.Get("/me", userHandler.GetProfile)
				r.Post("/me", userHandler.EnsureProfile)
				r.Patch("/me", userHandler.UpdateProfile)
				r.Get("/search", userHandler.Search)
				r.Get("/id/{id}", userHandler.GetByID)
				r.Get("/{username}", userHandler.GetByUsername)
			})

			r.Route("/tournaments", func(r chi.Router) {
				r.Get("/", tournamentHandler.List)
				r.Post("/", tournamentHandler.Create)
				r.Get("/my", tournamentHandler.ListByOrganizer)
				r.Get("/{id}", tournamentHandler.GetByID)
				r.Patch("/{id}", tournamentHandler.Update)
				r.Delete("/{id}", tournamentHandler.Delete)
				r.Patch("/{id}/status", tournamentHandler.UpdateStatus)

				r.Post("/{id}/join", registrationHandler.Join)
				r.Delete("/{id}/leave", registrationHandler.Leave)
				r.Get("/{id}/players", registrationHandler.ListByTournament)

				r.Post("/{id}/bracket/generate", bracketHandler.Generate)
				r.Get("/{id}/bracket", bracketHandler.GetBracket)

				r.Get("/{id}/payouts", payoutHandler.ListByTournament)
			})

			r.Get("/me/registrations", registrationHandler.ListByUser)
			r.Get("/me/matches", matchHandler.MyMatches)
			r.Get("/me/disputes", matchHandler.MyDisputes)
			r.Get("/me/payouts", payoutHandler.ListMine)
			r.Get("/me/organizer-payouts", payoutHandler.ListByOrganizer)
			r.Get("/me/organizer-payments", paymentHandler.ListByOrganizer)

			r.Route("/matches", func(r chi.Router) {
				r.Get("/{id}", matchHandler.GetByID)
				r.Post("/{id}/result", matchHandler.SubmitResult)
				r.Post("/{id}/confirm", matchHandler.ConfirmResult)
				r.Post("/{id}/dispute", matchHandler.Dispute)
				r.Post("/{id}/resolve", matchHandler.Resolve)
			})

			r.Route("/payouts", func(r chi.Router) {
				r.Post("/{id}/details", payoutHandler.SubmitDetails)
				r.Post("/{id}/mark-paid", payoutHandler.MarkPaid)
			})

			r.Route("/notifications", func(r chi.Router) {
				r.Get("/", notificationHandler.List)
				r.Get("/unread-count", notificationHandler.UnreadCount)
				r.Post("/read-all", notificationHandler.MarkAllRead)
				r.Post("/{id}/read", notificationHandler.MarkRead)
				r.Delete("/{id}", notificationHandler.Delete)
			})

			r.Route("/payments", func(r chi.Router) {
				r.Post("/create", paymentHandler.CreatePayment)
				r.Post("/return", paymentHandler.ConfirmReturn)
				r.Get("/status/{id}", paymentHandler.GetPaymentStatus)
				r.Post("/{id}/mark-refunded", paymentHandler.MarkRefunded)
			})

			r.Route("/admin", func(r chi.Router) {
				r.Get("/stats", adminHandler.Stats)
				r.Get("/audit", adminHandler.AuditLog)
				r.Get("/tournaments", adminHandler.ListAllTournaments)
				r.Get("/payments", adminHandler.ListAllPayments)
				r.Get("/payouts", adminHandler.ListAllPayouts)
			})
		})

		r.Post("/payments/webhook", paymentHandler.Webhook)
	})
}
