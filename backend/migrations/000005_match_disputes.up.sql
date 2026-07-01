ALTER TABLE matches ADD COLUMN evidence_screenshot_url TEXT;
ALTER TABLE matches ADD COLUMN evidence_video_url TEXT;
ALTER TABLE matches ADD COLUMN evidence_notes TEXT;
ALTER TABLE matches ADD COLUMN result_submitted_by UUID REFERENCES profiles(id) ON DELETE SET NULL;
ALTER TABLE matches ADD COLUMN result_confirmed_at TIMESTAMPTZ;
ALTER TABLE matches ADD COLUMN dispute_status VARCHAR(20) NOT NULL DEFAULT 'none'
    CHECK (dispute_status IN ('none', 'pending', 'resolved'));
ALTER TABLE matches ADD COLUMN dispute_reason TEXT;
ALTER TABLE matches ADD COLUMN dispute_opened_by UUID REFERENCES profiles(id) ON DELETE SET NULL;
ALTER TABLE matches ADD COLUMN dispute_opened_at TIMESTAMPTZ;
ALTER TABLE matches ADD COLUMN dispute_resolved_at TIMESTAMPTZ;

ALTER TABLE matches DROP CONSTRAINT IF EXISTS matches_status_check;
ALTER TABLE matches ADD CONSTRAINT matches_status_check CHECK (status IN ('pending', 'in_progress', 'awaiting_confirmation', 'disputed', 'completed', 'walkover'));
