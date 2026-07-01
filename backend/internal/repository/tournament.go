package repository

import (
	"context"
	"time"
)

type CreateTournamentParams struct {
	Title                string     `json:"title"`
	Game                 string     `json:"game"`
	Description          *string    `json:"description,omitempty"`
	Rules                *string    `json:"rules,omitempty"`
	Format               string     `json:"format"`
	MaxParticipants      int32      `json:"max_participants"`
	MinParticipants      int32      `json:"min_participants"`
	TeamSize             int32      `json:"team_size"`
	EntryFee             int32      `json:"entry_fee"`
	PrizePool            int32      `json:"prize_pool"`
	Currency             string     `json:"currency"`
	Location             string     `json:"location"`
	BestOf               int32      `json:"best_of"`
	PlatformFeePct       int32      `json:"platform_fee_pct"`
	OrganizerFeePct      int32      `json:"organizer_fee_pct"`
	StartDate            time.Time  `json:"start_date"`
	EndDate              *time.Time `json:"end_date,omitempty"`
	RegistrationDeadline *time.Time `json:"registration_deadline,omitempty"`
	RegistrationCloseAt  *time.Time `json:"registration_close_at,omitempty"`
	OrganizerID          string     `json:"organizer_id"`
	BannerUrl            *string    `json:"banner_url,omitempty"`
}

const tournamentColumns = `id, title, game, description, rules, format, status,
    max_participants, min_participants, team_size, entry_fee, prize_pool, currency,
    location, best_of, platform_fee_pct, organizer_fee_pct,
    start_date, end_date, registration_deadline, registration_close_at,
    organizer_id, banner_url, created_at, updated_at`

func scanTournament(row interface {
	Scan(...any) error
}) (Tournament, error) {
	var t Tournament
	err := row.Scan(
		&t.ID, &t.Title, &t.Game, &t.Description, &t.Rules, &t.Format, &t.Status,
		&t.MaxParticipants, &t.MinParticipants, &t.TeamSize, &t.EntryFee, &t.PrizePool, &t.Currency,
		&t.Location, &t.BestOf, &t.PlatformFeePct, &t.OrganizerFeePct,
		&t.StartDate, &t.EndDate, &t.RegistrationDeadline, &t.RegistrationCloseAt,
		&t.OrganizerID, &t.BannerUrl, &t.CreatedAt, &t.UpdatedAt,
	)
	return t, err
}

const createTournament = `
INSERT INTO tournaments (
    title, game, description, rules, format, max_participants, min_participants,
    team_size, entry_fee, prize_pool, currency, location, best_of, platform_fee_pct,
    organizer_fee_pct, start_date, end_date, registration_deadline,
    registration_close_at, organizer_id, banner_url
) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19, $20, $21)
RETURNING ` + tournamentColumns

func (q *Queries) CreateTournament(ctx context.Context, p CreateTournamentParams) (Tournament, error) {
	format := p.Format
	if format == "" {
		format = "single_elimination"
	}
	currency := p.Currency
	if currency == "" {
		currency = "ETB"
	}
	location := p.Location
	if location == "" {
		location = "Online"
	}
	minP := p.MinParticipants
	if minP < 2 {
		minP = 2
	}
	teamSize := p.TeamSize
	if teamSize == 0 {
		teamSize = 1
	}
	bestOf := p.BestOf
	if bestOf == 0 {
		bestOf = 1
	}
	platformFee := p.PlatformFeePct
	if platformFee == 0 {
		platformFee = 5
	}

	row := q.db.QueryRow(ctx, createTournament,
		p.Title, p.Game, p.Description, p.Rules, format, p.MaxParticipants, minP,
		teamSize, p.EntryFee, p.PrizePool, currency, location, bestOf, platformFee,
		p.OrganizerFeePct, p.StartDate, p.EndDate, p.RegistrationDeadline,
		p.RegistrationCloseAt, p.OrganizerID, p.BannerUrl,
	)
	return scanTournament(row)
}

const getTournament = `SELECT ` + tournamentColumns + ` FROM tournaments WHERE id = $1`

func (q *Queries) GetTournament(ctx context.Context, id string) (Tournament, error) {
	return scanTournament(q.db.QueryRow(ctx, getTournament, id))
}

type ListTournamentsParams struct {
	Status  string `json:"status"`
	Game    string `json:"game"`
	Query   string `json:"q"`
	Paid    string `json:"paid"`  // "", "free", "paid"
	Sort    string `json:"sort"`  // "newest", "starting_soon", "prize_high", "prize_low"
	Limit   int32  `json:"limit"`
	Offset  int32  `json:"offset"`
}

