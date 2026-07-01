-- name: CreatePayment :one
INSERT INTO payments (user_id, tournament_id, amount, currency, status)
VALUES ($1, $2, $3, $4, $5)
RETURNING id, user_id, tournament_id, amount, currency, status, addispay_ref, payment_url,
    provider_status, provider_payment_id, verified_at, webhook_received_at, failure_reason,
    refund_status, refund_reason, refund_requested_at, refunded_at, refunded_by,
    metadata, created_at, updated_at;

-- name: GetPayment :one
SELECT id, user_id, tournament_id, amount, currency, status, addispay_ref, payment_url,
    provider_status, provider_payment_id, verified_at, webhook_received_at, failure_reason,
    refund_status, refund_reason, refund_requested_at, refunded_at, refunded_by,
    metadata, created_at, updated_at
FROM payments
WHERE id = $1;

-- name: GetPaymentByAddispayRef :one
SELECT id, user_id, tournament_id, amount, currency, status, addispay_ref, payment_url,
    provider_status, provider_payment_id, verified_at, webhook_received_at, failure_reason,
    refund_status, refund_reason, refund_requested_at, refunded_at, refunded_by,
    metadata, created_at, updated_at
FROM payments
WHERE addispay_ref = $1;

-- name: UpdatePaymentStatus :one
UPDATE payments
SET status = $2, updated_at = NOW()
WHERE id = $1
RETURNING id, user_id, tournament_id, amount, currency, status, addispay_ref, payment_url,
    provider_status, provider_payment_id, verified_at, webhook_received_at, failure_reason,
    refund_status, refund_reason, refund_requested_at, refunded_at, refunded_by,
    metadata, created_at, updated_at;

-- name: UpdatePaymentAddispayRef :one
UPDATE payments
SET addispay_ref = $2, updated_at = NOW()
WHERE id = $1
RETURNING id, user_id, tournament_id, amount, currency, status, addispay_ref, payment_url,
    provider_status, provider_payment_id, verified_at, webhook_received_at, failure_reason,
    refund_status, refund_reason, refund_requested_at, refunded_at, refunded_by,
    metadata, created_at, updated_at;

-- name: UpdatePaymentGatewayFields :one
UPDATE payments
SET addispay_ref = $2, payment_url = $3, updated_at = NOW()
WHERE id = $1
RETURNING id, user_id, tournament_id, amount, currency, status, addispay_ref, payment_url,
    provider_status, provider_payment_id, verified_at, webhook_received_at, failure_reason,
    refund_status, refund_reason, refund_requested_at, refunded_at, refunded_by,
    metadata, created_at, updated_at;
