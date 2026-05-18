-- Migration: Create campaigns table
-- Version: 001

CREATE TABLE IF NOT EXISTS campaigns (
    id UUID PRIMARY KEY,
    ad_specialist_id UUID NOT NULL,
    client_id UUID NOT NULL,
    brief_id UUID,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    status VARCHAR(50) NOT NULL DEFAULT 'draft',
    platform VARCHAR(50) NOT NULL,
    ad_account_id VARCHAR(255),
    total_budget DECIMAL(15, 2) NOT NULL DEFAULT 0,
    daily_budget DECIMAL(15, 2) NOT NULL DEFAULT 0,
    start_date TIMESTAMP WITH TIME ZONE NOT NULL,
    end_date TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_campaigns_ad_specialist_id ON campaigns(ad_specialist_id);
CREATE INDEX IF NOT EXISTS idx_campaigns_client_id ON campaigns(client_id);
CREATE INDEX IF NOT EXISTS idx_campaigns_brief_id ON campaigns(brief_id);
CREATE INDEX IF NOT EXISTS idx_campaigns_status ON campaigns(status);
CREATE INDEX IF NOT EXISTS idx_campaigns_created_at ON campaigns(created_at DESC);