-- 006_create_ruul_payout_requests.sql
-- Ruul.io payout requests (individual payouts in a batch)

CREATE TABLE IF NOT EXISTS ruul_payout_requests (
    id UUID PRIMARY KEY,
    batch_id UUID NOT NULL REFERENCES ruul_payout_batches(id) ON DELETE CASCADE,
    user_id UUID NOT NULL,
    amount DECIMAL(18,2) NOT NULL,
    currency VARCHAR(10) NOT NULL DEFAULT 'USD',
    status VARCHAR(20) NOT NULL DEFAULT 'pending' CHECK (status IN ('pending', 'processing', 'completed', 'failed')),
    ruul_reference_id VARCHAR(255),
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_ruul_payout_requests_batch_id ON ruul_payout_requests(batch_id);
CREATE INDEX IF NOT EXISTS idx_ruul_payout_requests_user_id ON ruul_payout_requests(user_id);
CREATE INDEX IF NOT EXISTS idx_ruul_payout_requests_status ON ruul_payout_requests(status);