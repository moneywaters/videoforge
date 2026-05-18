-- Initialize Gateway schema
CREATE SCHEMA IF NOT EXISTS gateway;

-- API Routes table
CREATE TABLE IF NOT EXISTS gateway.routes (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    path VARCHAR(512) NOT NULL,
    methods TEXT[] NOT NULL,
    service VARCHAR(128) NOT NULL,
    endpoint VARCHAR(512) NOT NULL,
    active BOOLEAN DEFAULT true,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- API Keys table
CREATE TABLE IF NOT EXISTS gateway.api_keys (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    key VARCHAR(256) NOT NULL UNIQUE,
    user_id UUID NOT NULL REFERENCES users(id),
    name VARCHAR(256) NOT NULL,
    expires_at TIMESTAMP WITH TIME ZONE,
    active BOOLEAN DEFAULT true,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Create indexes
CREATE INDEX IF NOT EXISTS idx_routes_service ON gateway.routes(service);
CREATE INDEX IF NOT EXISTS idx_api_keys_user ON gateway.api_keys(user_id);
CREATE INDEX IF NOT EXISTS idx_api_keys_active ON gateway.api_keys(active);

-- Grant permissions
GRANT USAGE ON SCHEMA gateway TO videoforge;
GRANT SELECT, INSERT, UPDATE, DELETE ON ALL TABLES IN SCHEMA gateway TO videoforge;
GRANT USAGE, SELECT ON ALL SEQUENCES IN SCHEMA gateway TO videoforge;