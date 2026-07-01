package repository

import "context"

type CreateMatchParams struct {
	TournamentID string  `json:"tournament_id"`
	Round        int32   `json:"round"`
	Position     int32   `json:"position"`
	PlayerAID    *string `json:"player_a_id,omitempty"`
	PlayerBID    *string `json:"player_b_id,omitempty"`
	Status       string  `json:"status"`
}

const matchColumns = `id, tournament_id, round, position, player_a_id, player_b_id,
    winner_id, score, status, scheduled_at, completed_at,
    evidence_screenshot_url, evidence_video_url, evidence_notes,
    result_submitted_by, result_confirmed_at,
    dispute_status, dispute_reason, dispute_opened_by, dispute_opened_at, dispute_resolved_at,
    created_at, updated_at`

func scanMatch(row interface {
	Scan(...any) error
}) (Match, error) {
	var m Match
	err := row.Scan(
		&m.ID, &m.TournamentID, &m.Round, &m.Position, &m.PlayerAID, &m.PlayerBID,
		&m.WinnerID, &m.Score, &m.Status, &m.ScheduledAt, &m.CompletedAt,
		&m.EvidenceScreenshotURL, &m.EvidenceVideoURL, &m.EvidenceNotes,
		&m.ResultSubmittedBy, &m.ResultConfirmedAt,
		&m.DisputeStatus, &m.DisputeReason, &m.DisputeOpenedBy, &m.DisputeOpenedAt, &m.DisputeResolvedAt,
		&m.CreatedAt, &m.UpdatedAt,
	)
	return m, err
}

const createMatch = `
INSERT INTO matches (tournament_id, round, position, player_a_id, player_b_id, status)
VALUES ($1, $2, $3, $4, $5, $6)
RETURNING ` + matchColumns

func (q *Queries) CreateMatch(ctx context.Context, p CreateMatchParams) (Match, error) {
	row := q.db.QueryRow(ctx, createMatch, p.TournamentID, p.Round, p.Position, p.PlayerAID, p.PlayerBID, p.Status)
	return scanMatch(row)
}

const getMatch = `SELECT ` + matchColumns + ` FROM matches WHERE id = $1`

func (q *Queries) GetMatch(ctx context.Context, id string) (Match, error) {
	return scanMatch(q.db.QueryRow(ctx, getMatch, id))
}

const listMatchesByTournament = `
SELECT ` + matchColumns + `
FROM matches WHERE tournament_id = $1
ORDER BY round, position`

