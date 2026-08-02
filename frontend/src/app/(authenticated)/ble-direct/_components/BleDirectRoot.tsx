"use client";

import ConnectionBanner from "./ConnectionBanner";
import SystemInfoBlock from "./blocks/SystemInfoBlock";
import { useBleConnection } from "../_lib/useBleConnection";

export default function BleDirectRoot() {
  const { state, connect, disconnect } = useBleConnection();

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
