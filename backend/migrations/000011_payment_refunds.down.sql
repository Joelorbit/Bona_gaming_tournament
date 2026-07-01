DROP INDEX IF EXISTS idx_payments_refund_status;

ALTER TABLE payments
    DROP COLUMN IF EXISTS refunded_by,
    DROP COLUMN IF EXISTS refunded_at,
    DROP COLUMN IF EXISTS refund_requested_at,
    DROP COLUMN IF EXISTS refund_reason,
    DROP COLUMN IF EXISTS refund_status;
