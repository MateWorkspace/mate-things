SELECT remove_compression_policy ('node_logs', if_exists => TRUE);

DROP TABLE IF EXISTS node_logs;

DROP TYPE IF EXISTS node_log_level;
