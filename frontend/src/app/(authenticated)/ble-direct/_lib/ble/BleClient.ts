import {
  decodeBoolByte,
  decodeUtf8Json,
  decodeUtf8Text,
  encodeBoolByte,
  encodeUint32LE,
  encodeUtf8Json,
} from "./codec";
import {
  WIFI_COMMAND_OPCODES,
  type ConfigSchemaEntry,
  type SettingsSnapshot,
  type SystemInfo,
  type WifiCommand,
  type WifiStatus,
  type WifiStoredCredential,
} from "./protocol";
import {
  ALL_SERVICE_UUIDS,
  DEVICE_NAME_PREFIX,
  LOG_ENABLED_CHR,
  LOG_MESSAGE_CHR,
  LOG_SERVICE,
  SETTINGS_DATA_CHR,
  SETTINGS_RESTART_CHR,
  SETTINGS_RESTART_REQUIRED_CHR,
  SETTINGS_SERVICE,
  SETTINGS_UPDATE_CHR,
  SYSTEM_INFO_CONFIG_SCHEMA_CHR,
  SYSTEM_INFO_INFO_CHR,
  SYSTEM_INFO_SERVICE,
  WIFI_COMMAND_CHR,
  WIFI_CONNECT_CHR,
  WIFI_SERVICE,
  WIFI_STATUS_CHR,
  WIFI_STORED_CREDENTIAL_CHR,
} from "./uuid";

const RECONNECT_INTERVAL_MS = 3000;
const OPERATION_TIMEOUT_MS = 10000;
const RESTART_GRACE_MS = 5000;
const LOG_BUFFER_MAX = 500;

export type DisconnectReason = "lost" | "restarting";

export interface ConnectedData {
  systemInfo: SystemInfo;
  configSchema: ConfigSchemaEntry[];
  settingsSnapshot: SettingsSnapshot;
  restartRequired: boolean;
  wifiStatus: WifiStatus;
  wifiStoredCredential: WifiStoredCredential;
  logLines: string[];
  logEnabled: boolean;
}

export type BleClientState =
  | { status: "unsupported" }
  | { status: "disconnected"; error?: string }
  | { status: "requesting" }
  | { status: "connecting" }
  | { status: "connected"; data: ConnectedData }
  | { status: "reconnecting"; data: ConnectedData; reason: DisconnectReason };

interface CharacteristicRefs {
  settingsData: BluetoothRemoteGATTCharacteristic;
  settingsUpdate: BluetoothRemoteGATTCharacteristic;
  settingsRestartRequired: BluetoothRemoteGATTCharacteristic;
  settingsRestart: BluetoothRemoteGATTCharacteristic;
  wifiStatus: BluetoothRemoteGATTCharacteristic;
  wifiConnect: BluetoothRemoteGATTCharacteristic;
  wifiCommand: BluetoothRemoteGATTCharacteristic;
  wifiStoredCredential: BluetoothRemoteGATTCharacteristic;
  logMessage: BluetoothRemoteGATTCharacteristic;
  logEnabled: BluetoothRemoteGATTCharacteristic;
  systemInfoInfo: BluetoothRemoteGATTCharacteristic;
  systemInfoConfigSchema: BluetoothRemoteGATTCharacteristic;
}

function withTimeout<T>(
  promise: Promise<T>,
  ms: number,
  message: string,
): Promise<T> {
  return new Promise<T>((resolve, reject) => {
    const timer = setTimeout(() => reject(new Error(message)), ms);
    promise.then(
      (value) => {
        clearTimeout(timer);
        resolve(value);
      },
      (error: unknown) => {
        clearTimeout(timer);
        reject(error instanceof Error ? error : new Error(String(error)));
      },
    );
  });
}

async function readJson<T>(
  characteristic: BluetoothRemoteGATTCharacteristic,
): Promise<T> {
  const view = await characteristic.readValue();
  return decodeUtf8Json<T>(view);
}

async function readBool(
  characteristic: BluetoothRemoteGATTCharacteristic,
): Promise<boolean> {
  const view = await characteristic.readValue();
  return decodeBoolByte(view);
}

