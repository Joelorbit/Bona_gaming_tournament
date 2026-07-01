package repository

import "context"

type CreateRegistrationParams struct {
	TournamentID  string `json:"tournament_id"`
	UserID        string `json:"user_id"`
	PaymentStatus string `json:"payment_status"`
}

const createRegistration = `
INSERT INTO registrations (tournament_id, user_id, payment_status)
VALUES ($1, $2, $3)
RETURNING id, tournament_id, user_id, payment_status, seed, registered_at`

func (q *Queries) CreateRegistration(ctx context.Context, p CreateRegistrationParams) (Registration, error) {
	row := q.db.QueryRow(ctx, createRegistration, p.TournamentID, p.UserID, p.PaymentStatus)
	var r Registration
	err := row.Scan(&r.ID, &r.TournamentID, &r.UserID, &r.PaymentStatus, &r.Seed, &r.RegisteredAt)
	return r, err
}

type GetRegistrationByUserAndTournamentParams struct {
	UserID       string `json:"user_id"`
	TournamentID string `json:"tournament_id"`
}

const getRegistrationByUserAndTournament = `
SELECT id, tournament_id, user_id, payment_status, seed, registered_at
FROM registrations WHERE user_id = $1 AND tournament_id = $2`

func (q *Queries) GetRegistrationByUserAndTournament(ctx context.Context, p GetRegistrationByUserAndTournamentParams) (Registration, error) {
	row := q.db.QueryRow(ctx, getRegistrationByUserAndTournament, p.UserID, p.TournamentID)
	var r Registration
	err := row.Scan(&r.ID, &r.TournamentID, &r.UserID, &r.PaymentStatus, &r.Seed, &r.RegisteredAt)
	return r, err
}

const listRegistrationsByTournament = `
SELECT r.id, r.tournament_id, r.user_id, r.payment_status, r.seed, r.registered_at,
       p.username, p.display_name, p.avatar_url
FROM registrations r
JOIN profiles p ON r.user_id = p.id
WHERE r.tournament_id = $1
ORDER BY r.registered_at`

func (q *Queries) ListRegistrationsByTournament(ctx context.Context, tournamentID string) ([]ListRegistrationsByTournamentRow, error) {
	rows, err := q.db.Query(ctx, listRegistrationsByTournament, tournamentID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []ListRegistrationsByTournamentRow
	for rows.Next() {
		var r ListRegistrationsByTournamentRow
		err := rows.Scan(
			&r.ID, &r.TournamentID, &r.UserID, &r.PaymentStatus, &r.Seed, &r.RegisteredAt,
			&r.Username, &r.DisplayName, &r.AvatarUrl,
		)
		if err != nil {
			return nil, err
		}
		out = append(out, r)
	}
	return out, rows.Err()
}

const listRegistrationsByUser = `
SELECT r.id, r.tournament_id, r.user_id, r.payment_status, r.seed, r.registered_at,
       t.title, t.game, t.status, t.start_date, t.prize_pool
FROM registrations r
JOIN tournaments t ON r.tournament_id = t.id
WHERE r.user_id = $1
ORDER BY r.registered_at DESC`

func (q *Queries) ListRegistrationsByUser(ctx context.Context, userID string) ([]ListRegistrationsByUserRow, error) {
	rows, err := q.db.Query(ctx, listRegistrationsByUser, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []ListRegistrationsByUserRow
	for rows.Next() {
		var r ListRegistrationsByUserRow
		err := rows.Scan(
			&r.ID, &r.TournamentID, &r.UserID, &r.PaymentStatus, &r.Seed, &r.RegisteredAt,
			&r.Title, &r.Game, &r.Status, &r.StartDate, &r.PrizePool,
		)
		if err != nil {
			return nil, err
		}
		out = append(out, r)
	}
	return out, rows.Err()
}

const countPaidRegistrations = `
SELECT COUNT(*) FROM registrations WHERE tournament_id = $1 AND payment_status = 'paid'`

func (q *Queries) CountPaidRegistrations(ctx context.Context, tournamentID string) (int64, error) {
	var n int64
	err := q.db.QueryRow(ctx, countPaidRegistrations, tournamentID).Scan(&n)
	return n, err
}

const countActiveRegistrations = `
SELECT COUNT(*) FROM registrations WHERE tournament_id = $1 AND payment_status IN ('pending', 'paid')`

func (q *Queries) CountActiveRegistrations(ctx context.Context, tournamentID string) (int64, error) {
	var n int64
	err := q.db.QueryRow(ctx, countActiveRegistrations, tournamentID).Scan(&n)
	return n, err
}

type DeleteRegistrationParams struct {
	UserID       string `json:"user_id"`
	TournamentID string `json:"tournament_id"`
}

const deleteRegistration = `DELETE FROM registrations WHERE user_id = $1 AND tournament_id = $2`

func (q *Queries) DeleteRegistration(ctx context.Context, p DeleteRegistrationParams) error {
	_, err := q.db.Exec(ctx, deleteRegistration, p.UserID, p.TournamentID)
	return err
}

type UpdateRegistrationPaymentStatusParams struct {
	UserID        string `json:"user_id"`
	TournamentID  string `json:"tournament_id"`
	PaymentStatus string `json:"payment_status"`
}

const updateRegistrationPaymentStatus = `
UPDATE registrations SET payment_status = $3
WHERE user_id = $1 AND tournament_id = $2
RETURNING id, tournament_id, user_id, payment_status, seed, registered_at`

func (q *Queries) UpdateRegistrationPaymentStatus(ctx context.Context, p UpdateRegistrationPaymentStatusParams) (Registration, error) {
	row := q.db.QueryRow(ctx, updateRegistrationPaymentStatus, p.UserID, p.TournamentID, p.PaymentStatus)
	var r Registration
	err := row.Scan(&r.ID, &r.TournamentID, &r.UserID, &r.PaymentStatus, &r.Seed, &r.RegisteredAt)
	return r, err
}

const markTournamentRegistrationsRefunded = `
UPDATE registrations
SET payment_status = 'refunded'
WHERE tournament_id = $1 AND payment_status = 'paid'`

func (q *Queries) MarkTournamentRegistrationsRefunded(ctx context.Context, tournamentID string) error {
	_, err := q.db.Exec(ctx, markTournamentRegistrationsRefunded, tournamentID)
	return err
}
