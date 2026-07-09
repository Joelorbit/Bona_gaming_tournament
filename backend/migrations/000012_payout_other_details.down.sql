ALTER TABLE payouts
    DROP CONSTRAINT IF EXISTS payouts_payout_method_check,
    DROP COLUMN IF EXISTS payout_instructions,
    ADD CONSTRAINT payouts_payout_method_check
        CHECK (payout_method IS NULL OR payout_method IN ('telebirr', 'bank'));
