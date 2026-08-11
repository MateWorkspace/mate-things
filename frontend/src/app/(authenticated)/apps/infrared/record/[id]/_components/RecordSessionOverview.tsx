import LocalDateTime from "@/components/ui/local-date-time";
import StatusBadge from "@/components/ui/status-badge";
import type { InfraredRecordSessionResponse } from "@/lib/api/infrared";

import { RECORDING_STATE_LABELS } from "../../_lib/status";
import DeleteRecordSessionDialog from "./DeleteRecordSessionDialog";

export default function RecordSessionOverview({
  session,
  brand,
  model,
  deviceTypeName,
  canDelete,
}: {
  session: InfraredRecordSessionResponse;
  brand: string | null;
  model: string | null;
  deviceTypeName: string | null;
  canDelete: boolean;
}) {
  const status = RECORDING_STATE_LABELS[session.recording_state];
  const deviceLabel = [brand, model].filter(Boolean).join(" ") || "—";
  return (
    <div className="flex flex-wrap items-start justify-between gap-4">
      <dl className="grid grid-cols-2 gap-4 text-sm sm:grid-cols-4">
        <div>
          <dt className="text-muted-foreground">Device</dt>
          <dd className="mt-1 font-medium">{deviceLabel}</dd>
        </div>
        <div>
          <dt className="text-muted-foreground">Device type</dt>
          <dd className="mt-1 font-medium">{deviceTypeName ?? "—"}</dd>
        </div>
        <div>
          <dt className="text-muted-foreground">Status</dt>
          <dd className="mt-1">
            <StatusBadge variant={status.variant}>{status.label}</StatusBadge>
          </dd>
        </div>
        <div>
          <dt className="text-muted-foreground">Created</dt>
          <dd className="mt-1 font-medium">
            <LocalDateTime value={session.created_at} />
          </dd>
        </div>
      </dl>
      {canDelete ? <DeleteRecordSessionDialog sessionId={session.id} /> : null}
    </div>
  );
}
