-- 001_create_payouts.sql
-- Payouts table for tracking editor and specialist earnings

CREATE TABLE IF NOT EXISTS payouts (
    id UUID PRIMARY KEY,
    user_id UUID NOT NULL,
    type VARCHAR(50) NOT NULL CHECK (type IN ('editor_fee', 'specialist_fee', 'platform_fee')),
    amount DECIMAL(18,2) NOT NULL,
    currency VARCHAR(10) NOT NULL DEFAULT 'USD',
    status VARCHAR(20) NOT NULL DEFAULT 'pending' CHECK (status IN ('pending', 'eligible', 'paid', 'failed')),
    hold_until TIMESTAMP,
    paid_at TIMESTAMP,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_payouts_user_id ON payouts(user_id);
CREATE INDEX IF NOT EXISTS idx_payouts_status ON payouts(status);
CREATE INDEX IF NOT EXISTS idx_payouts_hold_until ON payouts(hold_until);
CREATE INDEX IF NOT EXISTS idx_payouts_type ON payouts(type);
CREATE INDEX IF NOT EXISTS idx_payouts_created_at ON payouts(created_at);