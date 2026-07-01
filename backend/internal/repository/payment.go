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
    webhook_received_at, failure_reason, metadata, created_at, updated_at`

func scanPayment(row interface {
	Scan(...any) error
}) (Payment, error) {
	var p Payment
	err := row.Scan(
		&p.ID, &p.UserID, &p.TournamentID, &p.Amount, &p.Currency, &p.Status,
		&p.AddispayRef, &p.PaymentURL, &p.ProviderStatus, &p.ProviderPaymentID,
		&p.VerifiedAt, &p.WebhookReceivedAt, &p.FailureReason, &p.Metadata,
		&p.CreatedAt, &p.UpdatedAt,
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
