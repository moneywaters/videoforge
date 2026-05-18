-- Initialize Admin schema
CREATE SCHEMA IF NOT EXISTS admin;

-- Audit logs table
CREATE TABLE IF NOT EXISTS admin.audit_logs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID REFERENCES users(id),
    action VARCHAR(64) NOT NULL,
    resource VARCHAR(64) NOT NULL,
    resource_id UUID,
    details JSONB,
    ip_address VARCHAR(64),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Admin sessions table
CREATE TABLE IF NOT EXISTS admin.sessions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id),
    token TEXT NOT NULL,
    ip_address VARCHAR(64),
    user_agent TEXT,
    expires_at TIMESTAMP WITH TIME ZONE NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- System settings table
CREATE TABLE IF NOT EXISTS admin.settings (
    key VARCHAR(128) PRIMARY KEY,
    value JSONB NOT NULL,
    description TEXT,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Indexes
CREATE INDEX IF NOT EXISTS idx_admin_audit_user ON admin.audit_logs(user_id);
CREATE INDEX IF NOT EXISTS idx_admin_audit_resource ON admin.audit_logs(resource, resource_id);
CREATE INDEX IF NOT EXISTS idx_admin_audit_created ON admin.audit_logs(created_at);
CREATE INDEX IF NOT EXISTS idx_admin_sessions_user ON admin.sessions(user_id);
CREATE INDEX IF NOT EXISTS idx_admin_sessions_token ON admin.sessions(token);

-- Grant permissions
GRANT USAGE ON SCHEMA admin TO videoforge;
GRANT SELECT, INSERT, UPDATE, DELETE ON ALL TABLES IN SCHEMA admin TO videoforge;
GRANT USAGE, SELECT ON ALL SEQUENCES IN SCHEMA admin TO videoforge;