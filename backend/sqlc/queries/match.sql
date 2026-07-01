-- name: CreateMatch :one
INSERT INTO matches (tournament_id, round, position, player_a_id, player_b_id, status)
VALUES ($1, $2, $3, $4, $5, $6)
RETURNING id, tournament_id, round, position, player_a_id, player_b_id, winner_id, score,
          status, scheduled_at, completed_at,
          evidence_screenshot_url, evidence_video_url, evidence_notes,
          result_submitted_by, result_confirmed_at,
          dispute_status, dispute_reason, dispute_opened_by, dispute_opened_at, dispute_resolved_at,
          created_at, updated_at;

-- name: GetMatch :one
SELECT id, tournament_id, round, position, player_a_id, player_b_id, winner_id, score,
       status, scheduled_at, completed_at,
       evidence_screenshot_url, evidence_video_url, evidence_notes,
       result_submitted_by, result_confirmed_at,
       dispute_status, dispute_reason, dispute_opened_by, dispute_opened_at, dispute_resolved_at,
       created_at, updated_at
FROM matches
WHERE id = $1;

-- name: ListMatchesByTournament :many
SELECT id, tournament_id, round, position, player_a_id, player_b_id, winner_id, score,
       status, scheduled_at, completed_at,
       evidence_screenshot_url, evidence_video_url, evidence_notes,
       result_submitted_by, result_confirmed_at,
       dispute_status, dispute_reason, dispute_opened_by, dispute_opened_at, dispute_resolved_at,
       created_at, updated_at
FROM matches
WHERE tournament_id = $1
ORDER BY round, position;

-- name: GetMatchesByTournamentAndRound :many
SELECT id, tournament_id, round, position, player_a_id, player_b_id, winner_id, score,
       status, scheduled_at, completed_at,
       evidence_screenshot_url, evidence_video_url, evidence_notes,
       result_submitted_by, result_confirmed_at,
       dispute_status, dispute_reason, dispute_opened_by, dispute_opened_at, dispute_resolved_at,
       created_at, updated_at
FROM matches
WHERE tournament_id = $1 AND round = $2
ORDER BY position;

-- name: GetMaxRound :one
SELECT COALESCE(MAX(round), 0)::int FROM matches WHERE tournament_id = $1;

-- name: SubmitMatchResult :one
UPDATE matches
SET winner_id = $2,
    score = $3,
    result_submitted_by = $4,
    evidence_screenshot_url = COALESCE($5, evidence_screenshot_url),
    evidence_video_url = COALESCE($6, evidence_video_url),
    evidence_notes = COALESCE($7, evidence_notes),
    status = 'awaiting_confirmation',
    updated_at = NOW()
WHERE id = $1
RETURNING id, tournament_id, round, position, player_a_id, player_b_id, winner_id, score,
          status, scheduled_at, completed_at,
          evidence_screenshot_url, evidence_video_url, evidence_notes,
          result_submitted_by, result_confirmed_at,
          dispute_status, dispute_reason, dispute_opened_by, dispute_opened_at, dispute_resolved_at,
          created_at, updated_at;

-- name: ConfirmMatchResult :one
UPDATE matches
SET status = 'completed', completed_at = NOW(), result_confirmed_at = NOW(), updated_at = NOW()
WHERE id = $1
RETURNING id, tournament_id, round, position, player_a_id, player_b_id, winner_id, score,
          status, scheduled_at, completed_at,
          evidence_screenshot_url, evidence_video_url, evidence_notes,
          result_submitted_by, result_confirmed_at,
          dispute_status, dispute_reason, dispute_opened_by, dispute_opened_at, dispute_resolved_at,
          created_at, updated_at;

-- name: OpenDispute :one
UPDATE matches
SET status = 'disputed',
    dispute_status = 'pending',
    dispute_reason = $3,
    dispute_opened_by = $2,
    dispute_opened_at = NOW(),
    updated_at = NOW()
WHERE id = $1
RETURNING id, tournament_id, round, position, player_a_id, player_b_id, winner_id, score,
          status, scheduled_at, completed_at,
          evidence_screenshot_url, evidence_video_url, evidence_notes,
          result_submitted_by, result_confirmed_at,
          dispute_status, dispute_reason, dispute_opened_by, dispute_opened_at, dispute_resolved_at,
          created_at, updated_at;

-- name: ResolveDispute :one
UPDATE matches
SET winner_id = $2,
    score = COALESCE(NULLIF($3, ''), score),
    status = 'completed',
    dispute_status = 'resolved',
    dispute_resolved_at = NOW(),
    completed_at = NOW(),
    result_confirmed_at = NOW(),
    updated_at = NOW()
WHERE id = $1
RETURNING id, tournament_id, round, position, player_a_id, player_b_id, winner_id, score,
          status, scheduled_at, completed_at,
          evidence_screenshot_url, evidence_video_url, evidence_notes,
          result_submitted_by, result_confirmed_at,
          dispute_status, dispute_reason, dispute_opened_by, dispute_opened_at, dispute_resolved_at,
          created_at, updated_at;

-- name: UpdateMatchPlayers :one
UPDATE matches
SET player_a_id = COALESCE(sqlc.narg(player_a_id), player_a_id),
    player_b_id = COALESCE(sqlc.narg(player_b_id), player_b_id),
    updated_at = NOW()
WHERE id = sqlc.arg(id)
RETURNING id, tournament_id, round, position, player_a_id, player_b_id, winner_id, score,
          status, scheduled_at, completed_at,
          evidence_screenshot_url, evidence_video_url, evidence_notes,
          result_submitted_by, result_confirmed_at,
          dispute_status, dispute_reason, dispute_opened_by, dispute_opened_at, dispute_resolved_at,
          created_at, updated_at;
