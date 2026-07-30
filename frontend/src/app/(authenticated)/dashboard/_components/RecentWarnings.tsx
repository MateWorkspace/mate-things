import Link from "next/link";
import { ArrowRight } from "lucide-react";

import Card from "@/components/ui/card";
import StatusBadge from "@/components/ui/status-badge";
import type { NodeLogResponse } from "@/lib/api/node-logs";

interface RecentWarningsProps {
  logs?: readonly NodeLogResponse[];
}

const DATE_FORMATTER = new Intl.DateTimeFormat("en", {
  dateStyle: "medium",
  timeStyle: "short",
  timeZone: "UTC",
});

export default function RecentWarnings({ logs }: RecentWarningsProps) {
  const visibleLogs = logs?.slice(0, 6) ?? [];

  if (visibleLogs.length === 0) {
    return null;
  }

  return (
    <section aria-labelledby="recent-warnings-heading">
      <div className="flex flex-wrap items-end justify-between gap-3">
        <div>
          <h2
            id="recent-warnings-heading"
            className="font-display text-primary text-2xl tracking-wide"
          >
            Recent warnings
          </h2>
          <p className="text-muted-foreground mt-1 text-sm">
            Latest warning and error messages reported by the fleet.
          </p>
        </div>
        <Link
          href="/node-logs?level=WARN"
          className="text-primary focus-visible:ring-focus inline-flex min-h-11 items-center gap-2 rounded-lg text-sm font-semibold underline-offset-4 hover:underline focus-visible:ring-2 focus-visible:outline-none"
        >
          View node logs
          <ArrowRight aria-hidden="true" className="size-4" />
        </Link>
      </div>

      <div className="mt-4 grid gap-3">
        {visibleLogs.map((log) => {
          const params = new URLSearchParams({
            node_device_id: log.node_device_id,
            level: log.level,
          });

          return (
            <Card
              key={log.id}
              className="grid min-w-0 gap-3 p-4 sm:grid-cols-[minmax(0,1fr)_auto] sm:items-center"
            >
              <div className="min-w-0">
                <div className="flex flex-wrap items-center gap-2">
                  <StatusBadge
                    variant={log.level === "ERROR" ? "critical" : "warning"}
                  >
                    {log.level === "ERROR" ? "Error" : "Warning"}
                  </StatusBadge>
                  <span className="text-muted-foreground truncate font-mono text-xs">
                    {log.node_device_id}
                  </span>
                  {log.tag ? (
                    <span className="bg-background text-foreground/75 rounded-md px-2 py-1 text-xs">
                      {log.tag}
                    </span>
                  ) : null}
                </div>
                <p className="mt-3 text-sm break-words">{log.message}</p>
                <p className="text-muted-foreground mt-2 text-xs">
                  <time dateTime={log.logged_at}>
                    {DATE_FORMATTER.format(new Date(log.logged_at))} UTC
                  </time>
                </p>
              </div>
              <Link
                href={`/node-logs?${params.toString()}`}
                aria-label={`Inspect ${log.level.toLowerCase()} logs for ${log.node_device_id}`}
                className="text-primary focus-visible:ring-focus inline-flex min-h-11 items-center gap-2 rounded-lg text-sm font-semibold underline-offset-4 hover:underline focus-visible:ring-2 focus-visible:outline-none"
              >
                Inspect
                <ArrowRight aria-hidden="true" className="size-4" />
              </Link>
            </Card>
          );
        })}
      </div>
    </section>
  );
}
