"use client";

import { useState, type FormEvent } from "react";

import Button from "@/components/ui/button";
import Card from "@/components/ui/card";
import Input from "@/components/ui/input";
import Label from "@/components/ui/label";
import StatusBadge from "@/components/ui/status-badge";

import type { WifiCommand, WifiStatus, WifiStoredCredential } from "../../_lib/ble/protocol";

function InfoRow({ label, value }: { label: string; value: string | number }) {
  return (
    <div className="flex items-start justify-between gap-4">
      <dt className="text-muted-foreground">{label}</dt>
      <dd className="text-right font-medium">{value}</dd>
    </div>
  );
}

export default function WifiManagerBlock({
  status,
  stored,
  onConnect,
  onCommand,
}: {
  status: WifiStatus;
  stored: WifiStoredCredential;
  onConnect: (ssid: string, password: string) => Promise<void>;
  onCommand: (command: WifiCommand) => Promise<void>;
}) {
  const [ssid, setSsid] = useState("");
  const [password, setPassword] = useState("");
  const [connectError, setConnectError] = useState<string>();
  const [connectPending, setConnectPending] = useState(false);
  const [commandError, setCommandError] = useState<string>();
  const [pendingCommand, setPendingCommand] = useState<WifiCommand>();

  async function handleConnect(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    setConnectError(undefined);
    setConnectPending(true);
    try {
      await onConnect(ssid, password);
      setSsid("");
      setPassword("");
    } catch (error) {
      setConnectError(
        error instanceof Error ? error.message : "Failed to connect.",
      );
    } finally {
      setConnectPending(false);
    }
  }

  async function handleCommand(command: WifiCommand) {
    setCommandError(undefined);
    setPendingCommand(command);
    try {
      await onCommand(command);
    } catch (error) {
      setCommandError(
        error instanceof Error ? error.message : "Command failed.",
      );
    } finally {
      setPendingCommand(undefined);
    }
  }

  return (
    <Card>
      <h2 className="font-display text-primary text-xl tracking-wide">
        WiFi manager
      </h2>

      <div className="mt-4 flex flex-wrap items-center gap-2">
        <StatusBadge variant={status.is_up ? "success" : "neutral"}>
          {status.is_up ? "Up" : "Down"}
        </StatusBadge>
        <StatusBadge
          variant={
            status.sta_connection_status === "connected" ? "success" : "warning"
          }
        >
          {status.sta_connection_status === "connected"
            ? "Connected"
            : "Disconnected"}
        </StatusBadge>
      </div>

      <dl className="mt-4 space-y-2 text-sm">
        <InfoRow label="SSID" value={status.ssid || "—"} />
        <InfoRow label="IP" value={status.ip} />
        <InfoRow label="Netmask" value={status.netmask} />
        <InfoRow label="Gateway" value={status.gateway} />
        <InfoRow label="RSSI" value={`${status.rssi} dBm`} />
        <InfoRow
          label="Try connect on init"
          value={status.try_connect_on_init_enabled ? "Enabled" : "Disabled"}
        />
      </dl>
      <p className="text-muted-foreground mt-1 text-xs">
        Edit &quot;Try connect on init&quot; in the Settings block below.
      </p>

      <div className="mt-4 flex flex-wrap gap-2">
        <Button
          variant="secondary"
          disabled={pendingCommand !== undefined}
          onClick={() => void handleCommand("start")}
        >
          Start
        </Button>
        <Button
          variant="secondary"
          disabled={pendingCommand !== undefined}
          onClick={() => void handleCommand("stop")}
        >
          Stop
        </Button>
        <Button
          variant="secondary"
          disabled={pendingCommand !== undefined || !stored.available}
          onClick={() => void handleCommand("connect_stored")}
        >
          Connect to stored
        </Button>
        <Button
          variant="secondary"
          disabled={pendingCommand !== undefined || !status.is_up}
          onClick={() => void handleCommand("disconnect")}
        >
          Disconnect
        </Button>
        <Button
          variant="critical"
          disabled={pendingCommand !== undefined || !stored.available}
          onClick={() => void handleCommand("forget_stored")}
        >
          Forget stored
        </Button>
      </div>
      {commandError ? (
        <p className="text-critical mt-2 text-sm">{commandError}</p>
      ) : null}

      <form
        onSubmit={handleConnect}
        className="border-control-border mt-5 space-y-3 border-t pt-4"
      >
        <h3 className="text-sm font-semibold">Connect to a network</h3>
        <div>
          <Label htmlFor="wifi-ssid">SSID</Label>
          <Input
            id="wifi-ssid"
            value={ssid}
            onChange={(event) => setSsid(event.target.value)}
            required
          />
        </div>
        <div>
          <Label htmlFor="wifi-password">Password</Label>
          <Input
            id="wifi-password"
            type="password"
            value={password}
            onChange={(event) => setPassword(event.target.value)}
          />
        </div>
        {connectError ? (
          <p className="text-critical text-sm">{connectError}</p>
        ) : null}
        <Button type="submit" disabled={connectPending}>
          {connectPending ? "Connecting…" : "Connect"}
        </Button>
      </form>
    </Card>
  );
}
