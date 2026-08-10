CREATE TABLE infrared_state_coder (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid (),
    infrared_device_id UUID NOT NULL REFERENCES infrared_device (id) ON DELETE CASCADE,
    infrared_record_session_id UUID NOT NULL REFERENCES infrared_record_session (id) ON DELETE CASCADE,
    encoder_source TEXT NOT NULL,
    decoder_source TEXT NOT NULL,
    summary_readme TEXT NOT NULL,
    detail_readme TEXT NOT NULL,
    status TEXT NOT NULL DEFAULT 'UNVERIFIED',
    CONSTRAINT chk_infrared_state_coder_status CHECK (status IN ('UNVERIFIED', 'ACTIVE', 'SUPERSEDED')),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_infrared_state_coder_infrared_device_id ON infrared_state_coder (infrared_device_id);
CREATE INDEX idx_infrared_state_coder_infrared_record_session_id ON infrared_state_coder (infrared_record_session_id);
