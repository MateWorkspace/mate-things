-- Convert plain UNIQUE constraints on soft-deletable name/identifier columns
-- into partial unique indexes scoped to non-deleted rows, so a name freed up
-- by a soft delete can be reused. Two columns are deliberately excluded
-- because another table's foreign key targets them by that column (not by
-- id), and PostgreSQL foreign keys can only target a real UNIQUE constraint,
-- never a partial index:
--   - payload_schemas(name, version): referenced by
--     actions.fk_actions_payload_schema and
--     telemetry_records.fk_telemetry_records_payload_schema.
--   - nodes(device_id): referenced by
--     telemetry_records.telemetry_records_node_device_id_fkey.

ALTER TABLE permissions DROP CONSTRAINT permissions_name_key;
CREATE UNIQUE INDEX uq_permissions_name ON permissions (name) WHERE deleted_at IS NULL;

ALTER TABLE roles DROP CONSTRAINT roles_name_key;
CREATE UNIQUE INDEX uq_roles_name ON roles (name) WHERE deleted_at IS NULL;

ALTER TABLE node_classes DROP CONSTRAINT node_classes_name_key;
CREATE UNIQUE INDEX uq_node_classes_name ON node_classes (name) WHERE deleted_at IS NULL;

ALTER TABLE actions DROP CONSTRAINT actions_name_key;
CREATE UNIQUE INDEX uq_actions_name ON actions (name) WHERE deleted_at IS NULL;

ALTER TABLE users DROP CONSTRAINT users_username_key;
CREATE UNIQUE INDEX uq_users_username ON users (username) WHERE deleted_at IS NULL;

ALTER TABLE firmwares DROP CONSTRAINT firmwares_name_key;
CREATE UNIQUE INDEX uq_firmwares_name ON firmwares (name) WHERE deleted_at IS NULL;
