-- Brief editor views table for tracking which editors have seen briefs
CREATE TABLE IF NOT EXISTS brief.brief_editor_views (
    brief_id UUID NOT NULL REFERENCES brief.briefs(id) ON DELETE CASCADE,
    editor_id UUID NOT NULL,
    viewed_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    PRIMARY KEY (brief_id, editor_id)
);

-- Indexes
CREATE INDEX IF NOT EXISTS idx_brief_editor_views_editor ON brief.brief_editor_views(editor_id);
CREATE INDEX IF NOT EXISTS idx_brief_editor_views_viewed ON brief.brief_editor_views(viewed_at DESC);

-- Grant permissions
GRANT SELECT, INSERT ON ALL TABLES IN SCHEMA brief TO videoforge;