async function writeJson(
  characteristic: BluetoothRemoteGATTCharacteristic,
  value: unknown,
): Promise<void> {
  // writeValueWithResponse (never writeValueWithoutResponse) is required:
  // it maps to the ATT "Write Request", which the platform Bluetooth stack
  // automatically splits into the "Write Long Characteristic Values"
  // prepare/execute sub-procedure when the payload exceeds the negotiated
  // MTU. writeValueWithoutResponse ("Write Command") has no such fallback
  // and would silently truncate a settings-update payload over MTU.
  await characteristic.writeValueWithResponse(encodeUtf8Json(value));
}

async function writeBool(
  characteristic: BluetoothRemoteGATTCharacteristic,
  value: boolean,
): Promise<void> {
  await characteristic.writeValueWithResponse(encodeBoolByte(value));
}

export class BleClient {
  private state: BleClientState =
    typeof navigator !== "undefined" && "bluetooth" in navigator
      ? { status: "disconnected" }
      : { status: "unsupported" };
  private listeners = new Set<() => void>();
  private device: BluetoothDevice | null = null;
  private chars: CharacteristicRefs | null = null;
  private reconnectTimer: ReturnType<typeof setInterval> | null = null;
  private reconnecting = false;
  private expectingRestart = false;
  private expectingRestartTimer: ReturnType<typeof setTimeout> | null = null;

  getState(): BleClientState {
    return this.state;
  }

  subscribe(listener: () => void): () => void {
    this.listeners.add(listener);
    return () => {
      this.listeners.delete(listener);
    };
  }

  async connect(): Promise<void> {
    if (this.state.status === "unsupported") return;

    this.setState({ status: "requesting" });

    let device: BluetoothDevice;
    try {
      device = await navigator.bluetooth!.requestDevice({
        filters: [{ namePrefix: DEVICE_NAME_PREFIX }],
        optionalServices: ALL_SERVICE_UUIDS,
      });
    } catch {
      this.setState({ status: "disconnected" });
      return;
    }

    this.device = device;
    device.addEventListener("gattserverdisconnected", this.handleDisconnected);

    await this.performConnect(false);
  }

  disconnect(): void {
    this.stopReconnectLoop();
    this.clearRestartExpectation();
    if (this.device) {
      this.device.removeEventListener(
        "gattserverdisconnected",
        this.handleDisconnected,
      );
      if (this.device.gatt?.connected) {
        this.device.gatt.disconnect();
      }
    }
    this.device = null;
    this.chars = null;
    this.setState({ status: "disconnected" });
  }

  async writeSettingsUpdate(partial: Record<string, unknown>): Promise<void> {
    if (!this.chars) throw new Error("Not connected.");
    await writeJson(this.chars.settingsUpdate, partial);
    const settingsSnapshot = await readJson<SettingsSnapshot>(
      this.chars.settingsData,
    );
    this.updateData((data) => ({ ...data, settingsSnapshot }));
  }

  async restartDevice(delayMs: number): Promise<void> {
    if (!this.chars) throw new Error("Not connected.");
    await this.chars.settingsRestart.writeValueWithResponse(
      encodeUint32LE(delayMs),
    );
    this.expectingRestart = true;
    if (this.expectingRestartTimer) {
      clearTimeout(this.expectingRestartTimer);
    }
    // Safety net: if no disconnect follows within delay_ms + slack, the
    // write likely didn't trigger a reboot - don't let a later, unrelated
    // drop get mislabeled as "restarting".
    this.expectingRestartTimer = setTimeout(() => {
      this.expectingRestart = false;
      this.expectingRestartTimer = null;
    }, delayMs + RESTART_GRACE_MS);
  }

  async writeWifiConnect(ssid: string, password: string): Promise<void> {
    if (!this.chars) throw new Error("Not connected.");
    await writeJson(this.chars.wifiConnect, { ssid, password });
  }

  async sendWifiCommand(command: WifiCommand): Promise<void> {
    if (!this.chars) throw new Error("Not connected.");
    await this.chars.wifiCommand.writeValueWithResponse(
      new Uint8Array([WIFI_COMMAND_OPCODES[command]]),
    );
  }

