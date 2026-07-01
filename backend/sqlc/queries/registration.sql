-- name: CreateRegistration :one
INSERT INTO registrations (tournament_id, user_id, payment_status)
VALUES ($1, $2, $3)
RETURNING id, tournament_id, user_id, payment_status, seed, registered_at;

-- name: GetRegistrationByUserAndTournament :one
SELECT id, tournament_id, user_id, payment_status, seed, registered_at
FROM registrations
WHERE user_id = $1 AND tournament_id = $2;

-- name: ListRegistrationsByTournament :many
SELECT r.id, r.tournament_id, r.user_id, r.payment_status, r.seed, r.registered_at,
       p.username, p.display_name, p.avatar_url
FROM registrations r
JOIN profiles p ON r.user_id = p.id
WHERE r.tournament_id = $1
ORDER BY r.registered_at;

-- name: ListRegistrationsByUser :many
SELECT r.id, r.tournament_id, r.user_id, r.payment_status, r.seed, r.registered_at,
       t.title, t.game, t.status, t.start_date, t.prize_pool
FROM registrations r
JOIN tournaments t ON r.tournament_id = t.id
WHERE r.user_id = $1
ORDER BY r.registered_at DESC;

-- name: CountPaidRegistrations :one
SELECT COUNT(*)
FROM registrations
WHERE tournament_id = $1 AND payment_status = 'paid';

-- name: CountActiveRegistrations :one
SELECT COUNT(*)
FROM registrations
WHERE tournament_id = $1 AND payment_status IN ('pending', 'paid');

-- name: DeleteRegistration :exec
DELETE FROM registrations WHERE user_id = $1 AND tournament_id = $2;

-- name: UpdateRegistrationPaymentStatus :one
UPDATE registrations
SET payment_status = $3
WHERE user_id = $1 AND tournament_id = $2
RETURNING id, tournament_id, user_id, payment_status, seed, registered_at;
