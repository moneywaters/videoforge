-- Create daily_metrics table
-- Migration: 006_create_daily_metrics.sql

CREATE TABLE IF NOT EXISTS daily_metrics (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    date DATE NOT NULL,
    video_id UUID NOT NULL,
    campaign_id UUID NOT NULL,
    orders INTEGER NOT NULL DEFAULT 0,
    revenue DECIMAL(15, 2) NOT NULL DEFAULT 0,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    UNIQUE(date, video_id)
);

CREATE INDEX idx_daily_metrics_date ON daily_metrics(date);
CREATE INDEX idx_daily_metrics_video_id ON daily_metrics(video_id);
CREATE INDEX idx_daily_metrics_campaign_id ON daily_metrics(campaign_id);