"use client";

import Button from "@/components/ui/button";

import type { BleClientState } from "../_lib/ble/BleClient";

export default function ConnectionBanner({
  state,
  onConnect,
  onDisconnect,
}: {
  state: BleClientState;
  onConnect: () => void;
  onDisconnect: () => void;
}) {
  if (state.status === "unsupported") {
    return (
      <div className="border-critical bg-critical/10 text-critical rounded-xl border p-4 text-sm">
        Web Bluetooth isn&apos;t available in this browser or context. Use a
        Chromium-based browser (Chrome, Edge) over HTTPS or localhost.
      </div>
    );
  }

  if (state.status === "disconnected") {
    return (
      <div className="border-control-border bg-muted flex flex-wrap items-center justify-between gap-3 rounded-xl border p-4">
        <div>
          <p className="font-semibold">Not connected</p>
          {state.error ? (
            <p className="text-critical mt-1 text-sm">{state.error}</p>
          ) : (
            <p className="text-muted-foreground mt-1 text-sm">
              Connect to a device advertising as matedev_* to get started.
            </p>
          )}
        </div>
        <Button onClick={onConnect}>Connect</Button>
      </div>
    );
  }

  if (state.status === "requesting" || state.status === "connecting") {
    return (
      <div className="border-control-border bg-muted rounded-xl border p-4">
        <p className="font-semibold">
          {state.status === "requesting" ? "Choose a device…" : "Connecting…"}
        </p>
      </div>
    );
  }

  if (state.status === "reconnecting") {
    const label =
      state.reason === "restarting"
        ? "Device is restarting — reconnecting…"
        : "Connection lost — reconnecting…";
    const tone =
      state.reason === "restarting"
        ? "border-control-border bg-muted"
        : "border-warning bg-warning/10";
    return (
      <div
        className={`flex flex-wrap items-center justify-between gap-3 rounded-xl border p-4 ${tone}`}
      >
        <p className="font-semibold">{label}</p>
        <Button variant="secondary" onClick={onDisconnect}>
          Disconnect
        </Button>
      </div>
    );
  }

  return (
    <div className="border-success bg-success/10 flex flex-wrap items-center justify-between gap-3 rounded-xl border p-4">
      <p className="font-semibold">Connected</p>
      <Button variant="secondary" onClick={onDisconnect}>
        Disconnect
      </Button>
    </div>
  );
}
