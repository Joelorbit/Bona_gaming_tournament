DROP TRIGGER IF EXISTS trg_payments_updated_at ON payments;
DROP TRIGGER IF EXISTS trg_matches_updated_at ON matches;
DROP TRIGGER IF EXISTS trg_tournaments_updated_at ON tournaments;
DROP TRIGGER IF EXISTS trg_profiles_updated_at ON profiles;

DROP TABLE IF EXISTS payments CASCADE;
DROP TABLE IF EXISTS matches CASCADE;
DROP TABLE IF EXISTS registrations CASCADE;
DROP TABLE IF EXISTS tournaments CASCADE;
DROP TABLE IF EXISTS profiles CASCADE;

DROP FUNCTION IF EXISTS set_updated_at();
DROP EXTENSION IF EXISTS "uuid-ossp";
