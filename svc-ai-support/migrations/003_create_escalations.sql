-- 003_create_escalations.sql
-- Creates the escalations table for support escalations

CREATE TABLE IF NOT EXISTS escalations (
    id UUID PRIMARY KEY,
    conversation_id UUID NOT NULL REFERENCES conversations(id) ON DELETE CASCADE,
    escalated_by VARCHAR(10) NOT NULL CHECK (escalated_by IN ('user', 'admin', 'auto')),
    reason TEXT,
    admin_id UUID,
    status VARCHAR(10) NOT NULL DEFAULT 'open' CHECK (status IN ('open', 'resolved')),
    notes TEXT,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    resolved_at TIMESTAMP WITH TIME ZONE
);

CREATE INDEX idx_escalations_conversation_id ON escalations(conversation_id);
CREATE INDEX idx_escalations_status ON escalations(status);
CREATE INDEX idx_escalations_admin_id ON escalations(admin_id);