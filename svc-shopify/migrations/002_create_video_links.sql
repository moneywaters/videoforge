-- Migration 002: Create video_links table
-- Stores custom discount/affiliate links per video

CREATE TABLE IF NOT EXISTS video_links (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    video_id UUID NOT NULL,
    campaign_id UUID,
    discount_code VARCHAR(64) UNIQUE NOT NULL,
    utm_source VARCHAR(128),
    utm_medium VARCHAR(128),
    utm_campaign VARCHAR(128),
    url TEXT NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Index for looking up links by video
CREATE INDEX IF NOT EXISTS idx_video_links_video ON video_links(video_id);

-- Index for looking up links by campaign
CREATE INDEX IF NOT EXISTS idx_video_links_campaign ON video_links(campaign_id);

-- Index for looking up links by discount code
CREATE INDEX IF NOT EXISTS idx_video_links_discount ON video_links(discount_code);

-- Index for looking up links by UTM source
CREATE INDEX IF NOT EXISTS idx_video_links_utm_source ON video_links(utm_source);

-- Index for looking up links by UTM campaign
CREATE INDEX IF NOT EXISTS idx_video_links_utm_campaign ON video_links(utm_campaign);