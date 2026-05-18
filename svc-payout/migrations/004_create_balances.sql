-- 004_create_balances.sql
-- User balances (available and pending)

CREATE TABLE IF NOT EXISTS balances (
    id UUID PRIMARY KEY,
    user_id UUID UNIQUE NOT NULL,
    available DECIMAL(18,2) NOT NULL DEFAULT 0.00,
    pending DECIMAL(18,2) NOT NULL DEFAULT 0.00,
    total_earned DECIMAL(18,2) NOT NULL DEFAULT 0.00,
    currency VARCHAR(10) NOT NULL DEFAULT 'USD',
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_balances_user_id ON balances(user_id);