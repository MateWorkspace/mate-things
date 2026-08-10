ALTER TABLE infrared_device_type
    ADD COLUMN created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    ADD COLUMN updated_at TIMESTAMPTZ,
    ADD COLUMN deleted_at TIMESTAMPTZ,
    ADD COLUMN created_by UUID,
    ADD COLUMN updated_by UUID,
    ADD COLUMN deleted_by UUID;
CREATE INDEX idx_infrared_device_type_deleted_at ON infrared_device_type (deleted_at);

ALTER TABLE infrared_state
    ADD COLUMN created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    ADD COLUMN updated_at TIMESTAMPTZ,
    ADD COLUMN deleted_at TIMESTAMPTZ,
    ADD COLUMN created_by UUID,
    ADD COLUMN updated_by UUID,
    ADD COLUMN deleted_by UUID;
CREATE INDEX idx_infrared_state_deleted_at ON infrared_state (deleted_at);

ALTER TABLE infrared_device
    ADD COLUMN created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    ADD COLUMN updated_at TIMESTAMPTZ,
    ADD COLUMN deleted_at TIMESTAMPTZ,
    ADD COLUMN created_by UUID,
    ADD COLUMN updated_by UUID,
    ADD COLUMN deleted_by UUID;
CREATE INDEX idx_infrared_device_deleted_at ON infrared_device (deleted_at);

ALTER TABLE infrared_state_device_definition
    ADD COLUMN created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    ADD COLUMN updated_at TIMESTAMPTZ,
    ADD COLUMN deleted_at TIMESTAMPTZ,
    ADD COLUMN created_by UUID,
    ADD COLUMN updated_by UUID,
    ADD COLUMN deleted_by UUID;
CREATE INDEX idx_infrared_state_device_definition_deleted_at ON infrared_state_device_definition (deleted_at);

ALTER TABLE infrared_record_session
    ADD COLUMN updated_at TIMESTAMPTZ,
    ADD COLUMN deleted_at TIMESTAMPTZ,
    ADD COLUMN created_by UUID,
    ADD COLUMN updated_by UUID,
    ADD COLUMN deleted_by UUID;
CREATE INDEX idx_infrared_record_session_deleted_at ON infrared_record_session (deleted_at);

ALTER TABLE infrared_state_device_record_case
    ADD COLUMN created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    ADD COLUMN updated_at TIMESTAMPTZ,
    ADD COLUMN deleted_at TIMESTAMPTZ,
    ADD COLUMN created_by UUID,
    ADD COLUMN updated_by UUID,
    ADD COLUMN deleted_by UUID;
CREATE INDEX idx_infrared_record_case_deleted_at ON infrared_state_device_record_case (deleted_at);

ALTER TABLE infrared_state_device_record_state
    ADD COLUMN created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    ADD COLUMN updated_at TIMESTAMPTZ,
    ADD COLUMN deleted_at TIMESTAMPTZ,
    ADD COLUMN created_by UUID,
    ADD COLUMN updated_by UUID,
    ADD COLUMN deleted_by UUID;
CREATE INDEX idx_infrared_record_state_deleted_at ON infrared_state_device_record_state (deleted_at);

ALTER TABLE infrared_state_device_record_raw
    ADD COLUMN created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    ADD COLUMN updated_at TIMESTAMPTZ,
    ADD COLUMN deleted_at TIMESTAMPTZ,
    ADD COLUMN created_by UUID,
    ADD COLUMN updated_by UUID,
    ADD COLUMN deleted_by UUID;
CREATE INDEX idx_infrared_record_raw_deleted_at ON infrared_state_device_record_raw (deleted_at);

ALTER TABLE infrared_state_coder
    ADD COLUMN updated_at TIMESTAMPTZ,
    ADD COLUMN deleted_at TIMESTAMPTZ,
    ADD COLUMN created_by UUID,
    ADD COLUMN updated_by UUID,
    ADD COLUMN deleted_by UUID;
CREATE INDEX idx_infrared_state_coder_deleted_at ON infrared_state_coder (deleted_at);

ALTER TABLE infrared_test_case
    ADD COLUMN created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    ADD COLUMN updated_at TIMESTAMPTZ,
    ADD COLUMN deleted_at TIMESTAMPTZ,
    ADD COLUMN created_by UUID,
    ADD COLUMN updated_by UUID,
    ADD COLUMN deleted_by UUID;
CREATE INDEX idx_infrared_test_case_deleted_at ON infrared_test_case (deleted_at);

ALTER TABLE infrared_test_case_state
    ADD COLUMN created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    ADD COLUMN updated_at TIMESTAMPTZ,
    ADD COLUMN deleted_at TIMESTAMPTZ,
    ADD COLUMN created_by UUID,
    ADD COLUMN updated_by UUID,
    ADD COLUMN deleted_by UUID;
CREATE INDEX idx_infrared_test_case_state_deleted_at ON infrared_test_case_state (deleted_at);
