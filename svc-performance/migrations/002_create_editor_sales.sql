-- Create editor_sales table
-- Migration: 002_create_editor_sales.sql

CREATE TABLE IF NOT EXISTS editor_sales (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    editor_id UUID NOT NULL UNIQUE,
    total_videos INTEGER NOT NULL DEFAULT 0,
    total_orders INTEGER NOT NULL DEFAULT 0,
    total_revenue DECIMAL(15, 2) NOT NULL DEFAULT 0,
    currency VARCHAR(3) NOT NULL DEFAULT 'USD',
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_editor_sales_editor_id ON editor_sales(editor_id);