-- Initial schema for Bona Tournament

CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

CREATE OR REPLACE FUNCTION set_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TABLE profiles (
    id UUID PRIMARY KEY,
    username VARCHAR(50) UNIQUE NOT NULL,
    display_name VARCHAR(100),
    email VARCHAR(255),
    avatar_url TEXT,
    role VARCHAR(20) NOT NULL DEFAULT 'player' CHECK (role IN ('player', 'organizer', 'admin')),
    bio TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TRIGGER trg_profiles_updated_at BEFORE UPDATE ON profiles
    FOR EACH ROW EXECUTE PROCEDURE set_updated_at();

CREATE TABLE tournaments (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    title VARCHAR(200) NOT NULL,
    game VARCHAR(100) NOT NULL,
    description TEXT,
    rules TEXT,
    format VARCHAR(30) NOT NULL DEFAULT 'single_elimination'
        CHECK (format IN ('single_elimination', 'double_elimination', 'round_robin', 'swiss')),
    status VARCHAR(30) NOT NULL DEFAULT 'draft'
        CHECK (status IN ('draft', 'open', 'registration_closed', 'in_progress', 'completed', 'cancelled')),
    max_participants INTEGER NOT NULL CHECK (max_participants >= 2),
    min_participants INTEGER NOT NULL DEFAULT 2 CHECK (min_participants >= 2),
    team_size INTEGER NOT NULL DEFAULT 1 CHECK (team_size IN (1, 2, 3, 5)),
    entry_fee INTEGER NOT NULL DEFAULT 0 CHECK (entry_fee >= 0),
    prize_pool INTEGER NOT NULL DEFAULT 0 CHECK (prize_pool >= 0),
    currency VARCHAR(3) NOT NULL DEFAULT 'ETB',
    location VARCHAR(100) NOT NULL DEFAULT 'Online',
    start_date TIMESTAMPTZ NOT NULL,
    end_date TIMESTAMPTZ,
    registration_deadline TIMESTAMPTZ,
    organizer_id UUID NOT NULL REFERENCES profiles(id) ON DELETE CASCADE,
    banner_url TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TRIGGER trg_tournaments_updated_at BEFORE UPDATE ON tournaments
    FOR EACH ROW EXECUTE PROCEDURE set_updated_at();

CREATE TABLE registrations (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    tournament_id UUID NOT NULL REFERENCES tournaments(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES profiles(id) ON DELETE CASCADE,
    payment_status VARCHAR(20) NOT NULL DEFAULT 'pending'
        CHECK (payment_status IN ('pending', 'paid', 'failed', 'refunded')),
    seed INTEGER,
    registered_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (tournament_id, user_id)
);

CREATE TABLE matches (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    tournament_id UUID NOT NULL REFERENCES tournaments(id) ON DELETE CASCADE,
    round INTEGER NOT NULL CHECK (round > 0),
    position INTEGER NOT NULL CHECK (position >= 0),
    player_a_id UUID REFERENCES profiles(id) ON DELETE SET NULL,
    player_b_id UUID REFERENCES profiles(id) ON DELETE SET NULL,
    winner_id UUID REFERENCES profiles(id) ON DELETE SET NULL,
    score VARCHAR(50),
    status VARCHAR(30) NOT NULL DEFAULT 'pending'
        CHECK (status IN ('pending', 'in_progress', 'completed', 'walkover')),
    scheduled_at TIMESTAMPTZ,
    completed_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (tournament_id, round, position)
);

CREATE TRIGGER trg_matches_updated_at BEFORE UPDATE ON matches
    FOR EACH ROW EXECUTE PROCEDURE set_updated_at();

CREATE TABLE payments (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL REFERENCES profiles(id) ON DELETE CASCADE,
    tournament_id UUID NOT NULL REFERENCES tournaments(id) ON DELETE CASCADE,
    amount INTEGER NOT NULL CHECK (amount >= 0),
    currency VARCHAR(3) NOT NULL DEFAULT 'ETB',
    status VARCHAR(20) NOT NULL DEFAULT 'pending'
        CHECK (status IN ('pending', 'paid', 'failed', 'refunded')),
    addispay_ref VARCHAR(100),
    payment_url TEXT,
    metadata JSONB,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TRIGGER trg_payments_updated_at BEFORE UPDATE ON payments
    FOR EACH ROW EXECUTE PROCEDURE set_updated_at();

CREATE INDEX idx_tournaments_status ON tournaments(status);
CREATE INDEX idx_tournaments_game ON tournaments(game);
CREATE INDEX idx_tournaments_organizer ON tournaments(organizer_id);
CREATE INDEX idx_tournaments_start_date ON tournaments(start_date);
CREATE INDEX idx_registrations_tournament ON registrations(tournament_id);
CREATE INDEX idx_registrations_user ON registrations(user_id);
CREATE INDEX idx_registrations_payment_status ON registrations(payment_status);
CREATE INDEX idx_matches_tournament ON matches(tournament_id);
CREATE INDEX idx_matches_round ON matches(tournament_id, round);
CREATE INDEX idx_matches_players ON matches(player_a_id, player_b_id);
CREATE INDEX idx_matches_status ON matches(status);
CREATE INDEX idx_payments_user ON payments(user_id);
CREATE INDEX idx_payments_tournament ON payments(tournament_id);
CREATE INDEX idx_payments_addispay_ref ON payments(addispay_ref);
CREATE INDEX idx_payments_status ON payments(status);
