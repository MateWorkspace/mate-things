CREATE TABLE role_permission (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid (),
    role_id UUID NOT NULL REFERENCES roles (id) ON DELETE CASCADE,
    permission_id UUID NOT NULL REFERENCES permissions (id) ON DELETE CASCADE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    created_by UUID,
    CONSTRAINT uq_role_permission_role_id_permission_id UNIQUE (role_id, permission_id)
);

CREATE INDEX idx_role_permission_role_id ON role_permission (role_id);

CREATE INDEX idx_role_permission_permission_id ON role_permission (permission_id);