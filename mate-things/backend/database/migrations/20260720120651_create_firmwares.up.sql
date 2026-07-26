CREATE TABLE firmwares (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid (),
    node_class_id UUID NOT NULL REFERENCES node_classes (id),
    name TEXT NOT NULL UNIQUE,
    size INT NOT NULL,
    checksum TEXT NOT NULL,
    binary_path TEXT NOT NULL,
    preferences JSONB NOT NULL DEFAULT '{}'::jsonb,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ,
    deleted_at TIMESTAMPTZ,
    created_by UUID,
    updated_by UUID,
    deleted_by UUID
);

CREATE INDEX idx_firmwares_node_class_id ON firmwares (node_class_id);

CREATE INDEX idx_firmwares_name_trgm ON firmwares USING GIN (name gin_trgm_ops);

CREATE INDEX idx_firmwares_binary_path_trgm ON firmwares USING GIN (binary_path gin_trgm_ops);

CREATE INDEX idx_firmwares_deleted_at ON firmwares (deleted_at);
