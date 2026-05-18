-- Video Service Migrations
-- Migration 001: Create videos table
CREATE SCHEMA IF NOT EXISTS video;

-- Videos table
CREATE TABLE IF NOT EXISTS video.videos (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    brief_id UUID NOT NULL,
    editor_id UUID NOT NULL,
    title VARCHAR(512) NOT NULL,
    description TEXT,
    storj_key VARCHAR(1024),
    status VARCHAR(32) DEFAULT 'draft' CHECK (status IN ('draft', 'submitted', 'approved', 'rejected', 'revision_requested')),
    current_revision_id UUID,
    duration INTEGER DEFAULT 0,
    resolution VARCHAR(16) DEFAULT '1080p',
    thumbnail_url TEXT,
    submitted_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Video revisions table
CREATE TABLE IF NOT EXISTS video.video_revisions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    video_id UUID NOT NULL REFERENCES video.videos(id) ON DELETE CASCADE,
    revision_number INTEGER NOT NULL,
    storj_key VARCHAR(1024),
    changelog TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    UNIQUE(video_id, revision_number)
);

-- Video approvals table
CREATE TABLE IF NOT EXISTS video.video_approvals (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    video_id UUID NOT NULL REFERENCES video.videos(id) ON DELETE CASCADE,
    status VARCHAR(32) NOT NULL CHECK (status IN ('approved', 'rejected')),
    approved_by UUID NOT NULL,
    approved_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    notes TEXT
);

-- Video feedback table
CREATE TABLE IF NOT EXISTS video.video_feedback (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    video_id UUID NOT NULL REFERENCES video.videos(id) ON DELETE CASCADE,
    revision_id UUID NOT NULL,
    feedback TEXT NOT NULL,
    created_by UUID NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Indexes
CREATE INDEX IF NOT EXISTS idx_videos_brief ON video.videos(brief_id);
CREATE INDEX IF NOT EXISTS idx_videos_editor ON video.videos(editor_id);
CREATE INDEX IF NOT EXISTS idx_videos_status ON video.videos(status);
CREATE INDEX IF NOT EXISTS idx_video_revisions_video ON video.video_revisions(video_id);
CREATE INDEX IF NOT EXISTS idx_video_approvals_video ON video.video_approvals(video_id);
CREATE INDEX IF NOT EXISTS idx_video_feedback_video ON video.video_feedback(video_id);

-- Grant permissions
GRANT USAGE ON SCHEMA video TO videoforge;
GRANT SELECT, INSERT, UPDATE, DELETE ON ALL TABLES IN SCHEMA video TO videoforge;
GRANT USAGE, SELECT ON ALL SEQUENCES IN SCHEMA video TO videoforge;