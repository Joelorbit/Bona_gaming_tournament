-- name: CreateProfile :one
INSERT INTO profiles (id, username, display_name, email, avatar_url, role, bio, country, country_code)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
RETURNING id, username, display_name, email, avatar_url, role, bio, country, country_code, created_at, updated_at;

-- name: GetProfile :one
SELECT id, username, display_name, email, avatar_url, role, bio, country, country_code, created_at, updated_at
FROM profiles
WHERE id = $1;

-- name: GetProfileByUsername :one
SELECT id, username, display_name, email, avatar_url, role, bio, country, country_code, created_at, updated_at
FROM profiles
WHERE username = $1;

-- name: UpdateProfile :one
UPDATE profiles
SET username = COALESCE(sqlc.narg(username), username),
    display_name = COALESCE(sqlc.narg(display_name), display_name),
    avatar_url = COALESCE(sqlc.narg(avatar_url), avatar_url),
    bio = COALESCE(sqlc.narg(bio), bio),
    country = COALESCE(sqlc.narg(country), country),
    country_code = COALESCE(sqlc.narg(country_code), country_code),
    updated_at = NOW()
WHERE id = $1
RETURNING id, username, display_name, email, avatar_url, role, bio, country, country_code, created_at, updated_at;

-- name: GetProfileStats :one
SELECT
    (SELECT COUNT(*) FROM registrations WHERE user_id = $1 AND payment_status = 'paid') AS tournaments_played,
    (SELECT COUNT(*) FROM tournaments WHERE organizer_id = $1) AS tournaments_hosted,
    (SELECT COUNT(DISTINCT tournament_id) FROM matches WHERE winner_id = $1
        AND round = (SELECT MAX(round) FROM matches m2 WHERE m2.tournament_id = matches.tournament_id)
    ) AS wins;

-- name: DeleteProfile :exec
DELETE FROM profiles WHERE id = $1;
