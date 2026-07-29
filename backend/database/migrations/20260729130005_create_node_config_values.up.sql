CREATE TABLE node_config_values (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid (),
    node_id UUID NOT NULL REFERENCES nodes (id),
    firmware_id UUID NOT NULL REFERENCES firmwares (id),
    key TEXT NOT NULL,
    value TEXT NOT NULL,
    preferences JSONB NOT NULL DEFAULT '{}'::jsonb,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ,
    deleted_at TIMESTAMPTZ,
    created_by UUID,
    updated_by UUID,
    deleted_by UUID,
    CONSTRAINT uq_node_config_values_node_id_key UNIQUE (node_id, key)
);

CREATE INDEX idx_node_config_values_node_id ON node_config_values (node_id);

CREATE INDEX idx_node_config_values_firmware_id ON node_config_values (firmware_id);

CREATE INDEX idx_node_config_values_deleted_at ON node_config_values (deleted_at);
