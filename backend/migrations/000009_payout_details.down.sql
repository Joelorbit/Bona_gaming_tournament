ALTER TABLE payouts
    DROP COLUMN IF EXISTS payout_details_submitted_at,
    DROP COLUMN IF EXISTS bank_account_number,
    DROP COLUMN IF EXISTS bank_account_name,
    DROP COLUMN IF EXISTS bank_name,
    DROP COLUMN IF EXISTS telebirr_number,
    DROP COLUMN IF EXISTS phone_number,
    DROP COLUMN IF EXISTS payout_method;
