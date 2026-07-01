package db

import (
	"context"
	"log"

	"github.com/jackc/pgx/v5/pgxpool"
)

func NewPostgres(url string) *pgxpool.Pool {
	config, err := pgxpool.ParseConfig(url)
	if err != nil {
		log.Fatal("invalid DB config:", err)
	}

	pool, err := pgxpool.NewWithConfig(context.Background(), config)
	if err != nil {
		log.Fatal("failed to connect DB:", err)
	}

	if err := pool.Ping(context.Background()); err != nil {
		log.Fatal("DB ping failed:", err)
	}

	log.Println("connected to PostgreSQL")
	return pool
}
