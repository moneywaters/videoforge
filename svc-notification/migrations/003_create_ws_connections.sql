-- Create ws_connections table
CREATE TABLE IF NOT EXISTS ws_connections (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL,
    connection_id VARCHAR(256) NOT NULL UNIQUE,
    connected_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    last_ping_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Indexes for ws_connections
CREATE INDEX IF NOT EXISTS idx_ws_connections_user_id ON ws_connections(user_id);
CREATE INDEX IF NOT EXISTS idx_ws_connections_connection_id ON ws_connections(connection_id);