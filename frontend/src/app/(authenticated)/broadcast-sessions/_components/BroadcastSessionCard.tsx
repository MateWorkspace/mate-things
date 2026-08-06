import ResourceCard from "@/components/collection/ResourceCard";
import LocalDateTime from "@/components/ui/local-date-time";
import type { BroadcastSessionResponse } from "@/lib/api/telemetry";

export default function BroadcastSessionCard({
  session,
}: {
  session: BroadcastSessionResponse;
}) {
  return (
    <ResourceCard
      title={session.node_device_id ?? "All devices"}
      summary={session.metric_name ?? "All metrics"}
    >
      <dl className="space-y-2.5">
        <div className="flex items-start justify-between gap-4">
          <dt className="text-muted-foreground">Remote address</dt>
          <dd className="min-w-0 text-right font-mono break-words">
            {session.remote_addr}
          </dd>
        </div>
        <div className="flex items-start justify-between gap-4">
          <dt className="text-muted-foreground">Connected</dt>
          <dd>
            <LocalDateTime value={session.connected_at} />
          </dd>
        </div>
      </dl>
    </ResourceCard>
  );
}
