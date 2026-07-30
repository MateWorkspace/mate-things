import TimeRangeFilter from "@/components/records/TimeRangeFilter";
import Input from "@/components/ui/input";
import type { NodeLogLevel } from "@/lib/api/node-logs";

interface NodeLogFiltersProps {
  start?: string;
  end?: string;
  nodeDeviceId?: string;
  level?: NodeLogLevel;
  immutableDevice?: boolean;
}

export default function NodeLogFilters(props: NodeLogFiltersProps) {
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
      <select
        name="level"
        defaultValue={props.level ?? ""}
        aria-label="Log level"
        className="border-control-border bg-background focus-visible:ring-focus min-h-11 rounded-xl border px-3.5 text-sm focus-visible:ring-2 focus-visible:outline-none"
      >
        <option value="">All levels</option>
        <option value="NONE">None</option>
        <option value="ERROR">Error</option>
        <option value="WARN">Warning</option>
        <option value="INFO">Info</option>
        <option value="DEBUG">Debug</option>
      </select>
      <div className="flex gap-2">
        <button className="bg-primary text-surface min-h-11 flex-1 rounded-xl px-4 text-sm font-semibold">
          Apply filters
        </button>
        {!props.immutableDevice ? (
          <a
            href="/node-logs"
            className="border-border text-primary inline-flex min-h-11 items-center rounded-xl border px-4 text-sm font-semibold"
          >
            Clear
          </a>
        ) : null}
      </div>
    </form>
  );
}
