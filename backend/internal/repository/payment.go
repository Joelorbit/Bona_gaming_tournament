package repository

import (
	"context"
	"encoding/json"
)

type CreatePaymentParams struct {
	UserID       string `json:"user_id"`
	TournamentID string `json:"tournament_id"`
	Amount       int32  `json:"amount"`
	Currency     string `json:"currency"`
	Status       string `json:"status"`
}

const paymentColumns = `id, user_id, tournament_id, amount, currency, status,
    addispay_ref, payment_url, provider_status, provider_payment_id, verified_at,
    webhook_received_at, failure_reason, refund_status, refund_reason,
    refund_requested_at, refunded_at, refunded_by, metadata, created_at, updated_at`

func scanPayment(row interface {
	Scan(...any) error
}) (Payment, error) {
	var p Payment
	err := row.Scan(
		&p.ID, &p.UserID, &p.TournamentID, &p.Amount, &p.Currency, &p.Status,
		&p.AddispayRef, &p.PaymentURL, &p.ProviderStatus, &p.ProviderPaymentID,
		&p.VerifiedAt, &p.WebhookReceivedAt, &p.FailureReason,
		&p.RefundStatus, &p.RefundReason, &p.RefundRequestedAt, &p.RefundedAt, &p.RefundedBy,
		&p.Metadata, &p.CreatedAt, &p.UpdatedAt,
	)
	return p, err
}

const createPayment = `
INSERT INTO payments (user_id, tournament_id, amount, currency, status)
VALUES ($1, $2, $3, $4, $5)
RETURNING ` + paymentColumns

func (q *Queries) CreatePayment(ctx context.Context, p CreatePaymentParams) (Payment, error) {
	return scanPayment(q.db.QueryRow(ctx, createPayment, p.UserID, p.TournamentID, p.Amount, p.Currency, p.Status))
}

const getPayment = `SELECT ` + paymentColumns + ` FROM payments WHERE id = $1`

func (q *Queries) GetPayment(ctx context.Context, id string) (Payment, error) {
	return scanPayment(q.db.QueryRow(ctx, getPayment, id))
}

const getPaymentByAddispayRef = `SELECT ` + paymentColumns + ` FROM payments WHERE addispay_ref = $1`

func (q *Queries) GetPaymentByAddispayRef(ctx context.Context, ref string) (Payment, error) {
	return scanPayment(q.db.QueryRow(ctx, getPaymentByAddispayRef, ref))
}

type UpdatePaymentStatusParams struct {
	ID     string `json:"id"`
	Status string `json:"status"`
}

const updatePaymentStatus = `
UPDATE payments SET status = $2, updated_at = NOW() WHERE id = $1
RETURNING ` + paymentColumns

func (q *Queries) UpdatePaymentStatus(ctx context.Context, p UpdatePaymentStatusParams) (Payment, error) {
	return scanPayment(q.db.QueryRow(ctx, updatePaymentStatus, p.ID, p.Status))
}

type UpdatePaymentAddispayRefParams struct {
	ID          string `json:"id"`
	AddispayRef string `json:"addispay_ref"`
}

const updatePaymentAddispayRef = `
UPDATE payments SET addispay_ref = $2, updated_at = NOW() WHERE id = $1
RETURNING ` + paymentColumns

func (q *Queries) UpdatePaymentAddispayRef(ctx context.Context, p UpdatePaymentAddispayRefParams) (Payment, error) {
	return scanPayment(q.db.QueryRow(ctx, updatePaymentAddispayRef, p.ID, p.AddispayRef))
}

type UpdatePaymentGatewayFieldsParams struct {
	ID          string `json:"id"`
	AddispayRef string `json:"addispay_ref"`
	PaymentURL  string `json:"payment_url"`
}

const updatePaymentGatewayFields = `
UPDATE payments SET addispay_ref = $2, payment_url = $3, updated_at = NOW() WHERE id = $1
RETURNING ` + paymentColumns

func (q *Queries) UpdatePaymentGatewayFields(ctx context.Context, p UpdatePaymentGatewayFieldsParams) (Payment, error) {
	return scanPayment(q.db.QueryRow(ctx, updatePaymentGatewayFields, p.ID, p.AddispayRef, p.PaymentURL))
}

type CompletePaymentFromWebhookParams struct {
	ID                string          `json:"id"`
	Status            string          `json:"status"`
	ProviderStatus    string          `json:"provider_status"`
	ProviderPaymentID *string         `json:"provider_payment_id,omitempty"`
	FailureReason     *string         `json:"failure_reason,omitempty"`
	Metadata          json.RawMessage `json:"metadata,omitempty"`
}

const completePaymentFromWebhook = `
UPDATE payments
SET status = CASE
        WHEN status = 'paid' THEN status
        ELSE $2
    END,
    provider_status = $3,
    provider_payment_id = COALESCE($4, provider_payment_id),
    verified_at = CASE
        WHEN $2 = 'paid' THEN COALESCE(verified_at, NOW())
        ELSE verified_at
    END,
    webhook_received_at = NOW(),
    failure_reason = $5,
    metadata = COALESCE($6, metadata),
    updated_at = NOW()
WHERE id = $1
RETURNING ` + paymentColumns