  async setLogEnabled(enabled: boolean): Promise<void> {
    if (!this.chars) throw new Error("Not connected.");
    await writeBool(this.chars.logEnabled, enabled);
    this.updateData((data) => ({ ...data, logEnabled: enabled }));
  }

  clearLog(): void {
    this.updateData((data) => ({ ...data, logLines: [] }));
  }

  private setState(next: BleClientState): void {
    this.state = next;
    for (const listener of this.listeners) {
      listener();
    }
  }

  private updateData(updater: (data: ConnectedData) => ConnectedData): void {
    if (this.state.status === "connected") {
      this.setState({ status: "connected", data: updater(this.state.data) });
    } else if (this.state.status === "reconnecting") {
      this.setState({
        status: "reconnecting",
        data: updater(this.state.data),
        reason: this.state.reason,
      });
    }
  }

  private clearRestartExpectation(): void {
    this.expectingRestart = false;
    if (this.expectingRestartTimer) {
      clearTimeout(this.expectingRestartTimer);
      this.expectingRestartTimer = null;
    }
  }

  private async performConnect(isReconnect: boolean): Promise<void> {
    const device = this.device;
    if (!device) return;

    if (!isReconnect) {
      this.setState({ status: "connecting" });
    }

    try {
      const server = await withTimeout(
        device.gatt!.connect(),
        OPERATION_TIMEOUT_MS,
        "Timed out connecting to device.",
      );

      const settingsService = await withTimeout(
        server.getPrimaryService(SETTINGS_SERVICE),
        OPERATION_TIMEOUT_MS,
        "Settings service not found on this device.",
      );
      const wifiService = await withTimeout(
        server.getPrimaryService(WIFI_SERVICE),
        OPERATION_TIMEOUT_MS,
        "WiFi manager service not found on this device.",
      );
      const logService = await withTimeout(
        server.getPrimaryService(LOG_SERVICE),
        OPERATION_TIMEOUT_MS,
        "Log service not found on this device.",
      );
      const systemInfoService = await withTimeout(
        server.getPrimaryService(SYSTEM_INFO_SERVICE),
        OPERATION_TIMEOUT_MS,
        "System info service not found on this device.",
      );

      const chars: CharacteristicRefs = {
        settingsData: await settingsService.getCharacteristic(SETTINGS_DATA_CHR),
        settingsUpdate: await settingsService.getCharacteristic(SETTINGS_UPDATE_CHR),
        settingsRestartRequired: await settingsService.getCharacteristic(SETTINGS_RESTART_REQUIRED_CHR),
        settingsRestart: await settingsService.getCharacteristic(SETTINGS_RESTART_CHR),
        wifiStatus: await wifiService.getCharacteristic(WIFI_STATUS_CHR),
        wifiConnect: await wifiService.getCharacteristic(WIFI_CONNECT_CHR),
        wifiCommand: await wifiService.getCharacteristic(WIFI_COMMAND_CHR),
        wifiStoredCredential: await wifiService.getCharacteristic(WIFI_STORED_CREDENTIAL_CHR),
        logMessage: await logService.getCharacteristic(LOG_MESSAGE_CHR),
        logEnabled: await logService.getCharacteristic(LOG_ENABLED_CHR),
        systemInfoInfo: await systemInfoService.getCharacteristic(SYSTEM_INFO_INFO_CHR),
        systemInfoConfigSchema: await systemInfoService.getCharacteristic(SYSTEM_INFO_CONFIG_SCHEMA_CHR),
      };

      const [
        systemInfo,
        configSchema,
        settingsSnapshot,
        restartRequired,
        wifiStatus,
        wifiStoredCredential,
      ] = await withTimeout(
        Promise.all([
          readJson<SystemInfo>(chars.systemInfoInfo),
          readJson<ConfigSchemaEntry[]>(chars.systemInfoConfigSchema),
          readJson<SettingsSnapshot>(chars.settingsData),
          readBool(chars.settingsRestartRequired),
          readJson<WifiStatus>(chars.wifiStatus),
          readJson<WifiStoredCredential>(chars.wifiStoredCredential),
        ]),
        OPERATION_TIMEOUT_MS,
        "Timed out reading device data.",
      );

      await writeBool(chars.logEnabled, true);

      chars.settingsRestartRequired.addEventListener(
        "characteristicvaluechanged",
        this.handleRestartRequiredChanged,
      );
      chars.wifiStatus.addEventListener(
        "characteristicvaluechanged",
        this.handleWifiStatusChanged,
      );
      chars.logMessage.addEventListener(
        "characteristicvaluechanged",
        this.handleLogMessageChanged,
      );
      await chars.settingsRestartRequired.startNotifications();
      await chars.wifiStatus.startNotifications();
      await chars.logMessage.startNotifications();

      this.chars = chars;

      const data: ConnectedData = {
        systemInfo,
        configSchema,
        settingsSnapshot,
        restartRequired,
        wifiStatus,
        wifiStoredCredential,
        logLines: [],
        logEnabled: true,
      };
      this.setState({ status: "connected", data });
    } catch (error) {
      if (isReconnect) {
        // Swallow - the reconnect loop retries on its next tick. Keep
        // whatever "reconnecting" state is already showing (stale data +
        // reason); don't touch this.device so the same paired
        // BluetoothDevice keeps being retried without a new chooser
        // prompt.
        if (device.gatt?.connected) {
          device.gatt.disconnect();
        }
        return;
      }

      device.removeEventListener(
        "gattserverdisconnected",
        this.handleDisconnected,
      );
      if (device.gatt?.connected) {
        device.gatt.disconnect();
      }
      this.device = null;
      this.chars = null;
      const message =
        error instanceof Error ? error.message : "Failed to connect to device.";
      this.setState({ status: "disconnected", error: message });
    }
  }

