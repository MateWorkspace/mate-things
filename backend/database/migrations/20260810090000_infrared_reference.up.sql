CREATE TABLE infrared_device_type (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid (),
    name TEXT NOT NULL UNIQUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ,
    deleted_at TIMESTAMPTZ,
    created_by UUID,
    updated_by UUID,
    deleted_by UUID
);

CREATE INDEX idx_infrared_device_type_deleted_at ON infrared_device_type (deleted_at);

CREATE TABLE infrared_state (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid (),
    infrared_device_type_id UUID NOT NULL REFERENCES infrared_device_type (id) ON DELETE CASCADE,
    name TEXT NOT NULL,
    type TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ,
    deleted_at TIMESTAMPTZ,
    created_by UUID,
    updated_by UUID,
    deleted_by UUID,
    CONSTRAINT chk_infrared_state_type CHECK (type IN ('RANGE', 'ENUM')),
    CONSTRAINT uq_infrared_state_device_type_id_name UNIQUE (infrared_device_type_id, name)
);

CREATE INDEX idx_infrared_state_infrared_device_type_id ON infrared_state (infrared_device_type_id);
CREATE INDEX idx_infrared_state_deleted_at ON infrared_state (deleted_at);
