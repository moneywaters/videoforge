-- Create video_sales table
-- Migration: 001_create_video_sales.sql

CREATE TABLE IF NOT EXISTS video_sales (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    video_id UUID NOT NULL UNIQUE,
    campaign_id UUID NOT NULL,
    total_orders INTEGER NOT NULL DEFAULT 0,
    total_revenue DECIMAL(15, 2) NOT NULL DEFAULT 0,
    currency VARCHAR(3) NOT NULL DEFAULT 'USD',
    first_sale_at TIMESTAMP WITH TIME ZONE,
    last_sale_at TIMESTAMP WITH TIME ZONE,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_video_sales_video_id ON video_sales(video_id);
CREATE INDEX idx_video_sales_campaign_id ON video_sales(campaign_id);