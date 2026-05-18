-- Migration 004: Create attributions table
-- Stores attribution records linking orders to videos

CREATE TABLE IF NOT EXISTS attributions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    order_id UUID NOT NULL REFERENCES orders(id) ON DELETE CASCADE,
    video_id UUID NOT NULL,
    campaign_id UUID,
    attributed_amount DECIMAL(12, 2) NOT NULL DEFAULT 0,
    attribution_method VARCHAR(32) NOT NULL CHECK (attribution_method IN ('discount_code', 'utm')),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Index for looking up attributions by order
CREATE INDEX IF NOT EXISTS idx_attributions_order ON attributions(order_id);

-- Index for looking up attributions by video
CREATE INDEX IF NOT EXISTS idx_attributions_video ON attributions(video_id);

-- Index for looking up attributions by campaign
CREATE INDEX IF NOT EXISTS idx_attributions_campaign ON attributions(campaign_id);

-- Index for looking up attributions by method
CREATE INDEX IF NOT EXISTS idx_attributions_method ON attributions(attribution_method);