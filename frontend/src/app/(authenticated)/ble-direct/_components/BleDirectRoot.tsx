"use client";

import ConnectionBanner from "./ConnectionBanner";
import LogBlock from "./blocks/LogBlock";
import SettingsBlock from "./blocks/SettingsBlock/SettingsBlock";
import SystemInfoBlock from "./blocks/SystemInfoBlock";
import WifiManagerBlock from "./blocks/WifiManagerBlock";
import { useBleConnection } from "../_lib/useBleConnection";

export default function BleDirectRoot() {
  const {
    state,
    connect,
    disconnect,
    writeSettingsUpdate,
    restartDevice,
    writeWifiConnect,
    sendWifiCommand,
    setLogEnabled,
    clearLog,
  } = useBleConnection();

  const data =
    state.status === "connected" || state.status === "reconnecting"
      ? state.data
      : null;

  return (
    <div className="space-y-6">
      <ConnectionBanner state={state} onConnect={connect} onDisconnect={disconnect} />

      {data ? (
        <div
          className={
            state.status === "reconnecting" ? "space-y-6 opacity-60" : "space-y-6"
          }
        >
          <SystemInfoBlock info={data.systemInfo} />
          <WifiManagerBlock
            status={data.wifiStatus}
            stored={data.wifiStoredCredential}
            onConnect={writeWifiConnect}
            onCommand={sendWifiCommand}
          />
          <SettingsBlock
            schema={data.configSchema}
            snapshot={data.settingsSnapshot}
            restartRequired={data.restartRequired}
            onSave={writeSettingsUpdate}
            onRestart={() => restartDevice(0)}
          />
          <LogBlock
            lines={data.logLines}
            enabled={data.logEnabled}
            onToggle={setLogEnabled}
            onClear={clearLog}
          />
        </div>
      ) : state.status === "disconnected" ? (
        <div className="border-control-border bg-muted rounded-2xl border border-dashed p-8 text-center">
          <p className="font-semibold">No device connected</p>
          <p className="text-muted-foreground mt-1 text-sm">
            Press Connect above and choose a device advertising as
            matedev_* to view its system info, WiFi status, settings, and
            live log.
          </p>
        </div>
      ) : null}
    </div>
  );
}
