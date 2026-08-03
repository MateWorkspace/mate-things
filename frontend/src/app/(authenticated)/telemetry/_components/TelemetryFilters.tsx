import TimeRangeFilter from "@/components/records/TimeRangeFilter";
import Input from "@/components/ui/input";

interface TelemetryFiltersProps {
  start?: string;
  end?: string;
  nodeDeviceId?: string;
  metricName?: string;
  schemaName?: string;
  schemaVersion?: number;
  immutableDevice?: boolean;
}

export default function TelemetryFilters(props: TelemetryFiltersProps) {
  return (
    <form className="border-border bg-muted grid gap-3 rounded-2xl border p-4 sm:grid-cols-2 lg:grid-cols-3">
      <TimeRangeFilter start={props.start} end={props.end} />
      <Input
        name="node_device_id"
        defaultValue={props.nodeDeviceId}
        placeholder="Node device ID"
        aria-label="Node device ID"
        readOnly={props.immutableDevice}
      />
      <Input
        name="metric_name"
        defaultValue={props.metricName}
        placeholder="Metric name"
        aria-label="Metric name"
      />
      <Input
        name="payload_schema_name"
        defaultValue={props.schemaName}
        placeholder="Schema name"
        aria-label="Schema name"
      />
      <Input
        name="payload_schema_version"
        type="number"
        min={1}
        defaultValue={props.schemaVersion}
        placeholder="Schema version"
        aria-label="Schema version"
      />
      <div className="flex gap-2">
        <button className="bg-primary text-surface min-h-11 flex-1 rounded-xl px-4 text-sm font-semibold">
          Apply
        </button>
        {!props.immutableDevice ? (
          <a
            href="/telemetry"
            className="border-border text-primary inline-flex min-h-11 items-center rounded-xl border px-4 text-sm font-semibold"
          >
            Clear
          </a>
        ) : null}
      </div>
    </form>
  );
}
