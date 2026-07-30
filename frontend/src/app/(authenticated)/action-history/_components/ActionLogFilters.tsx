import TimeRangeFilter from "@/components/records/TimeRangeFilter";
import Input from "@/components/ui/input";

interface ActionLogFiltersProps {
  start?: string;
  end?: string;
  actionId?: string;
  nodeId?: string;
  executionId?: string;
}

export default function ActionLogFilters(props: ActionLogFiltersProps) {
  return (
    <form className="border-border bg-muted grid gap-3 rounded-2xl border p-4 sm:grid-cols-2 lg:grid-cols-3">
      <TimeRangeFilter start={props.start} end={props.end} />
      <Input
        name="action_id"
        defaultValue={props.actionId}
        placeholder="Action ID"
        aria-label="Action ID"
      />
      <Input
        name="node_id"
        defaultValue={props.nodeId}
        placeholder="Node ID"
        aria-label="Node ID"
      />
      <Input
        name="execution_id"
        defaultValue={props.executionId}
        placeholder="Execution ID"
        aria-label="Execution ID"
      />
      <div className="flex gap-2">
        <button className="bg-primary text-surface min-h-11 flex-1 rounded-xl px-4 text-sm font-semibold">
          Apply filters
        </button>
        <a
          href="/action-history"
          className="border-border text-primary inline-flex min-h-11 items-center rounded-xl border px-4 text-sm font-semibold"
        >
          Clear
        </a>
      </div>
    </form>
  );
}
