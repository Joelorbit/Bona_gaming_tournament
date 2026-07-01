-- name: CreatePayout :one
INSERT INTO payouts (tournament_id, winner_id, amount, currency, status, note)
VALUES ($1, $2, $3, $4, $5, $6)
ON CONFLICT (tournament_id, winner_id) DO UPDATE SET amount = EXCLUDED.amount
RETURNING id, tournament_id, winner_id, amount, currency, status, paid_at, paid_by, note, created_at, updated_at;

-- name: GetPayout :one
SELECT id, tournament_id, winner_id, amount, currency, status, paid_at, paid_by, note, created_at, updated_at
FROM payouts WHERE id = $1;

-- name: ListPayoutsByTournament :many
SELECT id, tournament_id, winner_id, amount, currency, status, paid_at, paid_by, note, created_at, updated_at
FROM payouts WHERE tournament_id = $1 ORDER BY created_at DESC;

-- name: ListPayoutsByWinner :many
SELECT id, tournament_id, winner_id, amount, currency, status, paid_at, paid_by, note, created_at, updated_at
FROM payouts WHERE winner_id = $1 ORDER BY created_at DESC;

-- name: ListPayoutsByOrganizer :many
SELECT p.id, p.tournament_id, p.winner_id, p.amount, p.currency, p.status, p.paid_at, p.paid_by, p.note, p.created_at, p.updated_at
FROM payouts p
JOIN tournaments t ON t.id = p.tournament_id
WHERE t.organizer_id = $1
ORDER BY p.created_at DESC;

-- name: MarkPayoutPaid :one
UPDATE payouts
SET status = 'paid', paid_at = NOW(), paid_by = $2, note = COALESCE($3, note), updated_at = NOW()
WHERE id = $1 AND status != 'paid'
RETURNING id, tournament_id, winner_id, amount, currency, status, paid_at, paid_by, note, created_at, updated_at;
