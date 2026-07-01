package registration

import (
	"context"
	"fmt"

	"bona-backend/internal/modules/notification"
	"bona-backend/internal/repository"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Service struct {
	repo     *repository.Queries
	db       *pgxpool.Pool
	notifier *notification.Service
}

func NewService(db *pgxpool.Pool, notifier *notification.Service) *Service {
	return &Service{
		repo:     repository.New(db),
		db:       db,
		notifier: notifier,
	}
}

func (s *Service) Join(ctx context.Context, userID, tournamentID string) (*repository.CreateRegistrationRow, error) {
	tournament, err := s.repo.GetTournament(ctx, tournamentID)
	if err != nil {
		return nil, fmt.Errorf("tournament not found")
	}

	if tournament.Status != "open" {
		return nil, fmt.Errorf("registration is not open for this tournament")
	}

	_, lookupErr := s.repo.GetRegistrationByUserAndTournament(ctx, repository.GetRegistrationByUserAndTournamentParams{
		UserID:       userID,
		TournamentID: tournamentID,
	})
	if lookupErr == nil {
		return nil, fmt.Errorf("already registered for this tournament")
	}

	count, err := s.repo.CountActiveRegistrations(ctx, tournamentID)
	if err != nil {
		return nil, fmt.Errorf("check capacity: %w", err)
	}

	if count >= int64(tournament.MaxParticipants) {
		return nil, fmt.Errorf("tournament is full")
	}

	paymentStatus := "pending"
	if tournament.EntryFee == 0 {
		paymentStatus = "paid"
	}

	registration, err := s.repo.CreateRegistration(ctx, repository.CreateRegistrationParams{
		UserID:        userID,
		TournamentID:  tournamentID,
		PaymentStatus: paymentStatus,
	})
	if err != nil {
		return nil, fmt.Errorf("create registration: %w", err)
	}

	link := fmt.Sprintf("/tournaments/%s", tournamentID)
	if s.notifier != nil {
		s.notifier.Emit(ctx, repository.CreateNotificationParams{
			UserID:  userID,
			Type:    "registration_confirmed",
			Title:   "Registered for " + tournament.Title,
			Message: fmt.Sprintf("You're registered for %s. %s", tournament.Title, registrationStatusMessage(paymentStatus, tournament.EntryFee)),
			Link:    &link,
		})
		s.notifier.Emit(ctx, repository.CreateNotificationParams{
			UserID:  tournament.OrganizerID,
			Type:    "new_registration",
			Title:   "New player joined " + tournament.Title,
			Message: "A new player registered for your tournament.",
			Link:    &link,
		})
	}

	return &registration, nil
}

func registrationStatusMessage(status string, fee int32) string {
	if status == "paid" {
		return "Your spot is confirmed."
	}
	if fee > 0 {
		return "Complete payment to secure your spot."
	}
	return "Your spot is confirmed."
}

func (s *Service) Leave(ctx context.Context, userID, tournamentID string) error {
	return s.repo.DeleteRegistration(ctx, repository.DeleteRegistrationParams{
		UserID:       userID,
		TournamentID: tournamentID,
	})
}

func (s *Service) ListByTournament(ctx context.Context, tournamentID string) ([]repository.ListRegistrationsByTournamentRow, error) {
	return s.repo.ListRegistrationsByTournament(ctx, tournamentID)
}

func (s *Service) ListByUser(ctx context.Context, userID string) ([]repository.ListRegistrationsByUserRow, error) {
	return s.repo.ListRegistrationsByUser(ctx, userID)
}
