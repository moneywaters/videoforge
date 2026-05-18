-- Create specialist_sales table
-- Migration: 003_create_specialist_sales.sql

CREATE TABLE IF NOT EXISTS specialist_sales (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    specialist_id UUID NOT NULL UNIQUE,
    total_campaigns INTEGER NOT NULL DEFAULT 0,
    total_orders INTEGER NOT NULL DEFAULT 0,
    total_revenue DECIMAL(15, 2) NOT NULL DEFAULT 0,
    currency VARCHAR(3) NOT NULL DEFAULT 'USD',
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_specialist_sales_specialist_id ON specialist_sales(specialist_id);