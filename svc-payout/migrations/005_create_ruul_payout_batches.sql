-- 005_create_ruul_payout_batches.sql
-- Ruul.io payout batches for bulk payouts to freelancers

CREATE TABLE IF NOT EXISTS ruul_payout_batches (
    id UUID PRIMARY KEY,
    batch_name VARCHAR(255) NOT NULL,
    status VARCHAR(20) NOT NULL DEFAULT 'draft' CHECK (status IN ('draft', 'processing', 'completed', 'failed')),
    total_amount DECIMAL(18,2) NOT NULL DEFAULT 0.00,
    currency VARCHAR(10) NOT NULL DEFAULT 'USD',
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_ruul_payout_batches_status ON ruul_payout_batches(status);
CREATE INDEX IF NOT EXISTS idx_ruul_payout_batches_created_at ON ruul_payout_batches(created_at);