package repository

import "context"

const payoutColumns = `id, tournament_id, winner_id, amount, currency, status,
    payout_method, phone_number, telebirr_number, bank_name, bank_account_name,
    bank_account_number, payout_details_submitted_at,
    paid_at, paid_by, note, created_at, updated_at`

func scanPayout(row interface {
	Scan(...any) error
}) (Payout, error) {
	var p Payout
	err := row.Scan(
		&p.ID, &p.TournamentID, &p.WinnerID, &p.Amount, &p.Currency, &p.Status,
		&p.PayoutMethod, &p.PhoneNumber, &p.TelebirrNumber, &p.BankName, &p.BankAccountName,
		&p.BankAccountNumber, &p.PayoutDetailsSubmittedAt,
		&p.PaidAt, &p.PaidBy, &p.Note, &p.CreatedAt, &p.UpdatedAt,
	)
	return p, err
}

type CreatePayoutParams struct {
	TournamentID string  `json:"tournament_id"`
	WinnerID     string  `json:"winner_id"`
	Amount       int32   `json:"amount"`
	Currency     string  `json:"currency"`
	Status       string  `json:"status"`
	Note         *string `json:"note,omitempty"`
}

const createPayout = `
INSERT INTO payouts (tournament_id, winner_id, amount, currency, status, note)
VALUES ($1, $2, $3, $4, $5, $6)
ON CONFLICT (tournament_id, winner_id) DO UPDATE SET amount = EXCLUDED.amount, updated_at = NOW()
RETURNING ` + payoutColumns

func (q *Queries) CreatePayout(ctx context.Context, p CreatePayoutParams) (Payout, error) {
	row := q.db.QueryRow(ctx, createPayout, p.TournamentID, p.WinnerID, p.Amount, p.Currency, p.Status, p.Note)
	return scanPayout(row)
}

const getPayout = `SELECT ` + payoutColumns + ` FROM payouts WHERE id = $1`

func (q *Queries) GetPayout(ctx context.Context, id string) (Payout, error) {
	return scanPayout(q.db.QueryRow(ctx, getPayout, id))
}

const listPayoutsByTournament = `SELECT ` + payoutColumns + ` FROM payouts WHERE tournament_id = $1 ORDER BY created_at DESC`

func (q *Queries) ListPayoutsByTournament(ctx context.Context, tournamentID string) ([]Payout, error) {
	rows, err := q.db.Query(ctx, listPayoutsByTournament, tournamentID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []Payout
	for rows.Next() {
		p, err := scanPayout(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, p)
	}
	return out, rows.Err()
}

const listPayoutsByWinner = `SELECT ` + payoutColumns + ` FROM payouts WHERE winner_id = $1 ORDER BY created_at DESC`

func (q *Queries) ListPayoutsByWinner(ctx context.Context, winnerID string) ([]Payout, error) {
	rows, err := q.db.Query(ctx, listPayoutsByWinner, winnerID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []Payout
	for rows.Next() {
		p, err := scanPayout(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, p)
	}
	return out, rows.Err()
}

const listPayoutsByOrganizer = `
SELECT p.id, p.tournament_id, p.winner_id, p.amount, p.currency, p.status,
       p.payout_method, p.phone_number, p.telebirr_number, p.bank_name, p.bank_account_name,
       p.bank_account_number, p.payout_details_submitted_at,
       p.paid_at, p.paid_by, p.note, p.created_at, p.updated_at
FROM payouts p
JOIN tournaments t ON t.id = p.tournament_id
WHERE t.organizer_id = $1
ORDER BY p.created_at DESC`

func (q *Queries) ListPayoutsByOrganizer(ctx context.Context, organizerID string) ([]Payout, error) {
	rows, err := q.db.Query(ctx, listPayoutsByOrganizer, organizerID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []Payout
	for rows.Next() {
		p, err := scanPayout(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, p)
	}
	return out, rows.Err()
}

const listAllPayouts = `SELECT ` + payoutColumns + ` FROM payouts ORDER BY created_at DESC LIMIT $1 OFFSET $2`

func (q *Queries) ListAllPayouts(ctx context.Context, limit, offset int32) ([]Payout, error) {
	rows, err := q.db.Query(ctx, listAllPayouts, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []Payout
	for rows.Next() {
		p, err := scanPayout(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, p)
	}
	return out, rows.Err()
}

type MarkPayoutPaidParams struct {
	ID     string  `json:"id"`
	PaidBy string  `json:"paid_by"`
	Note   *string `json:"note,omitempty"`
}

const markPayoutPaid = `
UPDATE payouts
SET status = 'paid', paid_at = NOW(), paid_by = $2, note = COALESCE($3, note), updated_at = NOW()
WHERE id = $1 AND status != 'paid'
RETURNING ` + payoutColumns

func (q *Queries) MarkPayoutPaid(ctx context.Context, p MarkPayoutPaidParams) (Payout, error) {
	return scanPayout(q.db.QueryRow(ctx, markPayoutPaid, p.ID, p.PaidBy, p.Note))
}

type UpdatePayoutDetailsParams struct {
	ID                string  `json:"id"`
	PayoutMethod      string  `json:"payout_method"`
	PhoneNumber       *string `json:"phone_number,omitempty"`
	TelebirrNumber    *string `json:"telebirr_number,omitempty"`
	BankName          *string `json:"bank_name,omitempty"`
	BankAccountName   *string `json:"bank_account_name,omitempty"`
	BankAccountNumber *string `json:"bank_account_number,omitempty"`
}

const updatePayoutDetails = `
UPDATE payouts
SET payout_method = $2,
    phone_number = $3,
    telebirr_number = $4,
    bank_name = $5,
    bank_account_name = $6,
    bank_account_number = $7,
    payout_details_submitted_at = NOW(),
    updated_at = NOW()
WHERE id = $1
RETURNING ` + payoutColumns

func (q *Queries) UpdatePayoutDetails(ctx context.Context, p UpdatePayoutDetailsParams) (Payout, error) {
	return scanPayout(q.db.QueryRow(
		ctx,
		updatePayoutDetails,
		p.ID,
		p.PayoutMethod,
		p.PhoneNumber,
		p.TelebirrNumber,
		p.BankName,
		p.BankAccountName,
		p.BankAccountNumber,
	))
}
