DROP INDEX IF EXISTS uq_firmwares_name;
ALTER TABLE firmwares ADD CONSTRAINT firmwares_name_key UNIQUE (name);

DROP INDEX IF EXISTS uq_users_username;
ALTER TABLE users ADD CONSTRAINT users_username_key UNIQUE (username);

DROP INDEX IF EXISTS uq_actions_name;
ALTER TABLE actions ADD CONSTRAINT actions_name_key UNIQUE (name);

DROP INDEX IF EXISTS uq_node_classes_name;
ALTER TABLE node_classes ADD CONSTRAINT node_classes_name_key UNIQUE (name);

DROP INDEX IF EXISTS uq_roles_name;
ALTER TABLE roles ADD CONSTRAINT roles_name_key UNIQUE (name);

DROP INDEX IF EXISTS uq_permissions_name;
ALTER TABLE permissions ADD CONSTRAINT permissions_name_key UNIQUE (name);
