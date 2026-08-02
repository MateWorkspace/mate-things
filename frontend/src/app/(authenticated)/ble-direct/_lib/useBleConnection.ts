"use client";

import { useCallback, useEffect, useRef, useState, useSyncExternalStore } from "react";

import { useToast } from "@/hooks/use-toast";

import { BleClient, type BleClientState } from "./ble/BleClient";
import type { WifiCommand } from "./ble/protocol";

export function useBleConnection(): {
  state: BleClientState;
  connect: () => void;
  disconnect: () => void;
  writeSettingsUpdate: (partial: Record<string, unknown>) => Promise<void>;
  restartDevice: (delayMs: number) => Promise<void>;
  writeWifiConnect: (ssid: string, password: string) => Promise<void>;
  sendWifiCommand: (command: WifiCommand) => Promise<void>;
  setLogEnabled: (enabled: boolean) => Promise<void>;
  clearLog: () => void;
} {
  const [client] = useState(() => new BleClient());
  const toast = useToast();

  useEffect(() => {
    return () => {
      client.disconnect();
    };
  }, [client]);

  const state = useSyncExternalStore(
    useCallback((listener) => client.subscribe(listener), [client]),
    useCallback(() => client.getState(), [client]),
    useCallback(() => client.getState(), [client]),
  );

  const previousStateRef = useRef(state);
  useEffect(() => {
    const previous = previousStateRef.current;
    if (previous.status === "reconnecting" && state.status === "connected") {
      if (previous.reason === "restarting") {
        toast.success("Device restarted", "Device restarted successfully.");
      } else {
        toast.success("Reconnected", "The device connection was restored.");
      }
    }
    previousStateRef.current = state;
  }, [state, toast]);

  return {
    state,
    connect: useCallback(() => {
      void client.connect();
    }, [client]),
    disconnect: useCallback(() => {
      client.disconnect();
    }, [client]),
    writeSettingsUpdate: useCallback(
      (partial: Record<string, unknown>) => client.writeSettingsUpdate(partial),
      [client],
    ),
    restartDevice: useCallback(
      (delayMs: number) => client.restartDevice(delayMs),
      [client],
    ),
    writeWifiConnect: useCallback(
      (ssid: string, password: string) => client.writeWifiConnect(ssid, password),
      [client],
    ),
    sendWifiCommand: useCallback(
      (command: WifiCommand) => client.sendWifiCommand(command),
      [client],
    ),
    setLogEnabled: useCallback(
      (enabled: boolean) => client.setLogEnabled(enabled),
      [client],
    ),
    clearLog: useCallback(() => client.clearLog(), [client]),
  };
}
