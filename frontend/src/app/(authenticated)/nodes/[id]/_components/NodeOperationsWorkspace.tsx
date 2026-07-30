import Link from "next/link";

import ActionCard from "@/app/(authenticated)/actions/_components/ActionCard";
import DispatchActionDialog from "@/app/(authenticated)/actions/_components/DispatchActionDialog";
import NodeLogCard from "@/app/(authenticated)/node-logs/_components/NodeLogCard";
import TelemetryCard from "@/app/(authenticated)/telemetry/_components/TelemetryCard";
import RefreshBoundary from "@/components/refresh/RefreshBoundary";
import { EmptyState } from "@/components/ui/states";
import type { ActionResponse } from "@/lib/api/actions";
import type { NodeLogResponse } from "@/lib/api/node-logs";
import type { NodeResponse } from "@/lib/api/nodes";
import type { TelemetryRecordResponse } from "@/lib/api/telemetry";

const linkClass =
  "text-primary focus-visible:ring-focus inline-flex min-h-11 items-center rounded-xl text-sm font-semibold underline-offset-4 hover:underline focus-visible:ring-2 focus-visible:outline-none";

export function NodeActionsWorkspace({
  actions,
  node,
  canDispatch,
}: {
  actions: readonly ActionResponse[];
  node: NodeResponse;
  canDispatch: boolean;
}) {
  return (
    <div className="space-y-4">
      <div className="flex items-center justify-between gap-3">
        <p className="text-muted-foreground text-sm">
          {actions.length} compatible actions
        </p>
        <Link
          className={linkClass}
          href={`/actions?node_class_id=${encodeURIComponent(node.node_class_id)}`}
        >
          View all actions
        </Link>
      </div>
      {actions.length ? (
        <div className="grid gap-4 md:grid-cols-2 xl:grid-cols-3">
          {actions.map((action) => (
            <div key={action.id} className="space-y-2">
              <ActionCard action={action} />
              {canDispatch ? (
                <DispatchActionDialog action={action} nodes={[node]} />
              ) : null}
            </div>
          ))}
        </div>
      ) : (
        <EmptyState
          title="No compatible actions"
          description="No action definitions currently target this node class."
        />
      )}
    </div>
  );
}

export function NodeTelemetryWorkspace({
  records,
  node,
}: {
  records: readonly TelemetryRecordResponse[];
  node: NodeResponse;
}) {
  return (
    <RefreshBoundary updatedAt={records[0]?.created_at}>
      <div className="flex items-center justify-between gap-3">
        <p className="text-muted-foreground text-sm">
          {records.length} returned records
        </p>
        <Link
          className={linkClass}
          href={`/telemetry?node_device_id=${encodeURIComponent(node.device_id)}`}
        >
          View all telemetry
        </Link>
      </div>
      {records.length ? (
        <div className="grid gap-4 md:grid-cols-2 xl:grid-cols-3">
          {records.slice(0, 12).map((record) => (
            <TelemetryCard key={record.id} record={record} />
          ))}
        </div>
      ) : (
        <EmptyState
          title="No telemetry found"
          description="This node has not reported telemetry for the current scope."
        />
      )}
    </RefreshBoundary>
  );
}

export function NodeLogsWorkspace({
  logs,
  node,
}: {
  logs: readonly NodeLogResponse[];
  node: NodeResponse;
}) {
  return (
    <RefreshBoundary updatedAt={logs[0]?.created_at} intervalMs={15_000}>
      <div className="flex items-center justify-between gap-3">
        <p className="text-muted-foreground text-sm">
          {logs.length} returned records
        </p>
        <Link
          className={linkClass}
          href={`/node-logs?node_device_id=${encodeURIComponent(node.device_id)}`}
        >
          View all node logs
        </Link>
      </div>
      {logs.length ? (
        <div className="grid gap-4 md:grid-cols-2 xl:grid-cols-3">
          {logs.slice(0, 12).map((log) => (
            <NodeLogCard key={log.id} log={log} />
          ))}
        </div>
      ) : (
        <EmptyState
          title="No node logs found"
          description="This node has not published any log records."
        />
      )}
    </RefreshBoundary>
  );
}
