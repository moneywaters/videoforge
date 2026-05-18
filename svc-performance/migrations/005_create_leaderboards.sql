-- Create leaderboards table
-- Migration: 005_create_leaderboards.sql

CREATE TABLE IF NOT EXISTS leaderboards (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    brief_id UUID NOT NULL,
    entity_type VARCHAR(20) NOT NULL CHECK (entity_type IN ('editor', 'specialist', 'video')),
    entity_id UUID NOT NULL,
    rank INTEGER NOT NULL,
    total_revenue DECIMAL(15, 2) NOT NULL DEFAULT 0,
    total_orders INTEGER NOT NULL DEFAULT 0,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    UNIQUE(brief_id, entity_type, entity_id)
);

CREATE INDEX idx_leaderboards_brief_id ON leaderboards(brief_id);
CREATE INDEX idx_leaderboards_brief_entity ON leaderboards(brief_id, entity_type, rank);