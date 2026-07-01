ALTER TABLE payments
    ADD COLUMN provider_status VARCHAR(50),
    ADD COLUMN provider_payment_id VARCHAR(120),
    ADD COLUMN verified_at TIMESTAMPTZ,
    ADD COLUMN webhook_received_at TIMESTAMPTZ,
    ADD COLUMN failure_reason TEXT;

CREATE INDEX idx_payments_verified_at ON payments(verified_at);
CREATE INDEX idx_payments_provider_payment_id ON payments(provider_payment_id);
