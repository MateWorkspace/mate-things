import FilterBar from "@/components/collection/FilterBar";
import NodeDeviceIdSearchCombobox from "@/components/nodes/NodeDeviceIdSearchCombobox";
import TimeRangeFilter from "@/components/records/TimeRangeFilter";
import Input from "@/components/ui/input";
import Select from "@/components/ui/select";
import type { NodeLogLevel } from "@/lib/api/node-logs";

interface NodeLogFiltersProps {
  canReadNodes: boolean;
  start?: string;
  end?: string;
  nodeDeviceId?: string;
  nodeName?: string;
  level?: NodeLogLevel;
}

export default function NodeLogFilters(props: NodeLogFiltersProps) {
  return (
    <FilterBar clearHref="/node-logs">
      <TimeRangeFilter start={props.start} end={props.end} />
      {props.canReadNodes ? (
        <NodeDeviceIdSearchCombobox
          name="node_device_id"
          defaultDeviceId={props.nodeDeviceId}
          defaultNodeName={props.nodeName ?? props.nodeDeviceId}
        />
      ) : (
        <label>
          <span className="text-foreground/70 mb-1.5 block text-xs font-semibold">
            Node device ID
          </span>
          <Input
            aria-label="Node device ID"
            name="node_device_id"
            readOnly
            value={props.nodeDeviceId ?? ""}
          />
        </label>
      )}
      <div>
        <label
          className="text-foreground/70 mb-1.5 block text-xs font-semibold"
          htmlFor="node-log-level"
        >
          Level
        </label>
        <Select
          id="node-log-level"
          name="level"
          defaultValue={props.level ?? ""}
        >
          <option value="">All levels</option>
          <option value="NONE">None</option>
          <option value="ERROR">Error</option>
          <option value="WARN">Warning</option>
          <option value="INFO">Info</option>
          <option value="DEBUG">Debug</option>
        </Select>
      </div>
    </FilterBar>
  );
}
