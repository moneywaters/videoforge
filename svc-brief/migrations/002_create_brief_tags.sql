-- Brief tags table
CREATE TABLE IF NOT EXISTS brief.brief_tags (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    brief_id UUID NOT NULL REFERENCES brief.briefs(id) ON DELETE CASCADE,
    tag VARCHAR(128) NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    UNIQUE(brief_id, tag)
);

-- Indexes
CREATE INDEX IF NOT EXISTS idx_brief_tags_brief ON brief.brief_tags(brief_id);
CREATE INDEX IF NOT EXISTS idx_brief_tags_tag ON brief.brief_tags(tag);

-- Grant permissions
GRANT SELECT, INSERT, UPDATE, DELETE ON ALL TABLES IN SCHEMA brief TO videoforge;