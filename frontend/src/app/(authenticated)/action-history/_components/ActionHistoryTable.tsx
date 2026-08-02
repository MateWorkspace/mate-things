"use client";

import { ChevronRight } from "lucide-react";
import { Fragment, useState } from "react";

import JsonPayload from "@/components/records/JsonPayload";
import StatusBadge from "@/components/ui/status-badge";
import type { ActionLogResponse } from "@/lib/api/action-logs";

import { ACTION_STATUS_LABELS } from "../_lib/status";

const DATE_FORMATTER = new Intl.DateTimeFormat("en", {
  dateStyle: "medium",
  timeStyle: "short",
  timeZone: "UTC",
});

export default function ActionHistoryTable({
  records,
}: {
  records: readonly ActionLogResponse[];
}) {
  const [expandedId, setExpandedId] = useState<number | null>(null);

  return (
    <div className="border-border overflow-x-auto rounded-2xl border">
      <table className="w-full text-left text-sm">
        <thead className="bg-muted text-muted-foreground text-xs font-semibold tracking-wider uppercase">
          <tr>
            <th scope="col" className="px-4 py-3">
              Timestamp
            </th>
            <th scope="col" className="px-4 py-3">
              Action
            </th>
            <th scope="col" className="px-4 py-3">
              Node
            </th>
            <th scope="col" className="px-4 py-3">
              Status
            </th>
          </tr>
        </thead>
        <tbody className="divide-border divide-y">
          {records.map((record) => {
            const expanded = expandedId === record.id;
            const status = ACTION_STATUS_LABELS[record.action_status];
            const nodeLabel =
              record.node_name || record.node_device_id || "Not assigned";

            const detailId = `action-log-detail-${record.id}`;

            return (
              <Fragment key={record.id}>
                <tr className="hover:bg-highlight/20 transition-colors">
                  <td className="px-4 py-3">
                    <button
                      type="button"
                      onClick={() => setExpandedId(expanded ? null : record.id)}
                      aria-expanded={expanded}
                      aria-controls={detailId}
                      className="flex items-center gap-2"
                    >
                      <ChevronRight
                        aria-hidden="true"
                        className={`size-3.5 shrink-0 transition-transform ${expanded ? "rotate-90" : ""}`}
                      />
                      <time dateTime={record.executed_at}>
                        {DATE_FORMATTER.format(new Date(record.executed_at))}
                      </time>
                    </button>
                  </td>
                  <td className="px-4 py-3" title={record.action_id}>
                    {record.action_name ?? record.action_id}
                  </td>
                  <td className="px-4 py-3" title={record.node_id ?? undefined}>
                    {nodeLabel}
                  </td>
                  <td className="px-4 py-3">
                    <StatusBadge variant={status.variant}>
                      {status.label}
                    </StatusBadge>
                  </td>
                </tr>
                {expanded ? (
                  <tr id={detailId}>
                    <td colSpan={4} className="bg-muted/50 px-4 py-4">
                      {record.action_message ? (
                        <p className="mb-3 text-sm whitespace-pre-wrap">
                          {record.action_message}
                        </p>
                      ) : null}
                      <JsonPayload value={record.payload} hideToggle />
                    </td>
                  </tr>
                ) : null}
              </Fragment>
            );
          })}
        </tbody>
      </table>
    </div>
  );
}
