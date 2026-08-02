export type WifiCommand =
  | "stop"
  | "start"
  | "connect_stored"
  | "disconnect"
  | "forget_stored";

export const WIFI_COMMAND_OPCODES: Record<WifiCommand, number> = {
  stop: 0,
  start: 1,
  connect_stored: 2,
  disconnect: 3,
  forget_stored: 4,
};

/**
 * One entry from the system_info config_schema characteristic. `type` is
 * intentionally a plain string, not a narrowed union — the firmware schema
 * can introduce new types this frontend doesn't know how to render yet,
 * and the Settings block must degrade gracefully rather than assume a
 * fixed set.
 */
export interface ConfigSchemaEntry {
  key: string;
  type: string;
}

export interface SystemInfoProject {
  project_name: string;
  project_version: string;
  name: string;
  type: string;
  firmware_version: string;
}

export interface SystemInfoChip {
  hardware_mac: string;
  model: string;
  revision: number;
  cores: number;
}

export interface SystemInfo {
  project: SystemInfoProject;
  chip: SystemInfoChip;
}

/**
 * The settings `data` characteristic's JSON. Deliberately untyped beyond
 * "a JSON object" - which keys exist is driven entirely by the
 * config_schema characteristic (see ConfigSchemaEntry), not hardcoded
 * here. Some keys (e.g. mqtt_pass) never appear directly; only a
 * `<key>_set` boolean companion does, for secrets that are write-only over
 * BLE.
 */
export type SettingsSnapshot = Record<string, unknown>;

export interface WifiStatus {
  is_up: boolean;
  sta_connection_status: "connected" | "disconnected";
  ssid: string;
  ip: string;
  netmask: string;
  gateway: string;
  rssi: number;
  try_connect_on_init_enabled: boolean;
  connect_attempted: boolean;
}

export interface WifiStoredCredential {
  available: boolean;
  ssid: string;
}
