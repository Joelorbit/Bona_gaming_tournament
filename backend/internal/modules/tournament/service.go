package tournament

import (
	"context"
	"fmt"
	"log"

	"bona-backend/internal/modules/admin"
	"bona-backend/internal/modules/notification"
	"bona-backend/internal/modules/payout"
	"bona-backend/internal/repository"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

const (
	PlatformFeePct     int32 = 5
	MaxOrganizerFeePct int32 = 20
)

type Service struct {
	repo     *repository.Queries
	db       *pgxpool.Pool
	payout   *payout.Service
	auditor  *admin.Service
	notifier *notification.Service
}

func NewService(db *pgxpool.Pool, payouts *payout.Service, auditor *admin.Service, notifier *notification.Service) *Service {
	return &Service{
		repo:     repository.New(db),
		db:       db,
		payout:   payouts,
		auditor:  auditor,
		notifier: notifier,
	}
}

func (s *Service) Create(ctx context.Context, params repository.CreateTournamentParams) (*repository.CreateTournamentRow, error) {
	if params.Title == "" {
		return nil, fmt.Errorf("title is required")
	}
	if params.Game == "" {
		return nil, fmt.Errorf("game is required")
	}
	if params.MaxParticipants <= 1 {
		return nil, fmt.Errorf("max participants must be at least 2")
	}
	if params.OrganizerFeePct < 0 || params.OrganizerFeePct > MaxOrganizerFeePct {
		return nil, fmt.Errorf("organizer fee must be between 0%% and %d%%", MaxOrganizerFeePct)
	}
	if params.BestOf > 0 && params.BestOf%2 == 0 {
		return nil, fmt.Errorf("best_of must be an odd number")
	}

	params.PlatformFeePct = PlatformFeePct

	tournament, err := s.repo.CreateTournament(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("create tournament: %w", err)
	}
	return &tournament, nil
}

func (s *Service) GetByID(ctx context.Context, id string) (*repository.GetTournamentRow, error) {
	tournament, err := s.repo.GetTournament(ctx, id)
	if err != nil {
		return nil, err
	}
	return &tournament, nil
}

type ListFilters struct {
	Status string
	Game   string
	Query  string
	Paid   string
	Sort   string
	Limit  int32
	Offset int32
}

func (s *Service) List(ctx context.Context, f ListFilters) ([]repository.ListTournamentsRow, error) {
	tournaments, err := s.repo.ListTournaments(ctx, repository.ListTournamentsParams{
		Status: f.Status,
		Game:   f.Game,
		Query:  f.Query,
		Paid:   f.Paid,
		Sort:   f.Sort,
		Limit:  f.Limit,
		Offset: f.Offset,
	})
	if err != nil {
		return nil, err
	}
	return tournaments, nil
}

func (s *Service) Update(ctx context.Context, params repository.UpdateTournamentParams, userID string) (*repository.UpdateTournamentRow, error) {
	existing, err := s.repo.GetTournament(ctx, params.ID)
	if err != nil {
		return nil, fmt.Errorf("tournament not found")
	}
	if existing.OrganizerID != userID {
		return nil, fmt.Errorf("not authorized to update this tournament")
	}
	if existing.Status == "in_progress" || existing.Status == "completed" || existing.Status == "cancelled" {
		return nil, fmt.Errorf("tournament cannot be edited once it is %s", existing.Status)
	}
	if params.EntryFee != nil && existing.Status != "draft" && *params.EntryFee != existing.EntryFee {
		return nil, fmt.Errorf("entry fee cannot change after draft")
	}
	if params.OrganizerFeePct != nil && (*params.OrganizerFeePct < 0 || *params.OrganizerFeePct > MaxOrganizerFeePct) {
		return nil, fmt.Errorf("organizer fee must be between 0%% and %d%%", MaxOrganizerFeePct)
	}
	if params.BestOf != nil && *params.BestOf > 0 && *params.BestOf%2 == 0 {
		return nil, fmt.Errorf("best_of must be an odd number")
	}

	tournament, err := s.repo.UpdateTournament(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("update tournament: %w", err)
	}
	return &tournament, nil
}

func (s *Service) Delete(ctx context.Context, id, userID string) error {
	existing, err := s.repo.GetTournament(ctx, id)
	if err != nil {
		return fmt.Errorf("tournament not found")
	}
	if existing.OrganizerID != userID {
		return fmt.Errorf("not authorized to delete this tournament")
	}

	return s.repo.DeleteTournament(ctx, id)
}

func (s *Service) UpdateStatus(ctx context.Context, id, status, userID string) (*repository.UpdateTournamentStatusRow, error) {
	existing, err := s.repo.GetTournament(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("tournament not found")
	}
	if existing.OrganizerID != userID {
		return nil, fmt.Errorf("not authorized to update this tournament")
	}

	validTransitions := map[string][]string{
		"draft":               {"open"},
		"open":                {"registration_closed", "cancelled"},
		"registration_closed": {"in_progress", "open", "cancelled"},
		"in_progress":         {"completed", "cancelled"},
		"completed":           {},
		"cancelled":           {},
	}

	allowed, ok := validTransitions[existing.Status]
	if !ok {
		return nil, fmt.Errorf("invalid current status: %s", existing.Status)
	}

	valid := false
	for _, s := range allowed {
		if s == status {
			valid = true
			break
		}
	}
	if !valid {
		return nil, fmt.Errorf("cannot transition from %s to %s", existing.Status, status)
	}

	if status == "in_progress" {
		count, err := s.repo.CountPaidRegistrations(ctx, id)
		if err != nil {
			return nil, fmt.Errorf("check participants: %w", err)
		}
		if count < int64(existing.MinParticipants) {
			return nil, fmt.Errorf("need at least %d paid participants to start (have %d)", existing.MinParticipants, count)
		}
	}

	tournament, refundPayments, err := s.updateStatusWithSideEffects(ctx, existing, status)
	if err != nil {
		return nil, fmt.Errorf("update status: %w", err)
	}

	if s.auditor != nil {
		tid := id
		s.auditor.Log(ctx, userID, "organizer", "tournament.status_changed", "tournament", &tid, map[string]any{
			"from": existing.Status, "to": status,
		})
	}

	if status == "cancelled" {
		s.notifyRefundPending(ctx, tournament, refundPayments)
	}

	if status == "completed" && s.payout != nil {
		if _, err := s.payout.EnsureForTournament(ctx, id); err != nil {
			log.Printf("ensure payout for %s: %v", id, err)
		}
	}

	return &tournament, nil
}

func (s *Service) updateStatusWithSideEffects(ctx context.Context, existing repository.Tournament, status string) (repository.Tournament, []repository.Payment, error) {
	if status != "cancelled" {
		tournament, err := s.repo.UpdateTournamentStatus(ctx, repository.UpdateTournamentStatusParams{
			ID:     existing.ID,
			Status: status,
		})
		return tournament, nil, err
	}

	tx, err := s.db.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return repository.Tournament{}, nil, err
	}
	defer tx.Rollback(ctx)

	q := repository.New(tx)
	tournament, err := q.UpdateTournamentStatus(ctx, repository.UpdateTournamentStatusParams{
		ID:     existing.ID,
		Status: status,
	})
	if err != nil {
		return repository.Tournament{}, nil, err
	}

	reason := "Tournament cancelled by organizer"
	refundPayments, err := q.MarkTournamentPaymentsRefundPending(ctx, repository.MarkTournamentPaymentsRefundPendingParams{
		TournamentID: existing.ID,
		Reason:       reason,
	})
	if err != nil {
		return repository.Tournament{}, nil, err
	}

	if err := tx.Commit(ctx); err != nil {
		return repository.Tournament{}, nil, err
	}

	return tournament, refundPayments, nil
}

func (s *Service) notifyRefundPending(ctx context.Context, tournament repository.Tournament, payments []repository.Payment) {
	if s.notifier == nil {
		return
	}
	link := fmt.Sprintf("/tournaments/%s", tournament.ID)
	for _, payment := range payments {
		s.notifier.Emit(ctx, repository.CreateNotificationParams{
			UserID:  payment.UserID,
			Type:    "refund_pending",
			Title:   "Refund pending",
			Message: fmt.Sprintf("%s was cancelled. Your %d %s entry payment is pending organizer refund.", tournament.Title, payment.Amount, payment.Currency),
			Link:    &link,
		})
	}
}

func (s *Service) ListByOrganizer(ctx context.Context, organizerID string) ([]repository.ListTournamentsByOrganizerRow, error) {
	tournaments, err := s.repo.ListTournamentsByOrganizer(ctx, organizerID)
	if err != nil {
		return nil, err
	}
	return tournaments, nil
}
