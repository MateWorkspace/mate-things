"use client";

import { useCallback, useEffect, useState, useTransition } from "react";
import { useRouter } from "next/navigation";
import { Pause, Play, RefreshCw } from "lucide-react";

import Button from "@/components/ui/button";
import LocalDateTime from "@/components/ui/local-date-time";

const AUTO_REFRESH_INTERVAL_MS = 30_000;

interface DashboardRefreshProps {
  loadedAt: string;
}

export default function DashboardRefresh({ loadedAt }: DashboardRefreshProps) {
  const router = useRouter();
  const [paused, setPaused] = useState(false);
  const [documentHidden, setDocumentHidden] = useState(false);
  const [isPending, startTransition] = useTransition();

  const refresh = useCallback(() => {
    if (isPending) {
      return;
    }

    startTransition(() => router.refresh());
  }, [isPending, router]);

  useEffect(() => {
    const handleVisibility = () => setDocumentHidden(document.hidden);
    handleVisibility();
    document.addEventListener("visibilitychange", handleVisibility);
    return () =>
      document.removeEventListener("visibilitychange", handleVisibility);
  }, []);

  useEffect(() => {
    if (paused || documentHidden) {
      return;
    }

    const timer = window.setInterval(refresh, AUTO_REFRESH_INTERVAL_MS);
    return () => window.clearInterval(timer);
  }, [documentHidden, paused, refresh]);

  const live = !paused && !documentHidden;

  return (
    <div className="flex flex-wrap items-center justify-end gap-2">
      <p
        className="text-muted-foreground w-full text-left text-xs sm:w-auto sm:text-right"
        aria-live="polite"
      >
        {isPending
          ? "Refreshing fleet data…"
          : live
            ? "Live refresh on"
            : "Auto-refresh paused"}
        {" · "}
        Updated <LocalDateTime value={loadedAt} variant="time" />
      </p>
      <Button
        type="button"
        variant="secondary"
        className="min-h-11 gap-2"
        onClick={() => setPaused((current) => !current)}
      >
        {paused ? (
          <Play aria-hidden="true" className="size-4" />
        ) : (
          <Pause aria-hidden="true" className="size-4" />
        )}
        {paused ? "Resume" : "Pause"}
      </Button>
      <Button
        type="button"
        className="min-h-11 gap-2"
        disabled={isPending}
        onClick={refresh}
      >
        <RefreshCw
          aria-hidden="true"
          className={`size-4 ${isPending ? "animate-spin" : ""}`}
        />
        Refresh
      </Button>
    </div>
  );
}
