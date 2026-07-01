package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os/signal"
	"syscall"
	"time"

	"bona-backend/internal/config"
	"bona-backend/internal/db"
	"bona-backend/internal/router"
	"bona-backend/pkg/addispay"
	"bona-backend/pkg/supabase"
)

func main() {
	cfg := config.Load()

	if cfg.DatabaseURL == "" {
		log.Fatal("DATABASE_URL is required")
	}
	if cfg.SupabaseURL == "" || cfg.SupabaseAnonKey == "" {
		log.Fatal("SUPABASE_URL and SUPABASE_ANON_KEY are required")
	}

	pool := db.NewPostgres(cfg.DatabaseURL)
	defer pool.Close()

	supaClient := supabase.NewClient(cfg.SupabaseURL, cfg.SupabaseAnonKey)

	var addispayClient *addispay.Client
	if cfg.AddisPayAPIKey != "" && cfg.AddisPayAPIKey != "your_addispay_api_key" {
		addispayClient = addispay.NewClient(cfg.AddisPayAPIKey, cfg.AddisPayBaseURL)
		log.Println("AddisPay client initialized")
	} else {
		log.Println("AddisPay client not configured - payment features disabled")
	}

	r := router.New(pool, &router.RouterConfig{
		SupabaseClient:        supaClient,
		AddisPayClient:        addispayClient,
		AddisPayWebhookSecret: cfg.AddisPayWebhookSecret,
		AddisPayRedirectURL:   cfg.AddisPayRedirectURL,
		AddisPayCancelURL:     cfg.AddisPayCancelURL,
		AddisPaySuccessURL:    cfg.AddisPaySuccessURL,
		AddisPayErrorURL:      cfg.AddisPayErrorURL,
		Environment:           cfg.Environment,
		AllowedOrigins:        cfg.AllowedOrigins,
	})

	addr := ":" + cfg.Port
	srv := &http.Server{
		Addr:              addr,
		Handler:           r.Router,
		ReadHeaderTimeout: 10 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	go func() {
		log.Printf("server starting on %s (env: %s)", addr, cfg.Environment)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatal("server failed:", err)
		}
	}()

	<-ctx.Done()
	log.Println("shutting down...")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Printf("graceful shutdown failed: %v", err)
	}
	log.Println("server stopped")
}
