ALTER TABLE payouts
    DROP CONSTRAINT IF EXISTS payouts_payout_method_check,
    ADD COLUMN IF NOT EXISTS payout_instructions TEXT,
    ADD CONSTRAINT payouts_payout_method_check
        CHECK (payout_method IS NULL OR payout_method IN ('telebirr', 'bank', 'other'));
