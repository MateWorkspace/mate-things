import JsonPayload from "@/components/records/JsonPayload";
import Card from "@/components/ui/card";
import type { TelemetryRecordResponse } from "@/lib/api/telemetry";

export default function TelemetryCard({
  record,
}: {
  record: TelemetryRecordResponse;
}) {
  return (
    <Card className="h-full">
      <div className="flex items-start justify-between gap-3">
        <div>
          <h2 className="font-display text-primary text-lg">
            {record.metric_name}
          </h2>
          <p className="text-muted-foreground mt-1 font-mono text-xs">
            {record.node_device_id}
          </p>
        </div>
        <time
          className="text-muted-foreground text-right text-xs"
          dateTime={record.recorded_at}
        >
          {new Date(record.recorded_at).toLocaleString()}
        </time>
      </div>
      <p className="mt-4 text-sm">
        <span className="text-muted-foreground">Schema:</span>{" "}
        {record.payload_schema_name} · v{record.payload_schema_version}
      </p>
      <div className="mt-4">
        <JsonPayload value={record.payload} />
      </div>
    </Card>
  );
}
