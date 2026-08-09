CREATE TABLE infrared_device (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid (),
    infrared_device_type_id UUID NOT NULL REFERENCES infrared_device_type (id) ON DELETE CASCADE,
    brand TEXT NOT NULL,
    model TEXT NOT NULL
);

CREATE INDEX idx_infrared_device_infrared_device_type_id ON infrared_device (infrared_device_type_id);

CREATE TABLE infrared_state_device_definition (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid (),
    infrared_device_id UUID NOT NULL REFERENCES infrared_device (id) ON DELETE CASCADE,
    infrared_state_id UUID NOT NULL REFERENCES infrared_state (id) ON DELETE CASCADE,
    options TEXT[],
    minimum DOUBLE PRECISION,
    maximum DOUBLE PRECISION,
    step DOUBLE PRECISION,
    CONSTRAINT uq_infrared_state_device_definition_device_state UNIQUE (infrared_device_id, infrared_state_id)
);

CREATE TABLE infrared_record_session (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid (),
    node_id UUID NOT NULL REFERENCES nodes (id) ON DELETE CASCADE,
    infrared_device_id UUID NOT NULL REFERENCES infrared_device (id) ON DELETE CASCADE,
    recording_state TEXT NOT NULL DEFAULT 'DRAFT',
    current_record_case_id UUID,
    is_completed BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_infrared_record_session_node_id ON infrared_record_session (node_id);
CREATE INDEX idx_infrared_record_session_infrared_device_id ON infrared_record_session (infrared_device_id);

CREATE TABLE infrared_state_device_record_case (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid (),
    infrared_record_session_id UUID NOT NULL REFERENCES infrared_record_session (id) ON DELETE CASCADE,
    step INTEGER NOT NULL,
    description TEXT NOT NULL DEFAULT '',
    status TEXT NOT NULL DEFAULT 'PENDING',
    CONSTRAINT chk_infrared_record_case_status CHECK (status IN ('PENDING', 'ACTIVE', 'ACCEPTED')),
    CONSTRAINT uq_infrared_record_case_session_step UNIQUE (infrared_record_session_id, step)
);

CREATE INDEX idx_infrared_record_case_session_id ON infrared_state_device_record_case (infrared_record_session_id);

-- current_record_case_id is added as a foreign key here, once the table it
-- references exists, rather than on infrared_record_session's own CREATE.
ALTER TABLE infrared_record_session
    ADD CONSTRAINT fk_infrared_record_session_current_case
    FOREIGN KEY (current_record_case_id) REFERENCES infrared_state_device_record_case (id) ON DELETE SET NULL;

CREATE TABLE infrared_state_device_record_state (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid (),
    infrared_state_device_record_case_id UUID NOT NULL REFERENCES infrared_state_device_record_case (id) ON DELETE CASCADE,
    infrared_state_id UUID NOT NULL REFERENCES infrared_state (id) ON DELETE CASCADE,
    state_value TEXT NOT NULL,
    CONSTRAINT uq_infrared_record_state_case_state UNIQUE (infrared_state_device_record_case_id, infrared_state_id)
);

CREATE INDEX idx_infrared_record_state_case_id ON infrared_state_device_record_state (infrared_state_device_record_case_id);

CREATE TABLE infrared_state_device_record_raw (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid (),
    infrared_state_device_record_case_id UUID NOT NULL REFERENCES infrared_state_device_record_case (id) ON DELETE CASCADE,
    raw_data BYTEA NOT NULL,
    status TEXT NOT NULL DEFAULT 'CAPTURED',
    discarded_reason TEXT,
    CONSTRAINT chk_infrared_record_raw_status CHECK (status IN ('CAPTURED', 'ACCEPTED', 'DISCARDED'))
);

CREATE INDEX idx_infrared_record_raw_case_id ON infrared_state_device_record_raw (infrared_state_device_record_case_id);
