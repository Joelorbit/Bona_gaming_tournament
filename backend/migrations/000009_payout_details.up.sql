ALTER TABLE payouts
    ADD COLUMN payout_method VARCHAR(20)
        CHECK (payout_method IS NULL OR payout_method IN ('telebirr', 'bank')),
    ADD COLUMN phone_number VARCHAR(40),
    ADD COLUMN telebirr_number VARCHAR(40),
    ADD COLUMN bank_name VARCHAR(120),
    ADD COLUMN bank_account_name VARCHAR(120),
    ADD COLUMN bank_account_number VARCHAR(80),
    ADD COLUMN payout_details_submitted_at TIMESTAMPTZ;