func (q *Queries) ListMatchesByTournament(ctx context.Context, tournamentID string) ([]Match, error) {
	rows, err := q.db.Query(ctx, listMatchesByTournament, tournamentID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []Match
	for rows.Next() {
		m, err := scanMatch(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, m)
	}
	return out, rows.Err()
}

type GetMatchesByTournamentAndRoundParams struct {
	TournamentID string `json:"tournament_id"`
	Round        int32  `json:"round"`
}

const getMatchesByTournamentAndRound = `
SELECT ` + matchColumns + `
FROM matches WHERE tournament_id = $1 AND round = $2
ORDER BY position`

func (q *Queries) GetMatchesByTournamentAndRound(ctx context.Context, p GetMatchesByTournamentAndRoundParams) ([]Match, error) {
	rows, err := q.db.Query(ctx, getMatchesByTournamentAndRound, p.TournamentID, p.Round)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []Match
	for rows.Next() {
		m, err := scanMatch(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, m)
	}
	return out, rows.Err()
}

const listOpenMatchesForUser = `
SELECT ` + matchColumns + `
FROM matches
WHERE (player_a_id = $1 OR player_b_id = $1)
  AND status IN ('pending', 'in_progress', 'awaiting_confirmation', 'disputed')
ORDER BY round, position`

func (q *Queries) ListOpenMatchesForUser(ctx context.Context, userID string) ([]Match, error) {
	rows, err := q.db.Query(ctx, listOpenMatchesForUser, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []Match
	for rows.Next() {
		m, err := scanMatch(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, m)
	}
	return out, rows.Err()
}

const listDisputedMatchesByOrganizer = `
SELECT m.` + `id, m.tournament_id, m.round, m.position, m.player_a_id, m.player_b_id,
    m.winner_id, m.score, m.status, m.scheduled_at, m.completed_at,
    m.evidence_screenshot_url, m.evidence_video_url, m.evidence_notes,
    m.result_submitted_by, m.result_confirmed_at,
    m.dispute_status, m.dispute_reason, m.dispute_opened_by, m.dispute_opened_at, m.dispute_resolved_at,
    m.created_at, m.updated_at
FROM matches m
JOIN tournaments t ON t.id = m.tournament_id
WHERE t.organizer_id = $1 AND m.dispute_status = 'pending'
ORDER BY m.dispute_opened_at ASC`

func (q *Queries) ListDisputedMatchesByOrganizer(ctx context.Context, organizerID string) ([]Match, error) {
	rows, err := q.db.Query(ctx, listDisputedMatchesByOrganizer, organizerID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []Match
	for rows.Next() {
		m, err := scanMatch(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, m)
	}
	return out, rows.Err()
}

const getMaxRound = `SELECT COALESCE(MAX(round), 0)::int FROM matches WHERE tournament_id = $1`

func (q *Queries) GetMaxRound(ctx context.Context, tournamentID string) (int32, error) {
	var n int32
	err := q.db.QueryRow(ctx, getMaxRound, tournamentID).Scan(&n)
	return n, err
}

type SubmitMatchResultParams struct {
	ID                    string  `json:"id"`
	WinnerID              string  `json:"winner_id"`
	Score                 string  `json:"score"`
	SubmittedBy           string  `json:"submitted_by"`
	EvidenceScreenshotURL *string `json:"evidence_screenshot_url,omitempty"`
	EvidenceVideoURL      *string `json:"evidence_video_url,omitempty"`
	EvidenceNotes         *string `json:"evidence_notes,omitempty"`
}

const submitMatchResult = `
UPDATE matches SET
    winner_id = $2,
    score = $3,
    result_submitted_by = $4,
    evidence_screenshot_url = COALESCE($5, evidence_screenshot_url),
    evidence_video_url = COALESCE($6, evidence_video_url),
    evidence_notes = COALESCE($7, evidence_notes),
    status = 'awaiting_confirmation',
    updated_at = NOW()
WHERE id = $1
RETURNING ` + matchColumns

func (q *Queries) SubmitMatchResult(ctx context.Context, p SubmitMatchResultParams) (Match, error) {
	return scanMatch(q.db.QueryRow(ctx, submitMatchResult,
		p.ID, p.WinnerID, p.Score, p.SubmittedBy,
		p.EvidenceScreenshotURL, p.EvidenceVideoURL, p.EvidenceNotes,
	))
}

const confirmMatchResult = `
UPDATE matches SET
    status = 'completed',
    completed_at = NOW(),
    result_confirmed_at = NOW(),
    updated_at = NOW()
WHERE id = $1
RETURNING ` + matchColumns

func (q *Queries) ConfirmMatchResult(ctx context.Context, id string) (Match, error) {
	return scanMatch(q.db.QueryRow(ctx, confirmMatchResult, id))
}

type OpenDisputeParams struct {
	ID         string `json:"id"`
	OpenedBy   string `json:"opened_by"`
	Reason     string `json:"reason"`
}

const openDispute = `
UPDATE matches SET
    status = 'disputed',
    dispute_status = 'pending',
    dispute_reason = $3,
    dispute_opened_by = $2,
    dispute_opened_at = NOW(),
    updated_at = NOW()
WHERE id = $1
RETURNING ` + matchColumns

func (q *Queries) OpenDispute(ctx context.Context, p OpenDisputeParams) (Match, error) {
	return scanMatch(q.db.QueryRow(ctx, openDispute, p.ID, p.OpenedBy, p.Reason))
}

type ResolveDisputeParams struct {
	ID       string `json:"id"`
	WinnerID string `json:"winner_id"`
	Score    string `json:"score"`
}

const resolveDispute = `
UPDATE matches SET
    winner_id = $2,
    score = COALESCE(NULLIF($3, ''), score),
    status = 'completed',
    dispute_status = 'resolved',
    dispute_resolved_at = NOW(),
    completed_at = NOW(),
    result_confirmed_at = NOW(),
    updated_at = NOW()
WHERE id = $1
RETURNING ` + matchColumns

func (q *Queries) ResolveDispute(ctx context.Context, p ResolveDisputeParams) (Match, error) {
	return scanMatch(q.db.QueryRow(ctx, resolveDispute, p.ID, p.WinnerID, p.Score))
}

type UpdateMatchResultParams struct {
	ID       string  `json:"id"`
	WinnerID *string `json:"winner_id,omitempty"`
	Score    string  `json:"score"`
}

// UpdateMatchResult forcibly completes a match (used for byes during bracket generation).
const updateMatchResult = `
UPDATE matches SET
    winner_id = $2, score = $3, status = 'completed',
    completed_at = NOW(), result_confirmed_at = NOW(), updated_at = NOW()
WHERE id = $1
RETURNING ` + matchColumns

func (q *Queries) UpdateMatchResult(ctx context.Context, p UpdateMatchResultParams) (Match, error) {
	return scanMatch(q.db.QueryRow(ctx, updateMatchResult, p.ID, p.WinnerID, p.Score))
}

type UpdateMatchPlayersParams struct {
	ID        string  `json:"id"`
	PlayerAID *string `json:"player_a_id,omitempty"`
	PlayerBID *string `json:"player_b_id,omitempty"`
}

const updateMatchPlayers = `
UPDATE matches SET
    player_a_id = COALESCE($2, player_a_id),
    player_b_id = COALESCE($3, player_b_id)
WHERE id = $1
RETURNING ` + matchColumns

func (q *Queries) UpdateMatchPlayers(ctx context.Context, p UpdateMatchPlayersParams) (Match, error) {
	return scanMatch(q.db.QueryRow(ctx, updateMatchPlayers, p.ID, p.PlayerAID, p.PlayerBID))
}
