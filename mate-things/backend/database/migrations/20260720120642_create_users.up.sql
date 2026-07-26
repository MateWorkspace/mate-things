CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid (),
    role_id UUID NOT NULL REFERENCES roles (id),
    name TEXT NOT NULL,
    bio TEXT NOT NULL DEFAULT '',
    username TEXT NOT NULL UNIQUE,
    password_hash TEXT NOT NULL,
    preferences JSONB NOT NULL DEFAULT '{}'::jsonb,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ,
    deleted_at TIMESTAMPTZ,
    created_by UUID,
    updated_by UUID,
    deleted_by UUID
);

CREATE INDEX idx_users_role_id ON users (role_id);

CREATE INDEX idx_users_name_trgm ON users USING GIN (name gin_trgm_ops);

CREATE INDEX idx_users_username_trgm ON users USING GIN (username gin_trgm_ops);

CREATE INDEX idx_users_deleted_at ON users (deleted_at);
