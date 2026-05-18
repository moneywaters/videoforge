-- Migration: Create campaign_videos table
-- Version: 002

CREATE TABLE IF NOT EXISTS campaign_videos (
    id UUID PRIMARY KEY,
    campaign_id UUID NOT NULL REFERENCES campaigns(id) ON DELETE CASCADE,
    video_id UUID NOT NULL,
    status VARCHAR(50) NOT NULL DEFAULT 'active',
    added_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    UNIQUE(campaign_id, video_id)
);

CREATE INDEX IF NOT EXISTS idx_campaign_videos_campaign_id ON campaign_videos(campaign_id);
CREATE INDEX IF NOT EXISTS idx_campaign_videos_video_id ON campaign_videos(video_id);