"use client";

import { useCallback, useEffect, useRef, useState } from "react";

export interface RefreshState {
  status: "live" | "paused" | "stale";
  updatedAt: Date;
  refresh(): void;
  pause(): void;
  resume(): void;
}

interface SmartRefreshOptions {
  intervalMs: number;
  updatedAt?: Date;
  suspended?: boolean;
  onRefresh?: () => void | Promise<void>;
}

export function useSmartRefresh({
  intervalMs,
  updatedAt,
  suspended = false,
  onRefresh,
}: SmartRefreshOptions): RefreshState {
  const [manualPause, setManualPause] = useState(false);
  const [hidden, setHidden] = useState(
    typeof document !== "undefined" && document.visibilityState === "hidden",
  );
  const [stale, setStale] = useState(false);
  const [lastUpdated, setLastUpdated] = useState(updatedAt ?? new Date());
  const running = useRef(false);
  const callback = useRef(onRefresh);
  const paused = manualPause || hidden || suspended;

  useEffect(() => {
    callback.current = onRefresh;
  }, [onRefresh]);

  const refresh = useCallback(() => {
    if (running.current) return;
    running.current = true;
    Promise.resolve(callback.current?.())
      .then(() => {
        setLastUpdated(new Date());
        setStale(false);
      })
      .catch(() => setStale(true))
      .finally(() => {
        running.current = false;
      });
  }, []);

  useEffect(() => {
    const onVisibility = () => setHidden(document.visibilityState === "hidden");
    document.addEventListener("visibilitychange", onVisibility);
    return () => document.removeEventListener("visibilitychange", onVisibility);
  }, []);

  useEffect(() => {
    if (paused || intervalMs <= 0) return;
    const timeout = window.setTimeout(refresh, intervalMs);
    return () => window.clearTimeout(timeout);
  }, [intervalMs, lastUpdated, paused, refresh, stale]);

  return {
    status: paused ? "paused" : stale ? "stale" : "live",
    updatedAt: lastUpdated,
    refresh,
    pause: () => setManualPause(true),
    resume: () => setManualPause(false),
  };
}
