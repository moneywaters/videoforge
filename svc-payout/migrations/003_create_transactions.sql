-- 003_create_transactions.sql
-- Transactions table for financial ledger

CREATE TABLE IF NOT EXISTS transactions (
    id UUID PRIMARY KEY,
    payout_id UUID REFERENCES payouts(id) ON DELETE SET NULL,
    type VARCHAR(50) NOT NULL CHECK (type IN ('earning', 'hold_release', 'fee', 'payout', 'refund')),
    amount DECIMAL(18,2) NOT NULL,
    currency VARCHAR(10) NOT NULL DEFAULT 'USD',
    description TEXT,
    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_transactions_payout_id ON transactions(payout_id);
CREATE INDEX IF NOT EXISTS idx_transactions_type ON transactions(type);
CREATE INDEX IF NOT EXISTS idx_transactions_created_at ON transactions(created_at);