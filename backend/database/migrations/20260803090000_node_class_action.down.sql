DROP TABLE IF EXISTS node_class_action;

ALTER TABLE actions ADD COLUMN node_class_id UUID;
