-- Migration: Create ad_accounts table (placeholder)
-- Version: 003

CREATE TABLE IF NOT EXISTS ad_accounts (
    id UUID PRIMARY KEY,
    user_id UUID NOT NULL,
    platform VARCHAR(50) NOT NULL,
    account_id VARCHAR(255) NOT NULL,
    name VARCHAR(255),
    status VARCHAR(50) NOT NULL DEFAULT 'active',
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    UNIQUE(user_id, platform, account_id)
);

CREATE INDEX IF NOT EXISTS idx_ad_accounts_user_id ON ad_accounts(user_id);
CREATE INDEX IF NOT EXISTS idx_ad_accounts_platform ON ad_accounts(platform);