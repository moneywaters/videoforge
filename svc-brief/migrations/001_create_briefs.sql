-- Initialize Brief schema
CREATE SCHEMA IF NOT EXISTS brief;

-- Briefs table
CREATE TABLE IF NOT EXISTS brief.briefs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    client_id UUID NOT NULL,
    title VARCHAR(512) NOT NULL,
    description TEXT,
    goals TEXT,
    target_audience VARCHAR(256),
    tone VARCHAR(64),
    style_preferences VARCHAR(256),
    cta VARCHAR(512),
    status VARCHAR(32) DEFAULT 'draft',
    bounty_budget DECIMAL(12,2),
    bounty_deposited BOOLEAN DEFAULT FALSE,
    submissions_limit INTEGER DEFAULT 1,
    is_blind BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Indexes
CREATE INDEX IF NOT EXISTS idx_briefs_client ON brief.briefs(client_id);
CREATE INDEX IF NOT EXISTS idx_briefs_status ON brief.briefs(status);
CREATE INDEX IF NOT EXISTS idx_briefs_created ON brief.briefs(created_at DESC);

-- Grant permissions
GRANT USAGE ON SCHEMA brief TO videoforge;
GRANT SELECT, INSERT, UPDATE, DELETE ON ALL TABLES IN SCHEMA brief TO videoforge;
GRANT USAGE, SELECT ON ALL SEQUENCES IN SCHEMA brief TO videoforge;