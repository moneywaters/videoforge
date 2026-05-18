-- Admin actions table for tracking admin operations
CREATE TABLE IF NOT EXISTS admin_actions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    admin_id UUID NOT NULL REFERENCES users(id),
    action_type VARCHAR(32) NOT NULL CHECK (action_type IN ('ban', 'role_change', 'payout_override', 'content_flag')),
    target_user_id UUID REFERENCES users(id),
    target_type VARCHAR(32) NOT NULL CHECK (target_type IN ('user', 'video', 'brief', 'campaign')),
    target_id UUID,
    reason TEXT NOT NULL,
    metadata JSONB DEFAULT '{}',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Indexes for admin_actions
CREATE INDEX IF NOT EXISTS idx_admin_actions_admin_id ON admin_actions(admin_id);
CREATE INDEX IF NOT EXISTS idx_admin_actions_target_user ON admin_actions(target_user_id);
CREATE INDEX IF NOT EXISTS idx_admin_actions_target ON admin_actions(target_type, target_id);
CREATE INDEX IF NOT EXISTS idx_admin_actions_type ON admin_actions(action_type);
CREATE INDEX IF NOT EXISTS idx_admin_actions_created ON admin_actions(created_at);

-- Grant permissions
GRANT USAGE ON SCHEMA admin TO videoforge;
GRANT SELECT, INSERT, UPDATE, DELETE ON ALL TABLES IN SCHEMA admin TO videoforge;
GRANT USAGE, SELECT ON ALL SEQUENCES IN SCHEMA admin TO videoforge;