"use client";

import { useEffect, useRef, useState } from "react";

import Button from "@/components/ui/button";
import Card from "@/components/ui/card";

export default function LogBlock({
  lines,
  enabled,
  onToggle,
  onClear,
}: {
  lines: readonly string[];
  enabled: boolean;
  onToggle: (enabled: boolean) => Promise<void>;
  onClear: () => void;
}) {
  const containerRef = useRef<HTMLDivElement>(null);
  const [autoScroll, setAutoScroll] = useState(true);
  const [toggleError, setToggleError] = useState<string>();
  const [togglePending, setTogglePending] = useState(false);

  useEffect(() => {
    if (!autoScroll) return;
    const container = containerRef.current;
    if (!container) return;
    container.scrollTop = container.scrollHeight;
  }, [lines, autoScroll]);

  function handleScroll() {
    const container = containerRef.current;
    if (!container) return;
    const distanceFromBottom =
      container.scrollHeight - container.scrollTop - container.clientHeight;
    setAutoScroll(distanceFromBottom < 24);
  }

  async function handleToggle() {
    setToggleError(undefined);
    setTogglePending(true);
    try {
      await onToggle(!enabled);
    } catch (error) {
      setToggleError(
        error instanceof Error ? error.message : "Failed to update logging.",
      );
    } finally {
      setTogglePending(false);
    }
  }

  return (
    <Card>
      <div className="flex flex-wrap items-center justify-between gap-3">
        <h2 className="font-display text-primary text-xl tracking-wide">Log</h2>
        <div className="flex items-center gap-2">
          <Button
            variant="secondary"
            disabled={togglePending}
            onClick={() => void handleToggle()}
          >
            {enabled ? "Disable streaming" : "Enable streaming"}
          </Button>
          <Button variant="secondary" onClick={onClear}>
            Clear
          </Button>
        </div>
      </div>
      {toggleError ? (
        <p className="text-critical mt-2 text-sm">{toggleError}</p>
      ) : null}

      <div className="relative mt-4">
        <div
          ref={containerRef}
          onScroll={handleScroll}
          className="border-control-border bg-ink text-surface h-64 overflow-y-auto rounded-xl border p-3 font-mono text-xs leading-5"
        >
          {lines.length === 0 ? (
            <p className="text-surface/60">No log lines yet.</p>
          ) : (
            lines.map((line, index) => <div key={index}>{line}</div>)
          )}
        </div>
        {!autoScroll ? (
          <Button
            variant="secondary"
            className="absolute right-3 bottom-3"
            onClick={() => {
              setAutoScroll(true);
              const container = containerRef.current;
              if (container) container.scrollTop = container.scrollHeight;
            }}
          >
            Jump to latest
          </Button>
        ) : null}
      </div>
    </Card>
  );
}