func (q *Queries) CompletePaymentFromWebhook(ctx context.Context, p CompletePaymentFromWebhookParams) (Payment, error) {
	return scanPayment(q.db.QueryRow(
		ctx,
		completePaymentFromWebhook,
		p.ID,
		p.Status,
		p.ProviderStatus,
		p.ProviderPaymentID,
		p.FailureReason,
		p.Metadata,
	))
}

type MarkPaymentGatewayFailedParams struct {
	ID            string  `json:"id"`
	FailureReason *string `json:"failure_reason,omitempty"`
}

const markPaymentGatewayFailed = `
UPDATE payments
SET status = 'failed',
    failure_reason = $2,
    updated_at = NOW()
WHERE id = $1 AND status != 'paid'
RETURNING ` + paymentColumns

func (q *Queries) MarkPaymentGatewayFailed(ctx context.Context, p MarkPaymentGatewayFailedParams) (Payment, error) {
	return scanPayment(q.db.QueryRow(ctx, markPaymentGatewayFailed, p.ID, p.FailureReason))
}

const listAllPayments = `SELECT ` + paymentColumns + ` FROM payments ORDER BY created_at DESC LIMIT $1 OFFSET $2`

func (q *Queries) ListAllPayments(ctx context.Context, limit, offset int32) ([]Payment, error) {
	rows, err := q.db.Query(ctx, listAllPayments, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []Payment
	for rows.Next() {
		p, err := scanPayment(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, p)
	}
	return out, rows.Err()
}

const listPaidPaymentsByTournament = `
SELECT ` + paymentColumns + `
FROM payments
WHERE tournament_id = $1 AND status = 'paid'
ORDER BY created_at DESC`

func (q *Queries) ListPaidPaymentsByTournament(ctx context.Context, tournamentID string) ([]Payment, error) {
	rows, err := q.db.Query(ctx, listPaidPaymentsByTournament, tournamentID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []Payment
	for rows.Next() {
		p, err := scanPayment(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, p)
	}
	return out, rows.Err()
}

const listPaymentsByOrganizer = `
SELECT p.id, p.user_id, p.tournament_id, p.amount, p.currency, p.status,
       p.addispay_ref, p.payment_url, p.provider_status, p.provider_payment_id, p.verified_at,
       p.webhook_received_at, p.failure_reason, p.refund_status, p.refund_reason,
       p.refund_requested_at, p.refunded_at, p.refunded_by, p.metadata, p.created_at, p.updated_at
FROM payments p
JOIN tournaments t ON t.id = p.tournament_id
WHERE t.organizer_id = $1
ORDER BY p.created_at DESC`

func (q *Queries) ListPaymentsByOrganizer(ctx context.Context, organizerID string) ([]Payment, error) {
	rows, err := q.db.Query(ctx, listPaymentsByOrganizer, organizerID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []Payment
	for rows.Next() {
		p, err := scanPayment(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, p)
	}
	return out, rows.Err()
}

type MarkTournamentPaymentsRefundPendingParams struct {
	TournamentID string `json:"tournament_id"`
	Reason       string `json:"reason"`
}

const markTournamentPaymentsRefundPending = `
UPDATE payments
SET refund_status = CASE
        WHEN refund_status = 'refunded' THEN refund_status
        ELSE 'pending'
    END,
    refund_reason = COALESCE(NULLIF($2, ''), refund_reason),
    refund_requested_at = COALESCE(refund_requested_at, NOW()),
    updated_at = NOW()
WHERE tournament_id = $1
  AND status = 'paid'
  AND refund_status != 'refunded'
RETURNING ` + paymentColumns

func (q *Queries) MarkTournamentPaymentsRefundPending(ctx context.Context, p MarkTournamentPaymentsRefundPendingParams) ([]Payment, error) {
	rows, err := q.db.Query(ctx, markTournamentPaymentsRefundPending, p.TournamentID, p.Reason)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []Payment
	for rows.Next() {
		payment, err := scanPayment(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, payment)
	}
	return out, rows.Err()
}

type MarkPaymentRefundedParams struct {
	ID         string `json:"id"`
	RefundedBy string `json:"refunded_by"`
}

const markPaymentRefunded = `
UPDATE payments
SET status = 'refunded',
    refund_status = 'refunded',
    refunded_at = NOW(),
    refunded_by = $2,
    updated_at = NOW()
WHERE id = $1
  AND status = 'paid'
  AND refund_status = 'pending'
RETURNING ` + paymentColumns

func (q *Queries) MarkPaymentRefunded(ctx context.Context, p MarkPaymentRefundedParams) (Payment, error) {
	return scanPayment(q.db.QueryRow(ctx, markPaymentRefunded, p.ID, p.RefundedBy))
}
