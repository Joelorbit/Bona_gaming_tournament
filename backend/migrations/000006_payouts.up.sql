CREATE TABLE payouts (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    tournament_id UUID NOT NULL REFERENCES tournaments(id) ON DELETE CASCADE,
    winner_id UUID NOT NULL REFERENCES profiles(id) ON DELETE CASCADE,
    amount INTEGER NOT NULL CHECK (amount >= 0),
    currency VARCHAR(3) NOT NULL DEFAULT 'ETB',
    status VARCHAR(20) NOT NULL DEFAULT 'pending'
        CHECK (status IN ('pending', 'paid', 'cancelled')),
    paid_at TIMESTAMPTZ,
    paid_by UUID REFERENCES profiles(id) ON DELETE SET NULL,
    note TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (tournament_id, winner_id)
);

CREATE INDEX idx_payouts_tournament ON payouts(tournament_id);
CREATE INDEX idx_payouts_winner ON payouts(winner_id);
CREATE INDEX idx_payouts_status ON payouts(status);

CREATE TRIGGER trg_payouts_updated_at BEFORE UPDATE ON payouts
    FOR EACH ROW EXECUTE PROCEDURE set_updated_at();
