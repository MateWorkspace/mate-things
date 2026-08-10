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
    <form className="border-border bg-muted space-y-3 rounded-2xl border p-4">
      <div className="grid gap-3 sm:grid-cols-2 lg:grid-cols-3">
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
        <Input
          name="metric_name"
          defaultValue={props.metricName}
          placeholder="Metric name"
          aria-label="Metric name"
        />
      </div>
      <div className="border-border flex justify-end gap-2 border-t pt-3">
        <a
          href="/telemetry"
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
