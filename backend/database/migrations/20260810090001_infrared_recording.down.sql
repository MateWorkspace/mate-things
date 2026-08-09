DROP TABLE IF EXISTS infrared_state_device_record_raw;
DROP TABLE IF EXISTS infrared_state_device_record_state;
ALTER TABLE IF EXISTS infrared_record_session DROP CONSTRAINT IF EXISTS fk_infrared_record_session_current_case;
DROP TABLE IF EXISTS infrared_state_device_record_case;
DROP TABLE IF EXISTS infrared_record_session;
DROP TABLE IF EXISTS infrared_state_device_definition;
DROP TABLE IF EXISTS infrared_device;
