package user

import (
	"context"
	"errors"

	"bona-backend/internal/repository"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Service struct {
	repo *repository.Queries
	db   *pgxpool.Pool
}

func NewService(db *pgxpool.Pool) *Service {
	return &Service{
		repo: repository.New(db),
		db:   db,
	}
}

func (s *Service) GetProfile(ctx context.Context, userID string) (*repository.Profile, error) {
	profile, err := s.repo.GetProfile(ctx, userID)
	if err != nil {
		return nil, err
	}
	return &profile, nil
}

func (s *Service) GetProfileWithStats(ctx context.Context, userID string) (*repository.ProfileWithStats, error) {
	profile, err := s.repo.GetProfile(ctx, userID)
	if err != nil {
		return nil, err
	}
	stats, err := s.repo.GetProfileStats(ctx, userID)
	if err != nil {
		return nil, err
	}
	return &repository.ProfileWithStats{Profile: profile, Stats: stats}, nil
}

// EnsureProfile returns the profile for userID, creating one with defaults if
// none exists yet. This is the bootstrap call the frontend issues after a fresh
// Supabase signup so every authenticated user has a corresponding profiles row.
func (s *Service) EnsureProfile(ctx context.Context, userID string, params repository.CreateProfileParams) (*repository.Profile, error) {
	existing, err := s.repo.GetProfile(ctx, userID)
	if err == nil {
		return &existing, nil
	}
	if !errors.Is(err, pgx.ErrNoRows) {
		return nil, err
	}

	params.ID = userID
	if params.Username == "" {
		params.Username = "user_" + userID[:8]
	}
	if params.Role == "" {
		params.Role = "player"
	}

	profile, err := s.repo.CreateProfile(ctx, params)
	if err != nil {
		return nil, err
	}
	return &profile, nil
}

func (s *Service) UpdateProfile(ctx context.Context, params repository.UpdateProfileParams) (*repository.Profile, error) {
	profile, err := s.repo.UpdateProfile(ctx, params)
	if err != nil {
		return nil, err
	}
	return &profile, nil
}

func (s *Service) GetByUsername(ctx context.Context, username string) (*repository.ProfileWithStats, error) {
	profile, err := s.repo.GetProfileByUsername(ctx, username)
	if err != nil {
		return nil, err
	}
	stats, err := s.repo.GetProfileStats(ctx, profile.ID)
	if err != nil {
		return nil, err
	}
	return &repository.ProfileWithStats{Profile: profile, Stats: stats}, nil
}

func (s *Service) Search(ctx context.Context, query string, limit, offset int32) ([]repository.ProfileSearchResult, error) {
	if limit <= 0 || limit > 50 {
		limit = 20
	}
	if offset < 0 {
		offset = 0
	}
	return s.repo.SearchProfiles(ctx, repository.SearchProfilesParams{
		Query:  query,
		Limit:  limit,
		Offset: offset,
	})
}
