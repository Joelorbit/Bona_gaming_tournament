package match

import (
	"context"
	"fmt"
	"log"

	"bona-backend/internal/modules/admin"
	"bona-backend/internal/modules/notification"
	"bona-backend/internal/modules/payout"
	"bona-backend/internal/repository"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Service struct {
	repo     *repository.Queries
	db       *pgxpool.Pool
	notifier *notification.Service
	payout   *payout.Service
	auditor  *admin.Service
}

func NewService(db *pgxpool.Pool, notifier *notification.Service, payouts *payout.Service, auditor *admin.Service) *Service {
	return &Service{
		repo:     repository.New(db),
		db:       db,
		notifier: notifier,
		payout:   payouts,
		auditor:  auditor,
	}
}

type SubmitInput struct {
	MatchID               string
	UserID                string
	WinnerID              string
	Score                 string
	EvidenceScreenshotURL *string
	EvidenceVideoURL      *string
	EvidenceNotes         *string
}

// SubmitResult is called by either player (typically the winner) to record the
// outcome. The match enters 'awaiting_confirmation' until the other player
// accepts or disputes it.
func (s *Service) SubmitResult(ctx context.Context, in SubmitInput) (*repository.Match, error) {
	m, err := s.repo.GetMatch(ctx, in.MatchID)
	if err != nil {
		return nil, fmt.Errorf("match not found")
	}
	if m.Status == "completed" || m.Status == "walkover" {
		return nil, fmt.Errorf("match already completed")
	}
	if m.PlayerAID == nil || m.PlayerBID == nil {
		return nil, fmt.Errorf("match does not have both players assigned")
	}
	if in.UserID != *m.PlayerAID && in.UserID != *m.PlayerBID {
		return nil, fmt.Errorf("only the match participants can submit a result")
	}
	if in.WinnerID != *m.PlayerAID && in.WinnerID != *m.PlayerBID {
		return nil, fmt.Errorf("winner must be one of the match participants")
	}

	updated, err := s.repo.SubmitMatchResult(ctx, repository.SubmitMatchResultParams{
		ID:                    in.MatchID,
		WinnerID:              in.WinnerID,
		Score:                 in.Score,
		SubmittedBy:           in.UserID,
		EvidenceScreenshotURL: in.EvidenceScreenshotURL,
		EvidenceVideoURL:      in.EvidenceVideoURL,
		EvidenceNotes:         in.EvidenceNotes,
	})
	if err != nil {
		return nil, fmt.Errorf("submit result: %w", err)
	}

	// Notify the opponent to confirm/dispute.
	opponent := opponentID(&updated, in.UserID)
	if opponent != "" && s.notifier != nil {
		link := fmt.Sprintf("/matches/%s", updated.ID)
		s.notifier.Emit(ctx, repository.CreateNotificationParams{
			UserID:  opponent,
			Type:    "result_pending_confirmation",
			Title:   "Confirm match result",
			Message: "Your opponent submitted a result. Confirm or dispute it.",
			Link:    &link,
		})
	}

	return &updated, nil
}

// ConfirmResult is called by the player who did NOT submit the result.
// It finalises the match and auto-advances the winner.
func (s *Service) ConfirmResult(ctx context.Context, matchID, userID string) (*repository.Match, error) {
	m, err := s.repo.GetMatch(ctx, matchID)
	if err != nil {
		return nil, fmt.Errorf("match not found")
	}
	if m.Status != "awaiting_confirmation" {
		return nil, fmt.Errorf("no result is awaiting confirmation")
	}
	if m.PlayerAID == nil || m.PlayerBID == nil {
		return nil, fmt.Errorf("match missing players")
	}
	if userID != *m.PlayerAID && userID != *m.PlayerBID {
		return nil, fmt.Errorf("only participants can confirm")
	}
	if m.ResultSubmittedBy != nil && *m.ResultSubmittedBy == userID {
		return nil, fmt.Errorf("only the opposing player can confirm")
	}

	completed, err := s.repo.ConfirmMatchResult(ctx, matchID)
	if err != nil {
		return nil, fmt.Errorf("confirm: %w", err)
	}

	s.advanceWinner(ctx, &completed)
	s.emitResultNotifications(ctx, &completed)
	return &completed, nil
}

func (s *Service) Dispute(ctx context.Context, matchID, userID, reason string) (*repository.Match, error) {
	m, err := s.repo.GetMatch(ctx, matchID)
	if err != nil {
		return nil, fmt.Errorf("match not found")
	}
	if m.Status != "awaiting_confirmation" {
		return nil, fmt.Errorf("can only dispute a result that is awaiting confirmation")
	}
	if m.PlayerAID == nil || m.PlayerBID == nil {
		return nil, fmt.Errorf("match missing players")
	}
	if userID != *m.PlayerAID && userID != *m.PlayerBID {
		return nil, fmt.Errorf("only participants can dispute")
	}
	if reason == "" {
		return nil, fmt.Errorf("please describe the issue")
	}

	disputed, err := s.repo.OpenDispute(ctx, repository.OpenDisputeParams{
		ID:       matchID,
		OpenedBy: userID,
		Reason:   reason,
	})
	if err != nil {
		return nil, fmt.Errorf("open dispute: %w", err)
	}

	if s.auditor != nil {
		mid := matchID
		s.auditor.Log(ctx, userID, "player", "match.dispute_opened", "match", &mid, map[string]any{"reason": reason})
	}

	// Notify the organizer.
	tournament, err := s.repo.GetTournament(ctx, m.TournamentID)
	if err == nil && s.notifier != nil {
		link := fmt.Sprintf("/matches/%s", matchID)
		s.notifier.Emit(ctx, repository.CreateNotificationParams{
			UserID:  tournament.OrganizerID,
			Type:    "dispute_opened",
			Title:   "Match dispute opened",
			Message: fmt.Sprintf("A player disputed a match in %s.", tournament.Title),
			Link:    &link,
		})
	}

	return &disputed, nil
}

func (s *Service) ResolveDispute(ctx context.Context, matchID, organizerID, winnerID, score string) (*repository.Match, error) {
	m, err := s.repo.GetMatch(ctx, matchID)
	if err != nil {
		return nil, fmt.Errorf("match not found")
	}
	tournament, err := s.repo.GetTournament(ctx, m.TournamentID)
	if err != nil {
		return nil, fmt.Errorf("tournament not found")
	}
	if tournament.OrganizerID != organizerID {
		return nil, fmt.Errorf("only the organizer can resolve disputes")
	}
	if m.DisputeStatus != "pending" {
		return nil, fmt.Errorf("no dispute pending")
	}
	if m.PlayerAID == nil || m.PlayerBID == nil {
		return nil, fmt.Errorf("match missing players")
	}
	if winnerID != *m.PlayerAID && winnerID != *m.PlayerBID {
		return nil, fmt.Errorf("winner must be one of the match participants")
	}

	resolved, err := s.repo.ResolveDispute(ctx, repository.ResolveDisputeParams{
		ID:       matchID,
		WinnerID: winnerID,
		Score:    score,
	})
	if err != nil {
		return nil, fmt.Errorf("resolve dispute: %w", err)
	}

	if s.auditor != nil {
		mid := matchID
		s.auditor.Log(ctx, organizerID, "organizer", "match.dispute_resolved", "match", &mid, map[string]any{
			"winner": winnerID, "score": score,
		})
	}

	s.advanceWinner(ctx, &resolved)
	s.emitResultNotifications(ctx, &resolved)
	return &resolved, nil
}

func (s *Service) ListMyMatches(ctx context.Context, userID string) ([]repository.Match, error) {
	return s.repo.ListOpenMatchesForUser(ctx, userID)
}

func (s *Service) ListDisputesByOrganizer(ctx context.Context, organizerID string) ([]repository.Match, error) {
	return s.repo.ListDisputedMatchesByOrganizer(ctx, organizerID)
}

func (s *Service) GetByID(ctx context.Context, id string) (*repository.Match, error) {
	m, err := s.repo.GetMatch(ctx, id)
	if err != nil {
		return nil, err
	}
	return &m, nil
}

// advanceWinner promotes the match winner into the next-round bracket slot.
// Called after a result is confirmed or a dispute is resolved (F9).
func (s *Service) advanceWinner(ctx context.Context, m *repository.Match) {
	if m.WinnerID == nil {
		return
	}
	finalRound, err := s.finalRound(ctx, m.TournamentID)
	if err != nil {
		return
	}
	if finalRound == 0 {
		return
	}
	if m.Round >= finalRound {
		s.finalize(ctx, m.TournamentID)
		return
	}

	nextRound := m.Round + 1
	nextPosition := m.Position / 2

	existing, err := s.repo.GetMatchesByTournamentAndRound(ctx, repository.GetMatchesByTournamentAndRoundParams{
		TournamentID: m.TournamentID,
		Round:        nextRound,
	})
	if err != nil {
		return
	}

	var next *repository.Match
	for i := range existing {
		if existing[i].Position == nextPosition {
			next = &existing[i]
			break
		}
	}

	if next == nil {
		created, err := s.repo.CreateMatch(ctx, repository.CreateMatchParams{
			TournamentID: m.TournamentID,
			Round:        nextRound,
			Position:     nextPosition,
			Status:       "pending",
		})
		if err != nil {
			return
		}
		next = &created
	}

	winner := *m.WinnerID
	if m.Position%2 == 0 {
		s.repo.UpdateMatchPlayers(ctx, repository.UpdateMatchPlayersParams{ID: next.ID, PlayerAID: &winner})
	} else {
		s.repo.UpdateMatchPlayers(ctx, repository.UpdateMatchPlayersParams{ID: next.ID, PlayerBID: &winner})
	}
}

func (s *Service) finalRound(ctx context.Context, tournamentID string) (int32, error) {
	firstRound, err := s.repo.GetMatchesByTournamentAndRound(ctx, repository.GetMatchesByTournamentAndRoundParams{
		TournamentID: tournamentID,
		Round:        1,
	})
	if err != nil {
		return 0, err
	}
	if len(firstRound) == 0 {
		return 0, nil
	}

	playerSlots := len(firstRound) * 2
	var round int32
	for playerSlots > 1 {
		round++
		playerSlots = playerSlots / 2
	}
	return round, nil
}

func (s *Service) finalize(ctx context.Context, tournamentID string) {
	t, err := s.repo.GetTournament(ctx, tournamentID)
	if err != nil {
		return
	}
	if t.Status == "completed" {
		return
	}
	if _, err := s.repo.UpdateTournamentStatus(ctx, repository.UpdateTournamentStatusParams{
		ID:     tournamentID,
		Status: "completed",
	}); err != nil {
		log.Printf("finalize tournament %s: %v", tournamentID, err)
		return
	}
	if s.payout != nil {
		if _, err := s.payout.EnsureForTournament(ctx, tournamentID); err != nil {
			log.Printf("ensure payout %s: %v", tournamentID, err)
		}
	}
}

func (s *Service) emitResultNotifications(ctx context.Context, m *repository.Match) {
	if s.notifier == nil || m.WinnerID == nil {
		return
	}
	link := fmt.Sprintf("/tournaments/%s/bracket", m.TournamentID)
	loser := opponentID(m, *m.WinnerID)
	s.notifier.Emit(ctx, repository.CreateNotificationParams{
		UserID:  *m.WinnerID,
		Type:    "match_won",
		Title:   "Match won",
		Message: "You advanced! Check the bracket for your next match.",
		Link:    &link,
	})
	if loser != "" {
		s.notifier.Emit(ctx, repository.CreateNotificationParams{
			UserID:  loser,
			Type:    "match_lost",
			Title:   "Match result",
			Message: "Your match is over. Review the result in the bracket.",
			Link:    &link,
		})
	}
}

func opponentID(m *repository.Match, userID string) string {
	if m.PlayerAID != nil && *m.PlayerAID == userID && m.PlayerBID != nil {
		return *m.PlayerBID
	}
	if m.PlayerBID != nil && *m.PlayerBID == userID && m.PlayerAID != nil {
		return *m.PlayerAID
	}
	return ""
}
