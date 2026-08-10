CREATE TABLE infrared_test_case (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid (),
    infrared_state_coder_id UUID NOT NULL REFERENCES infrared_state_coder (id) ON DELETE CASCADE,
    step INTEGER NOT NULL,
    description TEXT NOT NULL DEFAULT '',
    status TEXT NOT NULL DEFAULT 'PENDING',
    CONSTRAINT chk_infrared_test_case_status CHECK (status IN ('PENDING', 'PASSED', 'FAILED')),
    CONSTRAINT uq_infrared_test_case_coder_step UNIQUE (infrared_state_coder_id, step)
);

CREATE INDEX idx_infrared_test_case_coder_id ON infrared_test_case (infrared_state_coder_id);

CREATE TABLE infrared_test_case_state (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid (),
    infrared_test_case_id UUID NOT NULL REFERENCES infrared_test_case (id) ON DELETE CASCADE,
    infrared_state_id UUID NOT NULL REFERENCES infrared_state (id) ON DELETE CASCADE,
    state_value TEXT NOT NULL,
    CONSTRAINT uq_infrared_test_case_state_case_state UNIQUE (infrared_test_case_id, infrared_state_id)
);

CREATE INDEX idx_infrared_test_case_state_test_case_id ON infrared_test_case_state (infrared_test_case_id);
