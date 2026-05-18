-- 002_create_payout_rules.sql
-- Payout rules for platform fee configuration

CREATE TABLE IF NOT EXISTS payout_rules (
    id UUID PRIMARY KEY,
    name VARCHAR(100) UNIQUE NOT NULL,
    threshold_amount DECIMAL(18,2) NOT NULL DEFAULT 500.00,
    platform_fee_percent DECIMAL(5,2) NOT NULL DEFAULT 5.00,
    description TEXT,
    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);

-- Insert default rule
INSERT INTO payout_rules (id, name, threshold_amount, platform_fee_percent, description, created_at)
VALUES (
    gen_random_uuid(),
    'default',
    500.00,
    5.00,
    'Default rule: $0 platform fee for first $500 in verified sales, then 5% platform fee after',
    NOW()
) ON CONFLICT (name) DO NOTHING;