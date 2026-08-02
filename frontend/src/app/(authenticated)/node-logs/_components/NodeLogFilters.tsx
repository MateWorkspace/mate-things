import NodeDeviceIdSearchCombobox from "@/components/nodes/NodeDeviceIdSearchCombobox";
import TimeRangeFilter from "@/components/records/TimeRangeFilter";
import type { NodeLogLevel } from "@/lib/api/node-logs";

interface NodeLogFiltersProps {
  start?: string;
  end?: string;
  nodeDeviceId?: string;
  nodeName?: string;
  level?: NodeLogLevel;
}

export default function NodeLogFilters(props: NodeLogFiltersProps) {
  return (
    <form className="border-border bg-muted space-y-3 rounded-2xl border p-4">
      <div className="grid gap-3 sm:grid-cols-2 lg:grid-cols-3">
        <TimeRangeFilter start={props.start} end={props.end} />
        <NodeDeviceIdSearchCombobox
          name="node_device_id"
          defaultDeviceId={props.nodeDeviceId}
          defaultNodeName={props.nodeName}
        />
        <div>
          <label
            className="text-foreground/70 mb-1.5 block text-xs font-semibold"
            htmlFor="node-log-level"
          >
            Level
          </label>
          <select
            id="node-log-level"
            name="level"
            defaultValue={props.level ?? ""}
            className="border-control-border bg-background text-foreground focus-visible:border-focus focus-visible:ring-focus min-h-11 w-full rounded-xl border px-3.5 py-2.5 text-sm focus-visible:ring-2 focus-visible:outline-none"
          >
            <option value="">All levels</option>
            <option value="NONE">None</option>
            <option value="ERROR">Error</option>
            <option value="WARN">Warning</option>
            <option value="INFO">Info</option>
            <option value="DEBUG">Debug</option>
          </select>
        </div>
      </div>
      <div className="border-border flex justify-end gap-2 border-t pt-3">
        <a
          href="/node-logs"
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
