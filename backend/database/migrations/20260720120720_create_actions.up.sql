CREATE TABLE actions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid (),
    node_class_id UUID NOT NULL REFERENCES node_classes (id),
    name TEXT NOT NULL UNIQUE,
    description TEXT NOT NULL DEFAULT '',
    payload_schema_name TEXT NOT NULL,
    payload_schema_version INT NOT NULL DEFAULT 1,
    preferences JSONB NOT NULL DEFAULT '{}'::jsonb,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ,
    deleted_at TIMESTAMPTZ,
    created_by UUID,
    updated_by UUID,
    deleted_by UUID,
    CONSTRAINT fk_actions_payload_schema FOREIGN KEY (payload_schema_name, payload_schema_version) REFERENCES payload_schemas (name, version)
);

CREATE INDEX idx_actions_node_class_id ON actions (node_class_id);

CREATE INDEX idx_actions_name_trgm ON actions USING GIN (name gin_trgm_ops);

CREATE INDEX idx_actions_payload_schema_name_version ON actions (payload_schema_name, payload_schema_version);

CREATE INDEX idx_actions_deleted_at ON actions (deleted_at);
