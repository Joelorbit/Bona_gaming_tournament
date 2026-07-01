-- name: CreateTournament :one
INSERT INTO tournaments (
    title, game, description, rules, format, max_participants, min_participants,
    team_size, entry_fee, prize_pool, currency, location, best_of, platform_fee_pct,
    organizer_fee_pct, start_date, end_date, registration_deadline,
    registration_close_at, organizer_id, banner_url
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19, $20, $21
)
RETURNING id, title, game, description, rules, format, status, max_participants,
          min_participants, team_size, entry_fee, prize_pool, currency, location,
          best_of, platform_fee_pct, organizer_fee_pct,
          start_date, end_date, registration_deadline, registration_close_at,
          organizer_id, banner_url, created_at, updated_at;

-- name: GetTournament :one
SELECT id, title, game, description, rules, format, status, max_participants,
       min_participants, team_size, entry_fee, prize_pool, currency, location,
       best_of, platform_fee_pct, organizer_fee_pct,
       start_date, end_date, registration_deadline, registration_close_at,
       organizer_id, banner_url, created_at, updated_at
FROM tournaments
WHERE id = $1;

-- name: ListTournaments :many
SELECT id, title, game, description, rules, format, status, max_participants,
       min_participants, team_size, entry_fee, prize_pool, currency, location,
       best_of, platform_fee_pct, organizer_fee_pct,
       start_date, end_date, registration_deadline, registration_close_at,
       organizer_id, banner_url, created_at, updated_at
FROM tournaments
WHERE (sqlc.arg(status) = '' OR status = sqlc.arg(status))
  AND (sqlc.arg(game) = '' OR game = sqlc.arg(game))
  AND (sqlc.arg(q) = '' OR title ILIKE '%' || sqlc.arg(q) || '%' OR game ILIKE '%' || sqlc.arg(q) || '%')
ORDER BY created_at DESC
LIMIT sqlc.arg(limit) OFFSET sqlc.arg(offset);

-- name: ListTournamentsByOrganizer :many
SELECT id, title, game, description, rules, format, status, max_participants,
       min_participants, team_size, entry_fee, prize_pool, currency, location,
       best_of, platform_fee_pct, organizer_fee_pct,
       start_date, end_date, registration_deadline, registration_close_at,
       organizer_id, banner_url, created_at, updated_at
FROM tournaments
WHERE organizer_id = $1
ORDER BY created_at DESC;

-- name: UpdateTournament :one
UPDATE tournaments
SET title = COALESCE(sqlc.narg('title'), title),
    game = COALESCE(sqlc.narg('game'), game),
    description = COALESCE(sqlc.narg('description'), description),
    rules = COALESCE(sqlc.narg('rules'), rules),
    format = COALESCE(sqlc.narg('format'), format),
    max_participants = COALESCE(sqlc.narg('max_participants'), max_participants),
    min_participants = COALESCE(sqlc.narg('min_participants'), min_participants),
    team_size = COALESCE(sqlc.narg('team_size'), team_size),
    entry_fee = COALESCE(sqlc.narg('entry_fee'), entry_fee),
    prize_pool = COALESCE(sqlc.narg('prize_pool'), prize_pool),
    currency = COALESCE(sqlc.narg('currency'), currency),
    location = COALESCE(sqlc.narg('location'), location),
    best_of = COALESCE(sqlc.narg('best_of'), best_of),
    organizer_fee_pct = COALESCE(sqlc.narg('organizer_fee_pct'), organizer_fee_pct),
    start_date = COALESCE(sqlc.narg('start_date'), start_date),
    end_date = COALESCE(sqlc.narg('end_date'), end_date),
    registration_deadline = COALESCE(sqlc.narg('registration_deadline'), registration_deadline),
    registration_close_at = COALESCE(sqlc.narg('registration_close_at'), registration_close_at),
    banner_url = COALESCE(sqlc.narg('banner_url'), banner_url),
    updated_at = NOW()
WHERE id = sqlc.arg('id')
RETURNING id, title, game, description, rules, format, status, max_participants,
          min_participants, team_size, entry_fee, prize_pool, currency, location,
          best_of, platform_fee_pct, organizer_fee_pct,
          start_date, end_date, registration_deadline, registration_close_at,
          organizer_id, banner_url, created_at, updated_at;

-- name: UpdateTournamentStatus :one
UPDATE tournaments
SET status = $2, updated_at = NOW()
WHERE id = $1
RETURNING id, title, game, description, rules, format, status, max_participants,
          min_participants, team_size, entry_fee, prize_pool, currency, location,
          best_of, platform_fee_pct, organizer_fee_pct,
          start_date, end_date, registration_deadline, registration_close_at,
          organizer_id, banner_url, created_at, updated_at;

-- name: DeleteTournament :exec
DELETE FROM tournaments WHERE id = $1;
