import ActionSearchCombobox from "@/components/actions/ActionSearchCombobox";
import NodeSearchCombobox from "@/components/nodes/NodeSearchCombobox";
import TimeRangeFilter from "@/components/records/TimeRangeFilter";
import type { ActionStatus } from "@/lib/api/action-logs";

import { ACTION_STATUS_LABELS } from "../_lib/status";

interface ActionLogFiltersProps {
  start?: string;
  end?: string;
  actionId?: string;
  actionName?: string;
  nodeId?: string;
  nodeName?: string;
  status?: ActionStatus;
}

const STATUS_KEYS = Object.keys(ACTION_STATUS_LABELS) as ActionStatus[];

export default function ActionLogFilters(props: ActionLogFiltersProps) {
  return (
    <form className="border-border bg-muted space-y-3 rounded-2xl border p-4">
      <div className="grid gap-3 sm:grid-cols-2 lg:grid-cols-3">
        <TimeRangeFilter start={props.start} end={props.end} />
        <ActionSearchCombobox
          name="action_id"
          defaultActionId={props.actionId}
          defaultActionName={props.actionName}
        />
        <NodeSearchCombobox
          name="node_id"
          defaultNodeId={props.nodeId}
          defaultNodeName={props.nodeName}
        />
        <div>
          <label
            className="text-foreground/70 mb-1.5 block text-xs font-semibold"
            htmlFor="action-log-status"
          >
            Status
          </label>
          <select
            id="action-log-status"
            name="status"
            defaultValue={props.status ?? ""}
            className="border-control-border bg-background text-foreground focus-visible:border-focus focus-visible:ring-focus min-h-11 w-full rounded-xl border px-3.5 py-2.5 text-sm focus-visible:ring-2 focus-visible:outline-none"
          >
            <option value="">Any status</option>
            {STATUS_KEYS.map((status) => (
              <option key={status} value={status}>
                {ACTION_STATUS_LABELS[status].label}
              </option>
            ))}
          </select>
        </div>
      </div>
      <div className="border-border flex justify-end gap-2 border-t pt-3">
        <a
          href="/action-history"
          className="border-border text-primary inline-flex min-h-11 items-center rounded-xl border px-4 text-sm font-semibold"
        >
          Clear
        </a>
        <button className="bg-primary text-surface min-h-11 rounded-xl px-6 text-sm font-semibold">
          Apply
        </button>
      </div>
    </form>
  );
}