  private startReconnectLoop(): void {
    if (this.reconnectTimer) return;
    this.reconnectTimer = setInterval(() => {
      void this.attemptReconnect();
    }, RECONNECT_INTERVAL_MS);
  }

  private stopReconnectLoop(): void {
    if (this.reconnectTimer) {
      clearInterval(this.reconnectTimer);
      this.reconnectTimer = null;
    }
  }

  private async attemptReconnect(): Promise<void> {
    if (this.reconnecting || !this.device) return;
    this.reconnecting = true;
    try {
      await this.performConnect(true);
      if (this.state.status === "connected") {
        this.stopReconnectLoop();
      }
    } finally {
      this.reconnecting = false;
    }
  }

  private readonly handleDisconnected = (): void => {
    this.stopReconnectLoop();
    const previous = this.state;
    const data =
      previous.status === "connected" || previous.status === "reconnecting"
        ? previous.data
        : null;
    if (!data) {
      // Disconnected before we ever reached "connected" - the in-flight
      // performConnect()'s own try/catch will handle this via a failed
      // GATT operation.
      return;
    }

    const reason: DisconnectReason = this.expectingRestart
      ? "restarting"
      : "lost";
    this.clearRestartExpectation();

    this.chars = null;
    this.setState({ status: "reconnecting", data, reason });
    this.startReconnectLoop();
  };

  private readonly handleRestartRequiredChanged = (event: Event): void => {
    const characteristic = event.target as BluetoothRemoteGATTCharacteristic;
    if (!characteristic.value) return;
    const restartRequired = decodeBoolByte(characteristic.value);
    this.updateData((data) => ({ ...data, restartRequired }));
  };

  private readonly handleWifiStatusChanged = (event: Event): void => {
    const characteristic = event.target as BluetoothRemoteGATTCharacteristic;
    if (!characteristic.value) return;
    const wifiStatus = decodeUtf8Json<WifiStatus>(characteristic.value);
    this.updateData((data) => ({ ...data, wifiStatus }));
  };

  private readonly handleLogMessageChanged = (event: Event): void => {
    const characteristic = event.target as BluetoothRemoteGATTCharacteristic;
    if (!characteristic.value) return;
    const line = decodeUtf8Text(characteristic.value);
    if (!line) return;
    this.updateData((data) => {
      const logLines = [...data.logLines, line];
      if (logLines.length > LOG_BUFFER_MAX) {
        logLines.splice(0, logLines.length - LOG_BUFFER_MAX);
      }
      return { ...data, logLines };
    });
  };
}
