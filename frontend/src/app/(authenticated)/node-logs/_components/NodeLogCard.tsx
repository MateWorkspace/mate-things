import Card from "@/components/ui/card";
import StatusBadge, { type StatusVariant } from "@/components/ui/status-badge";
import type { NodeLogLevel, NodeLogResponse } from "@/lib/api/node-logs";

const LEVEL: Record<NodeLogLevel, { label: string; variant: StatusVariant }> = {
  NONE: { label: "None", variant: "neutral" },
  ERROR: { label: "Error", variant: "critical" },
  WARN: { label: "Warning", variant: "warning" },
  INFO: { label: "Info", variant: "info" },
  DEBUG: { label: "Debug", variant: "neutral" },
};

export default function NodeLogCard({ log }: { log: NodeLogResponse }) {
  const level = LEVEL[log.level];
  return (
    <Card className="h-full">
      <div className="flex items-start justify-between gap-3">
        <div>
          <h2 className="font-display text-primary text-lg">
            {log.tag || "untagged"}
          </h2>
          <p className="text-muted-foreground mt-1 font-mono text-xs">
            {log.node_device_id}
          </p>
        </div>
        <StatusBadge variant={level.variant}>{level.label}</StatusBadge>
      </div>
      <time
        className="text-muted-foreground mt-4 block text-sm"
        dateTime={log.logged_at}
      >
        {new Date(log.logged_at).toLocaleString()}
      </time>
      <details className="border-border bg-background mt-4 rounded-xl border p-3">
        <summary className="text-primary cursor-pointer text-sm font-semibold">
          Show message
        </summary>
        <p className="mt-3 text-sm break-words whitespace-pre-wrap">
          {log.message}
        </p>
      </details>
    </Card>
  );
}
