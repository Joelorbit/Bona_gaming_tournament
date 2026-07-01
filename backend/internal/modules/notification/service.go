package notification

import (
	"context"
	"log"

	"bona-backend/internal/repository"
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

// Emit creates a notification for a user. Failures are logged but never returned —
// notification delivery should never block the caller's primary action.
func (s *Service) Emit(ctx context.Context, p repository.CreateNotificationParams) {
	if _, err := s.repo.CreateNotification(ctx, p); err != nil {
		log.Printf("notification emit failed (user=%s type=%s): %v", p.UserID, p.Type, err)
	}
}

func (s *Service) List(ctx context.Context, userID string, limit, offset int32) ([]repository.Notification, error) {
	return s.repo.ListNotificationsByUser(ctx, userID, limit, offset)
}

func (s *Service) CountUnread(ctx context.Context, userID string) (int64, error) {
	return s.repo.CountUnreadNotifications(ctx, userID)
}

func (s *Service) MarkRead(ctx context.Context, id, userID string) (*repository.Notification, error) {
	n, err := s.repo.MarkNotificationRead(ctx, id, userID)
	if err != nil {
		return nil, err
	}
	return &n, nil
}

func (s *Service) MarkAllRead(ctx context.Context, userID string) error {
	return s.repo.MarkAllNotificationsRead(ctx, userID)
}

func (s *Service) Delete(ctx context.Context, id, userID string) error {
	return s.repo.DeleteNotification(ctx, id, userID)
}
