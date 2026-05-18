-- Migration 003: Create orders table
-- Stores Shopify orders received via webhooks

CREATE TABLE IF NOT EXISTS orders (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    shopify_order_id VARCHAR(64) NOT NULL,
    store_id UUID NOT NULL,
    customer_email VARCHAR(256),
    total_price DECIMAL(12, 2) NOT NULL DEFAULT 0,
    currency VARCHAR(8) NOT NULL DEFAULT 'USD',
    discount_code VARCHAR(64),
    utm_source VARCHAR(128),
    utm_medium VARCHAR(128),
    utm_campaign VARCHAR(128),
    status VARCHAR(32) NOT NULL DEFAULT 'pending' CHECK (status IN ('pending', 'confirmed', 'cancelled')),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    UNIQUE(shopify_order_id, store_id)
);

-- Index for looking up orders by store
CREATE INDEX IF NOT EXISTS idx_orders_store ON orders(store_id);

-- Index for looking up orders by status
CREATE INDEX IF NOT EXISTS idx_orders_status ON orders(status);

-- Index for looking up orders by discount code
CREATE INDEX IF NOT EXISTS idx_orders_discount ON orders(discount_code);

-- Index for looking up orders by UTM campaign
CREATE INDEX IF NOT EXISTS idx_orders_utm_campaign ON orders(utm_campaign);

-- Index for looking up orders by creation date
CREATE INDEX IF NOT EXISTS idx_orders_created ON orders(created_at);