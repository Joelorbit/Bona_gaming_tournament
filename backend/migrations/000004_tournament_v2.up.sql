ALTER TABLE tournaments ADD COLUMN best_of INTEGER NOT NULL DEFAULT 1 CHECK (best_of >= 1 AND best_of % 2 = 1);
ALTER TABLE tournaments ADD COLUMN platform_fee_pct INTEGER NOT NULL DEFAULT 5 CHECK (platform_fee_pct >= 0 AND platform_fee_pct <= 100);
ALTER TABLE tournaments ADD COLUMN organizer_fee_pct INTEGER NOT NULL DEFAULT 0 CHECK (organizer_fee_pct >= 0 AND organizer_fee_pct <= 20);
ALTER TABLE tournaments ADD COLUMN registration_close_at TIMESTAMPTZ;
