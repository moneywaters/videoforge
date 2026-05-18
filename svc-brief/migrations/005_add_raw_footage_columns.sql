-- Add raw footage columns to briefs table
ALTER TABLE brief.briefs
ADD COLUMN IF NOT EXISTS raw_footage_storj_key VARCHAR(512),
ADD COLUMN IF NOT EXISTS has_raw_footage BOOLEAN DEFAULT FALSE;

-- Index for raw footage query
CREATE INDEX IF NOT EXISTS idx_briefs_has_raw_footage ON brief.briefs(has_raw_footage) WHERE has_raw_footage = TRUE;

-- Update permissions
GRANT SELECT, INSERT, UPDATE ON ALL TABLES IN SCHEMA brief TO videoforge;