-- Create roles and permissions tables
-- +migrate Up

-- Roles table
CREATE TABLE IF NOT EXISTS roles (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v7(),
    name VARCHAR(50) NOT NULL UNIQUE,
    description TEXT
);

CREATE INDEX IF NOT EXISTS idx_roles_name ON roles(name);

-- Permissions table
CREATE TABLE IF NOT EXISTS permissions (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v7(),
    name VARCHAR(100) NOT NULL UNIQUE,
    description TEXT
);

CREATE INDEX IF NOT EXISTS idx_permissions_name ON permissions(name);

-- User-Roles junction table
CREATE TABLE IF NOT EXISTS user_roles (
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    role_id UUID NOT NULL REFERENCES roles(id) ON DELETE CASCADE,
    PRIMARY KEY (user_id, role_id)
);

CREATE INDEX IF NOT EXISTS idx_user_roles_user ON user_roles(user_id);
CREATE INDEX IF NOT EXISTS idx_user_roles_role ON user_roles(role_id);

-- Insert default roles
INSERT INTO roles (id, name, description) VALUES 
    (uuid_generate_v7(), 'client', 'Regular client user'),
    (uuid_generate_v7(), 'editor', 'Content editor with modification access'),
    (uuid_generate_v7(), 'ad_specialist', 'Advertising specialist'),
    (uuid_generate_v7(), 'admin', 'System administrator'),
    (uuid_generate_v7(), 'support_ai', 'AI-powered support assistant')
ON CONFLICT (name) DO NOTHING;

-- Insert default permissions
INSERT INTO permissions (id, name, description) VALUES 
    (uuid_generate_v7(), 'users:read', 'Read user information'),
    (uuid_generate_v7(), 'users:ban', 'Ban user accounts'),
    (uuid_generate_v7(), 'payouts:override', 'Override payout calculations'),
    (uuid_generate_v7(), 'campaigns:audit', 'Audit campaign data'),
    (uuid_generate_v7(), 'support:escalate', 'Escalate support tickets')
ON CONFLICT (name) DO NOTHING;

-- +migrate Down
DROP TABLE IF EXISTS user_roles;
DROP TABLE IF EXISTS permissions;
DROP TABLE IF EXISTS roles;