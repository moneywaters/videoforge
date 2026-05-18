-- Migration 001: Create shopify_stores table
-- Stores connected Shopify stores for clients

CREATE SCHEMA IF NOT EXISTS shopify;

-- Shopify stores table
CREATE TABLE IF NOT EXISTS shopify_stores (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    client_id UUID NOT NULL,
    shop_domain VARCHAR(256) NOT NULL,
    access_token TEXT NOT NULL,
    status VARCHAR(32) NOT NULL DEFAULT 'active' CHECK (status IN ('active', 'disconnected')),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    UNIQUE(shop_domain)
);

-- Index for looking up stores by client
CREATE INDEX IF NOT EXISTS idx_shopify_stores_client ON shopify_stores(client_id);

-- Index for looking up stores by domain
CREATE INDEX IF NOT EXISTS idx_shopify_stores_domain ON shopify_stores(shop_domain);

GRANT USAGE ON SCHEMA shopify TO videoforge;
GRANT SELECT, INSERT, UPDATE, DELETE ON ALL TABLES IN SCHEMA shopify TO videoforge;
GRANT USAGE, SELECT ON ALL SEQUENCES IN SCHEMA shopify TO videoforge;