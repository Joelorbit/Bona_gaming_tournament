ALTER TABLE payments
    ADD COLUMN refund_status VARCHAR(20) NOT NULL DEFAULT 'none'
        CHECK (refund_status IN ('none', 'pending', 'refunded', 'failed', 'not_required')),
    ADD COLUMN refund_reason TEXT,
    ADD COLUMN refund_requested_at TIMESTAMPTZ,
    ADD COLUMN refunded_at TIMESTAMPTZ,
    ADD COLUMN refunded_by UUID REFERENCES profiles(id) ON DELETE SET NULL;

CREATE INDEX idx_payments_refund_status ON payments(refund_status);
