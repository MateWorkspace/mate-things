CREATE TABLE firmware_config_parameters (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid (),
    firmware_id UUID NOT NULL REFERENCES firmwares (id),
    key TEXT NOT NULL,
    value_type TEXT NOT NULL,
    preferences JSONB NOT NULL DEFAULT '{}'::jsonb,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ,
    deleted_at TIMESTAMPTZ,
    created_by UUID,
    updated_by UUID,
    deleted_by UUID,
    CONSTRAINT uq_firmware_config_parameters_firmware_id_key UNIQUE (firmware_id, key)
);

CREATE INDEX idx_firmware_config_parameters_firmware_id ON firmware_config_parameters (firmware_id);

CREATE INDEX idx_firmware_config_parameters_deleted_at ON firmware_config_parameters (deleted_at);