func (q *Queries) ListTournaments(ctx context.Context, p ListTournamentsParams) ([]Tournament, error) {
	orderBy := "created_at DESC"
	switch p.Sort {
	case "starting_soon":
		orderBy = "start_date ASC"
	case "prize_high":
		orderBy = "prize_pool DESC"
	case "prize_low":
		orderBy = "prize_pool ASC"
	case "newest", "":
		orderBy = "created_at DESC"
	}

	paidFilter := ""
	switch p.Paid {
	case "free":
		paidFilter = "AND entry_fee = 0"
	case "paid":
		paidFilter = "AND entry_fee > 0"
	}

	sql := `SELECT ` + tournamentColumns + `
FROM tournaments
WHERE ($1 = '' OR status = $1)
  AND ($2 = '' OR game = $2)
  AND ($3 = '' OR title ILIKE '%' || $3 || '%' OR game ILIKE '%' || $3 || '%')
  ` + paidFilter + `
ORDER BY ` + orderBy + `
LIMIT $4 OFFSET $5`

	rows, err := q.db.Query(ctx, sql, p.Status, p.Game, p.Query, p.Limit, p.Offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []Tournament
	for rows.Next() {
		t, err := scanTournament(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, t)
	}
	return out, rows.Err()
}

const listTournamentsByOrganizer = `
SELECT ` + tournamentColumns + `
FROM tournaments WHERE organizer_id = $1 ORDER BY created_at DESC`

func (q *Queries) ListTournamentsByOrganizer(ctx context.Context, organizerID string) ([]Tournament, error) {
	rows, err := q.db.Query(ctx, listTournamentsByOrganizer, organizerID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []Tournament
	for rows.Next() {
		t, err := scanTournament(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, t)
	}
	return out, rows.Err()
}

type UpdateTournamentParams struct {
	ID                   string     `json:"id"`
	Title                *string    `json:"title,omitempty"`
	Game                 *string    `json:"game,omitempty"`
	Description          *string    `json:"description,omitempty"`
	Rules                *string    `json:"rules,omitempty"`
	Format               *string    `json:"format,omitempty"`
	MaxParticipants      *int32     `json:"max_participants,omitempty"`
	MinParticipants      *int32     `json:"min_participants,omitempty"`
	TeamSize             *int32     `json:"team_size,omitempty"`
	EntryFee             *int32     `json:"entry_fee,omitempty"`
	PrizePool            *int32     `json:"prize_pool,omitempty"`
	Currency             *string    `json:"currency,omitempty"`
	Location             *string    `json:"location,omitempty"`
	BestOf               *int32     `json:"best_of,omitempty"`
	OrganizerFeePct      *int32     `json:"organizer_fee_pct,omitempty"`
	StartDate            *time.Time `json:"start_date,omitempty"`
	EndDate              *time.Time `json:"end_date,omitempty"`
	RegistrationDeadline *time.Time `json:"registration_deadline,omitempty"`
	RegistrationCloseAt  *time.Time `json:"registration_close_at,omitempty"`
	BannerUrl            *string    `json:"banner_url,omitempty"`
}

const updateTournament = `
UPDATE tournaments SET
    title = COALESCE($2, title),
    game = COALESCE($3, game),
    description = COALESCE($4, description),
    rules = COALESCE($5, rules),
    format = COALESCE($6, format),
    max_participants = COALESCE($7, max_participants),
    min_participants = COALESCE($8, min_participants),
    team_size = COALESCE($9, team_size),
    entry_fee = COALESCE($10, entry_fee),
    prize_pool = COALESCE($11, prize_pool),
    currency = COALESCE($12, currency),
    location = COALESCE($13, location),
    best_of = COALESCE($14, best_of),
    organizer_fee_pct = COALESCE($15, organizer_fee_pct),
    start_date = COALESCE($16, start_date),
    end_date = COALESCE($17, end_date),
    registration_deadline = COALESCE($18, registration_deadline),
    registration_close_at = COALESCE($19, registration_close_at),
    banner_url = COALESCE($20, banner_url)
WHERE id = $1
RETURNING ` + tournamentColumns

func (q *Queries) UpdateTournament(ctx context.Context, p UpdateTournamentParams) (Tournament, error) {
	row := q.db.QueryRow(ctx, updateTournament,
		p.ID, p.Title, p.Game, p.Description, p.Rules, p.Format,
		p.MaxParticipants, p.MinParticipants, p.TeamSize, p.EntryFee, p.PrizePool,
		p.Currency, p.Location, p.BestOf, p.OrganizerFeePct, p.StartDate,
		p.EndDate, p.RegistrationDeadline, p.RegistrationCloseAt, p.BannerUrl,
	)
	return scanTournament(row)
}

type UpdateTournamentStatusParams struct {
	ID     string `json:"id"`
	Status string `json:"status"`
}

const updateTournamentStatus = `
UPDATE tournaments SET status = $2 WHERE id = $1
RETURNING ` + tournamentColumns

func (q *Queries) UpdateTournamentStatus(ctx context.Context, p UpdateTournamentStatusParams) (Tournament, error) {
	return scanTournament(q.db.QueryRow(ctx, updateTournamentStatus, p.ID, p.Status))
}

const deleteTournament = `DELETE FROM tournaments WHERE id = $1`

func (q *Queries) DeleteTournament(ctx context.Context, id string) error {
	_, err := q.db.Exec(ctx, deleteTournament, id)
	return err
}
