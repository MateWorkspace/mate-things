import Link from "next/link";

import LocalDateTime from "@/components/ui/local-date-time";
import StatusBadge from "@/components/ui/status-badge";
import type { InfraredRecordSessionListItemResponse } from "@/lib/api/infrared";

import { RECORDING_STATE_LABELS } from "../_lib/status";

export default function RecordSessionsTable({
  records,
}: {
  records: readonly InfraredRecordSessionListItemResponse[];
}) {
  return (
    <div className="border-border overflow-x-auto rounded-2xl border">
      <table className="w-full text-left text-sm">
        <thead className="bg-muted text-muted-foreground text-xs font-semibold tracking-wider uppercase">
          <tr>
            <th scope="col" className="px-4 py-3">
              Device
            </th>
            <th scope="col" className="px-4 py-3">
              Device type
            </th>
            <th scope="col" className="px-4 py-3">
              Status
            </th>
            <th scope="col" className="px-4 py-3">
              Created
            </th>
          </tr>
        </thead>
        <tbody className="divide-border divide-y">
          {records.map((record) => {
            const status = RECORDING_STATE_LABELS[record.recording_state];
            return (
              <tr
                key={record.id}
                className="hover:bg-highlight/20 transition-colors"
              >
                <td className="px-4 py-3">
                  <Link
                    href={`/apps/infrared/record/${record.id}`}
                    className="text-primary font-semibold"
                  >
                    {record.brand} {record.model}
                  </Link>
                </td>
                <td className="px-4 py-3">{record.device_type_name}</td>
                <td className="px-4 py-3">
                  <StatusBadge variant={status.variant}>
                    {status.label}
                  </StatusBadge>
                </td>
                <td className="px-4 py-3">
                  <LocalDateTime value={record.created_at} />
                </td>
              </tr>
            );
          })}
        </tbody>
      </table>
    </div>
  );
}
