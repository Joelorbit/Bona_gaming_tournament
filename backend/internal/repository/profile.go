package repository

import "context"

type CreateProfileParams struct {
	ID          string  `json:"id"`
	Username    string  `json:"username"`
	DisplayName *string `json:"display_name,omitempty"`
	Email       *string `json:"email,omitempty"`
	AvatarUrl   *string `json:"avatar_url,omitempty"`
	Role        string  `json:"role"`
	Bio         *string `json:"bio,omitempty"`
	Country     *string `json:"country,omitempty"`
	CountryCode *string `json:"country_code,omitempty"`
}

const createProfile = `
INSERT INTO profiles (id, username, display_name, email, avatar_url, role, bio, country, country_code)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
RETURNING id, username, display_name, email, avatar_url, role, bio, country, country_code, created_at, updated_at`

func (q *Queries) CreateProfile(ctx context.Context, p CreateProfileParams) (Profile, error) {
	row := q.db.QueryRow(ctx, createProfile, p.ID, p.Username, p.DisplayName, p.Email, p.AvatarUrl, p.Role, p.Bio, p.Country, p.CountryCode)
	var pr Profile
	err := row.Scan(&pr.ID, &pr.Username, &pr.DisplayName, &pr.Email, &pr.AvatarUrl, &pr.Role, &pr.Bio, &pr.Country, &pr.CountryCode, &pr.CreatedAt, &pr.UpdatedAt)
	return pr, err
}

const getProfile = `
SELECT id, username, display_name, email, avatar_url, role, bio, country, country_code, created_at, updated_at
FROM profiles WHERE id = $1`

func (q *Queries) GetProfile(ctx context.Context, id string) (Profile, error) {
	row := q.db.QueryRow(ctx, getProfile, id)
	var pr Profile
	err := row.Scan(&pr.ID, &pr.Username, &pr.DisplayName, &pr.Email, &pr.AvatarUrl, &pr.Role, &pr.Bio, &pr.Country, &pr.CountryCode, &pr.CreatedAt, &pr.UpdatedAt)
	return pr, err
}

const getProfileByUsername = `
SELECT id, username, display_name, email, avatar_url, role, bio, country, country_code, created_at, updated_at
FROM profiles WHERE username = $1`

func (q *Queries) GetProfileByUsername(ctx context.Context, username string) (Profile, error) {
	row := q.db.QueryRow(ctx, getProfileByUsername, username)
	var pr Profile
	err := row.Scan(&pr.ID, &pr.Username, &pr.DisplayName, &pr.Email, &pr.AvatarUrl, &pr.Role, &pr.Bio, &pr.Country, &pr.CountryCode, &pr.CreatedAt, &pr.UpdatedAt)
	return pr, err
}

type UpdateProfileParams struct {
	ID          string  `json:"id"`
	Username    *string `json:"username,omitempty"`
	DisplayName *string `json:"display_name,omitempty"`
	AvatarUrl   *string `json:"avatar_url,omitempty"`
	Bio         *string `json:"bio,omitempty"`
	Country     *string `json:"country,omitempty"`
	CountryCode *string `json:"country_code,omitempty"`
}

const updateProfile = `
UPDATE profiles SET
    username = COALESCE($2, username),
    display_name = COALESCE($3, display_name),
    avatar_url = COALESCE($4, avatar_url),
    bio = COALESCE($5, bio),
    country = COALESCE($6, country),
    country_code = COALESCE($7, country_code)
WHERE id = $1
RETURNING id, username, display_name, email, avatar_url, role, bio, country, country_code, created_at, updated_at`

func (q *Queries) UpdateProfile(ctx context.Context, p UpdateProfileParams) (Profile, error) {
	row := q.db.QueryRow(ctx, updateProfile, p.ID, p.Username, p.DisplayName, p.AvatarUrl, p.Bio, p.Country, p.CountryCode)
	var pr Profile
	err := row.Scan(&pr.ID, &pr.Username, &pr.DisplayName, &pr.Email, &pr.AvatarUrl, &pr.Role, &pr.Bio, &pr.Country, &pr.CountryCode, &pr.CreatedAt, &pr.UpdatedAt)
	return pr, err
}

const getProfileStats = `
SELECT
    (SELECT COUNT(*) FROM registrations WHERE user_id = $1 AND payment_status = 'paid') AS tournaments_played,
    (SELECT COUNT(*) FROM tournaments WHERE organizer_id = $1) AS tournaments_hosted,
    (SELECT COUNT(DISTINCT tournament_id) FROM matches WHERE winner_id = $1
        AND round = (SELECT MAX(round) FROM matches m2 WHERE m2.tournament_id = matches.tournament_id)
    ) AS wins`

func (q *Queries) GetProfileStats(ctx context.Context, userID string) (ProfileStats, error) {
	var s ProfileStats
	row := q.db.QueryRow(ctx, getProfileStats, userID)
	err := row.Scan(&s.TournamentsPlayed, &s.TournamentsHosted, &s.Wins)
	return s, err
}

type SearchProfilesParams struct {
	Query  string `json:"q"`
	Limit  int32  `json:"limit"`
	Offset int32  `json:"offset"`
}

const searchProfiles = `
SELECT id, username, display_name, avatar_url, role
FROM profiles
WHERE
    $1 = ''
    OR username ILIKE '%' || $1 || '%'
    OR COALESCE(display_name, '') ILIKE '%' || $1 || '%'
ORDER BY
    CASE WHEN username ILIKE $1 || '%' THEN 0 ELSE 1 END,
    username ASC
LIMIT $2 OFFSET $3`

func (q *Queries) SearchProfiles(ctx context.Context, p SearchProfilesParams) ([]ProfileSearchResult, error) {
	rows, err := q.db.Query(ctx, searchProfiles, p.Query, p.Limit, p.Offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []ProfileSearchResult
	for rows.Next() {
		var profile ProfileSearchResult
		if err := rows.Scan(&profile.ID, &profile.Username, &profile.DisplayName, &profile.AvatarUrl, &profile.Role); err != nil {
			return nil, err
		}
		out = append(out, profile)
	}
	return out, rows.Err()
}

const deleteProfile = `DELETE FROM profiles WHERE id = $1`

func (q *Queries) DeleteProfile(ctx context.Context, id string) error {
	_, err := q.db.Exec(ctx, deleteProfile, id)
	return err
}
