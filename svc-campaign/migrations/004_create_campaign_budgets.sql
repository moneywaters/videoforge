-- Migration: Create campaign_budgets table
-- Version: 004

CREATE TABLE IF NOT EXISTS campaign_budgets (
    id UUID PRIMARY KEY,
    campaign_id UUID NOT NULL UNIQUE REFERENCES campaigns(id) ON DELETE CASCADE,
    amount DECIMAL(15, 2) NOT NULL DEFAULT 0,
    type VARCHAR(50) NOT NULL DEFAULT 'total',
    spent DECIMAL(15, 2) NOT NULL DEFAULT 0,
    remaining DECIMAL(15, 2) NOT NULL DEFAULT 0,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_campaign_budgets_campaign_id ON campaign_budgets(campaign_id);