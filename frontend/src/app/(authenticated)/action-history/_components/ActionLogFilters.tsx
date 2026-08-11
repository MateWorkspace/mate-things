import FilterBar from "@/components/collection/FilterBar";
import ActionSearchCombobox from "@/components/actions/ActionSearchCombobox";
import NodeSearchCombobox from "@/components/nodes/NodeSearchCombobox";
import TimeRangeFilter from "@/components/records/TimeRangeFilter";
import Input from "@/components/ui/input";
import Select from "@/components/ui/select";
import type { ActionStatus } from "@/lib/api/action-logs";

import { ACTION_STATUS_LABELS } from "../_lib/status";

interface ActionLogFiltersProps {
  canReadActions: boolean;
  canReadNodes: boolean;
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
    <FilterBar clearHref="/action-history">
      <TimeRangeFilter start={props.start} end={props.end} />
      {props.canReadActions ? (
        <ActionSearchCombobox
          name="action_id"
          defaultActionId={props.actionId}
          defaultActionName={props.actionName ?? props.actionId}
        />
      ) : (
        <label>
          <span className="text-foreground/70 mb-1.5 block text-xs font-semibold">
            Action ID
          </span>
          <Input
            aria-label="Action ID"
            name="action_id"
            readOnly
            value={props.actionId ?? ""}
          />
        </label>
      )}
      {props.canReadNodes ? (
        <NodeSearchCombobox
          name="node_id"
          defaultNodeId={props.nodeId}
          defaultNodeName={props.nodeName ?? props.nodeId}
        />
      ) : (
        <label>
          <span className="text-foreground/70 mb-1.5 block text-xs font-semibold">
            Node ID
          </span>
          <Input
            aria-label="Node ID"
            name="node_id"
            readOnly
            value={props.nodeId ?? ""}
          />
        </label>
      )}
      <div>
        <label
          className="text-foreground/70 mb-1.5 block text-xs font-semibold"
          htmlFor="action-log-status"
        >
          Status
        </label>
        <Select
          id="action-log-status"
          name="status"
          defaultValue={props.status ?? ""}
        >
          <option value="">Any status</option>
          {STATUS_KEYS.map((status) => (
            <option key={status} value={status}>
              {ACTION_STATUS_LABELS[status].label}
            </option>
          ))}
        </Select>
      </div>
    </FilterBar>
  );
}
