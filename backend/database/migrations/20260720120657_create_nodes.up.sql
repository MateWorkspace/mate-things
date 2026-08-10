CREATE TABLE nodes (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid (),
    node_class_id UUID NOT NULL REFERENCES node_classes (id),
    device_id TEXT NOT NULL UNIQUE,
    device_info TEXT NOT NULL,
    firmware_id UUID REFERENCES firmwares (id),
    name TEXT NOT NULL,
    description TEXT NOT NULL DEFAULT '',
    is_connected BOOLEAN NOT NULL DEFAULT FALSE,
    preferences JSONB NOT NULL DEFAULT '{}'::jsonb,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ,
    deleted_at TIMESTAMPTZ,
    created_by UUID,
    updated_by UUID,
    deleted_by UUID
);

CREATE INDEX idx_nodes_node_class_id ON nodes (node_class_id);

CREATE INDEX idx_nodes_device_id_trgm ON nodes USING GIN (device_id gin_trgm_ops);

CREATE INDEX idx_nodes_device_info_trgm ON nodes USING GIN (device_info gin_trgm_ops);

CREATE INDEX idx_nodes_name_trgm ON nodes USING GIN (name gin_trgm_ops);

CREATE INDEX idx_nodes_firmware_id ON nodes (firmware_id);

CREATE INDEX idx_nodes_deleted_at ON nodes (deleted_at);
