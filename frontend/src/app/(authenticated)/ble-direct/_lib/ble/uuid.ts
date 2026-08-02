/**
 * Every custom UUID this device exposes follows
 * 4d415445-SSSS-4700-CCCC-000000000000, where SSSS identifies the service
 * and CCCC the characteristic within it (0000 for the service UUID
 * itself). Mirrors mate-espidf-base's presentation/ble/gatt/uuid.h/.c —
 * see that file's comment for the byte-order derivation.
 */
function uuid(service: number, characteristic: number): string {
  const svc = service.toString(16).padStart(4, "0");
  const chr = characteristic.toString(16).padStart(4, "0");
  return `4d415445-${svc}-4700-${chr}-000000000000`;
}

export const SETTINGS_SERVICE = uuid(0x0001, 0x0000);
export const SETTINGS_DATA_CHR = uuid(0x0001, 0x0001);
export const SETTINGS_UPDATE_CHR = uuid(0x0001, 0x0002);
export const SETTINGS_RESTART_REQUIRED_CHR = uuid(0x0001, 0x0003);
export const SETTINGS_RESTART_CHR = uuid(0x0001, 0x0004);

export const WIFI_SERVICE = uuid(0x0002, 0x0000);
export const WIFI_STATUS_CHR = uuid(0x0002, 0x0001);
export const WIFI_CONNECT_CHR = uuid(0x0002, 0x0002);
export const WIFI_COMMAND_CHR = uuid(0x0002, 0x0003);
export const WIFI_STORED_CREDENTIAL_CHR = uuid(0x0002, 0x0004);

export const LOG_SERVICE = uuid(0x0003, 0x0000);
export const LOG_MESSAGE_CHR = uuid(0x0003, 0x0001);
export const LOG_ENABLED_CHR = uuid(0x0003, 0x0002);

export const SYSTEM_INFO_SERVICE = uuid(0x0004, 0x0000);
export const SYSTEM_INFO_INFO_CHR = uuid(0x0004, 0x0001);
export const SYSTEM_INFO_CONFIG_SCHEMA_CHR = uuid(0x0004, 0x0002);

export const ALL_SERVICE_UUIDS = [
  SETTINGS_SERVICE,
  WIFI_SERVICE,
  LOG_SERVICE,
  SYSTEM_INFO_SERVICE,
];

export const DEVICE_NAME_PREFIX = "matedev_";
