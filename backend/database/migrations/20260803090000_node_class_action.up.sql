ALTER TABLE actions DROP CONSTRAINT actions_node_class_id_fkey;

DROP INDEX IF EXISTS idx_actions_node_class_id;

ALTER TABLE actions DROP COLUMN node_class_id;

CREATE TABLE node_class_action (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid (),
    node_class_id UUID NOT NULL REFERENCES node_classes (id) ON DELETE CASCADE,
    action_id UUID NOT NULL REFERENCES actions (id) ON DELETE CASCADE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    created_by UUID,
    CONSTRAINT uq_node_class_action_node_class_id_action_id UNIQUE (node_class_id, action_id)
);

CREATE INDEX idx_node_class_action_node_class_id ON node_class_action (node_class_id);

CREATE INDEX idx_node_class_action_action_id ON node_class_action (action_id);
