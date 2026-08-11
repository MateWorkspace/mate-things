import FilterBar from "@/components/collection/FilterBar";
import NodeDeviceIdSearchCombobox from "@/components/nodes/NodeDeviceIdSearchCombobox";
import TimeRangeFilter from "@/components/records/TimeRangeFilter";
import Input from "@/components/ui/input";

interface TelemetryFiltersProps {
  canReadNodes: boolean;
  start?: string;
  end?: string;
  nodeDeviceId?: string;
  nodeName?: string;
  metricName?: string;
}

export default function TelemetryFilters(props: TelemetryFiltersProps) {
  return (
    <FilterBar clearHref="/telemetry">
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
      <label>
        <span className="text-foreground/70 mb-1.5 block text-xs font-semibold">
          Metric name
        </span>
        <Input
          name="metric_name"
          defaultValue={props.metricName}
          placeholder="Metric name"
        />
      </label>
    </FilterBar>
  );
}
