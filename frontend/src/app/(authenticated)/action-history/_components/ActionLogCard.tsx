import Card from "@/components/ui/card";
import JsonPayload from "@/components/records/JsonPayload";
import StatusBadge, { type StatusVariant } from "@/components/ui/status-badge";
import type { ActionLogResponse, ActionStatus } from "@/lib/api/action-logs";

const STATUS: Record<ActionStatus, { label: string; variant: StatusVariant }> =
  {
    UNEXECUTED: { label: "Unexecuted", variant: "neutral" },
    UNRESPONDED: { label: "Unresponded", variant: "warning" },
    FAILED: { label: "Failed", variant: "critical" },
    SUCCESS: { label: "Success", variant: "success" },
  };

export function statusLabel(status: ActionStatus): string {
  return STATUS[status].label;
}

export default function ActionLogCard({ log }: { log: ActionLogResponse }) {
  const status = STATUS[log.action_status];
  return (
    <Card className="h-full">
      <div className="flex items-start justify-between gap-3">
        <div>
          <p className="font-display text-primary text-lg">Execution</p>
          <p className="text-muted-foreground mt-1 font-mono text-xs break-all">
            {log.execution_id}
          </p>
        </div>
        <StatusBadge variant={status.variant}>{status.label}</StatusBadge>
      </div>
      <dl className="mt-4 space-y-2 text-sm">
        <div className="flex justify-between gap-3">
          <dt className="text-muted-foreground">Executed</dt>
          <dd>
            <time dateTime={log.executed_at}>
              {new Date(log.executed_at).toLocaleString()}
            </time>
          </dd>
        </div>
        <div className="flex justify-between gap-3">
          <dt className="text-muted-foreground">Action</dt>
          <dd className="max-w-48 truncate font-mono text-xs">
            {log.action_id}
          </dd>
        </div>
        <div className="flex justify-between gap-3">
          <dt className="text-muted-foreground">Node</dt>
          <dd className="max-w-48 truncate font-mono text-xs">
            {log.node_id ?? "Not assigned"}
          </dd>
        </div>
      </dl>
      {log.action_message ? (
        <details className="mt-4">
          <summary className="text-primary cursor-pointer text-sm font-semibold">
            Show message
          </summary>
          <p className="bg-muted mt-2 rounded-xl p-3 text-sm whitespace-pre-wrap">
            {log.action_message}
          </p>
        </details>
      ) : null}
      <div className="mt-4">
        <JsonPayload value={log.payload} />
      </div>
    </Card>
  );
}
