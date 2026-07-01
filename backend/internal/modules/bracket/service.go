package bracket

import (
	"context"
	"fmt"
	"math"
	"sort"

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

func (s *Service) Generate(ctx context.Context, tournamentID, userID string) ([]repository.Match, error) {
	tournament, err := s.repo.GetTournament(ctx, tournamentID)
	if err != nil {
		return nil, fmt.Errorf("tournament not found")
	}

	if tournament.OrganizerID != userID {
		return nil, fmt.Errorf("only the organizer can generate the bracket")
	}

	if tournament.Status != "registration_closed" {
		return nil, fmt.Errorf("registration must be closed before generating bracket")
	}

	registrations, err := s.repo.ListRegistrationsByTournament(ctx, tournamentID)
	if err != nil {
		return nil, fmt.Errorf("fetch registrations: %w", err)
	}

	var paidPlayers []string
	for _, reg := range registrations {
		if reg.PaymentStatus == "paid" {
			paidPlayers = append(paidPlayers, reg.UserID)
		}
	}

	if len(paidPlayers) < 2 {
		return nil, fmt.Errorf("need at least 2 paid players to generate bracket")
	}

	playerCount := nextPowerOf2(len(paidPlayers))
	byes := playerCount - len(paidPlayers)

	sort.Strings(paidPlayers)
	seeded := make([]string, playerCount)
	for i, p := range paidPlayers {
		if i%2 == 0 {
			seeded[i/2] = p
		} else {
			seeded[playerCount-1-i/2] = p
		}
	}

	_, err = s.repo.UpdateTournamentStatus(ctx, repository.UpdateTournamentStatusParams{
		ID:     tournamentID,
		Status: "in_progress",
	})
	if err != nil {
		return nil, fmt.Errorf("update tournament status: %w", err)
	}

	var createdMatches []repository.Match
	round := 1
	for i := 0; i < playerCount; i += 2 {
		playerA := seeded[i]
		playerB := ""
		if i+1 < playerCount {
			playerB = seeded[i+1]
		}
		if playerB == "" && byes > 0 {
			byes--
		}

		var pA, pB *string
		if playerA != "" {
			pA = &playerA
		}
		if playerB != "" {
			pB = &playerB
		}

		match, err := s.repo.CreateMatch(ctx, repository.CreateMatchParams{
			TournamentID: tournamentID,
			Round:        int32(round),
			Position:     int32(i / 2),
			PlayerAID:    pA,
			PlayerBID:    pB,
			Status:       "pending",
		})
		if err != nil {
			return nil, fmt.Errorf("create match: %w", err)
		}

		if pA == nil && pB != nil {
			pA = pB
			pB = nil
			match, err = s.repo.UpdateMatchPlayers(ctx, repository.UpdateMatchPlayersParams{
				ID:        match.ID,
				PlayerAID: pA,
			})
			if err != nil {
				return nil, fmt.Errorf("assign bye player: %w", err)
			}
		}

		if pA != nil && pB == nil {
			_, err = s.repo.UpdateMatchResult(ctx, repository.UpdateMatchResultParams{
				ID:       match.ID,
				WinnerID: pA,
				Score:    "BYE",
			})
			if err != nil {
				return nil, fmt.Errorf("auto-advance bye: %w", err)
			}
		}

		createdMatches = append(createdMatches, match)
	}

	if s.notifier != nil {
		link := fmt.Sprintf("/tournaments/%s/bracket", tournamentID)
		for _, p := range paidPlayers {
			s.notifier.Emit(ctx, repository.CreateNotificationParams{
				UserID:  p,
				Type:    "bracket_generated",
				Title:   tournament.Title + " bracket is live",
				Message: "The bracket has been generated. Check your first opponent.",
				Link:    &link,
			})
		}
	}

	return createdMatches, nil
}

func (s *Service) GetBracket(ctx context.Context, tournamentID string) ([]repository.Match, error) {
	matches, err := s.repo.ListMatchesByTournament(ctx, tournamentID)
	if err != nil {
		return nil, err
	}
	return matches, nil
}

func nextPowerOf2(n int) int {
	if n <= 0 {
		return 1
	}
	return int(math.Pow(2, math.Ceil(math.Log2(float64(n)))))
}
