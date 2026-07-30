"use client";

import { Pause, Play, RefreshCw } from "lucide-react";

import Button from "@/components/ui/button";
import type { RefreshState } from "@/hooks/use-smart-refresh";

interface RefreshControlProps {
  state: RefreshState;
}

function relative(date: Date): string {
  const seconds = Math.max(0, Math.round((Date.now() - date.getTime()) / 1000));
  if (seconds < 60) return `${seconds} seconds ago`;
  const minutes = Math.floor(seconds / 60);
  return `${minutes} minute${minutes === 1 ? "" : "s"} ago`;
}

export default function RefreshControl({ state }: RefreshControlProps) {
  const label =
    state.status === "live"
      ? "Live"
      : state.status === "paused"
        ? "Paused"
        : "Stale";

  return (
    <div className="flex flex-wrap items-center justify-end gap-2">
      <span aria-live="polite" className="text-muted-foreground text-sm">
        {label} · updated {relative(state.updatedAt)}
      </span>
      <Button
        type="button"
        variant="secondary"
        className="gap-2"
        onClick={state.status === "paused" ? state.resume : state.pause}
      >
        {state.status === "paused" ? (
          <Play aria-hidden="true" className="size-4" />
        ) : (
          <Pause aria-hidden="true" className="size-4" />
        )}
        {state.status === "paused" ? "Resume" : "Pause"}
      </Button>
      <Button
        type="button"
        variant="secondary"
        className="gap-2"
        onClick={state.refresh}
      >
        <RefreshCw aria-hidden="true" className="size-4" />
        Refresh
      </Button>
    </div>
  );
}
