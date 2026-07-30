import Link from "next/link";
import { ArrowRight, CircleX, RadioTower } from "lucide-react";

import Card from "@/components/ui/card";
import StatusBadge from "@/components/ui/status-badge";
import type { ActionLogResponse } from "@/lib/api/action-logs";
import type { NodeResponse } from "@/lib/api/nodes";

interface UrgentAttentionProps {
  disconnectedNodes?: readonly NodeResponse[];
  failedActions?: readonly ActionLogResponse[];
}

const DATE_FORMATTER = new Intl.DateTimeFormat("en", {
  dateStyle: "medium",
  timeStyle: "short",
  timeZone: "UTC",
});

function actionHref(action: ActionLogResponse): string {
  const params = new URLSearchParams({
    execution_id: action.execution_id,
    start: action.executed_at,
  });
  return `/action-history?${params.toString()}`;
}

export default function UrgentAttention({
  disconnectedNodes,
  failedActions,
}: UrgentAttentionProps) {
  const nodes = disconnectedNodes?.slice(0, 4) ?? [];
  const actions = failedActions?.slice(0, 4) ?? [];

  if (nodes.length === 0 && actions.length === 0) {
    return null;
  }

  return (
    <section aria-labelledby="urgent-attention-heading">
      <div className="flex flex-wrap items-end justify-between gap-3">
        <div>
          <h2
            id="urgent-attention-heading"
            className="font-display text-primary text-2xl tracking-wide"
          >
            Urgent attention
          </h2>
          <p className="text-muted-foreground mt-1 text-sm">
            Highest-signal issues from the current dashboard window.
          </p>
        </div>
      </div>

      <div className="mt-4 grid gap-4 lg:grid-cols-2">
        {nodes.map((node) => (
          <Card key={node.id} className="flex min-w-0 flex-col gap-4 p-4">
            <div className="flex items-start justify-between gap-3">
              <span className="bg-background text-critical flex size-10 shrink-0 items-center justify-center rounded-xl">
                <RadioTower aria-hidden="true" className="size-5" />
              </span>
              <StatusBadge variant="critical">Disconnected</StatusBadge>
            </div>
            <div className="min-w-0">
              <h3 className="truncate font-semibold">{node.name}</h3>
              <p className="text-muted-foreground mt-1 truncate font-mono text-xs">
                {node.device_id}
              </p>
              <p className="text-foreground/70 mt-2 line-clamp-2 text-sm">
                {node.description || "This node is not currently connected."}
              </p>
            </div>
            <Link
              href={`/nodes/${node.id}`}
              className="text-primary focus-visible:ring-focus mt-auto inline-flex min-h-11 w-fit items-center gap-2 rounded-lg text-sm font-semibold underline-offset-4 hover:underline focus-visible:ring-2 focus-visible:outline-none"
            >
              Inspect node
              <ArrowRight aria-hidden="true" className="size-4" />
            </Link>
          </Card>
        ))}

        {actions.map((action) => (
          <Card key={action.id} className="flex min-w-0 flex-col gap-4 p-4">
            <div className="flex items-start justify-between gap-3">
              <span className="bg-background text-critical flex size-10 shrink-0 items-center justify-center rounded-xl">
                <CircleX aria-hidden="true" className="size-5" />
              </span>
              <StatusBadge
                variant={
                  action.action_status === "FAILED" ? "critical" : "warning"
                }
              >
                {action.action_status === "FAILED" ? "Failed" : "Unresponded"}
              </StatusBadge>
            </div>
            <div className="min-w-0">
              <h3 className="truncate font-semibold">
                Action {action.action_id}
              </h3>
              <p className="text-muted-foreground mt-1 text-xs">
                <time dateTime={action.executed_at}>
                  {DATE_FORMATTER.format(new Date(action.executed_at))} UTC
                </time>
              </p>
              <p className="text-foreground/70 mt-2 line-clamp-2 text-sm">
                {action.action_message || "No response message was recorded."}
              </p>
            </div>
            <Link
              href={actionHref(action)}
              className="text-primary focus-visible:ring-focus mt-auto inline-flex min-h-11 w-fit items-center gap-2 rounded-lg text-sm font-semibold underline-offset-4 hover:underline focus-visible:ring-2 focus-visible:outline-none"
            >
              Review action history
              <ArrowRight aria-hidden="true" className="size-4" />
            </Link>
          </Card>
        ))}
      </div>
    </section>
  );
}
