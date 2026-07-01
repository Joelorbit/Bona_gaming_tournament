ALTER TABLE matches DROP COLUMN IF EXISTS evidence_screenshot_url;
ALTER TABLE matches DROP COLUMN IF EXISTS evidence_video_url;
ALTER TABLE matches DROP COLUMN IF EXISTS evidence_notes;
ALTER TABLE matches DROP COLUMN IF EXISTS result_submitted_by;
ALTER TABLE matches DROP COLUMN IF EXISTS result_confirmed_at;
ALTER TABLE matches DROP COLUMN IF EXISTS dispute_status;
ALTER TABLE matches DROP COLUMN IF EXISTS dispute_reason;
ALTER TABLE matches DROP COLUMN IF EXISTS dispute_opened_by;
ALTER TABLE matches DROP COLUMN IF EXISTS dispute_opened_at;
ALTER TABLE matches DROP COLUMN IF EXISTS dispute_resolved_at;

ALTER TABLE matches DROP CONSTRAINT IF EXISTS matches_status_check;
ALTER TABLE matches ADD CONSTRAINT matches_status_check CHECK (status IN ('pending', 'in_progress', 'completed', 'walkover'));
