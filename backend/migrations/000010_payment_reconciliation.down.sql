DROP INDEX IF EXISTS idx_payments_provider_payment_id;
DROP INDEX IF EXISTS idx_payments_verified_at;

ALTER TABLE payments
    DROP COLUMN IF EXISTS failure_reason,
    DROP COLUMN IF EXISTS webhook_received_at,
    DROP COLUMN IF EXISTS verified_at,
    DROP COLUMN IF EXISTS provider_payment_id,
    DROP COLUMN IF EXISTS provider_status;
