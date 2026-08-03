"use client";

import { ChevronRight } from "lucide-react";
import { Fragment, useState } from "react";

import JsonPayload from "@/components/records/JsonPayload";
import Button from "@/components/ui/button";
import LocalDateTime from "@/components/ui/local-date-time";
import type { TelemetryRecordResponse } from "@/lib/api/telemetry";

const BATCH_SIZE = 100;

export default function TelemetryTable({
  records,
}: {
  records: readonly TelemetryRecordResponse[];
}) {
  const [visible, setVisible] = useState(BATCH_SIZE);
  const [expandedId, setExpandedId] = useState<number | null>(null);
  const shown = records.slice(0, visible);

  return (
    <div className="space-y-4">
      <div className="border-border overflow-x-auto rounded-2xl border">
        <table className="w-full text-left text-sm">
          <thead className="bg-muted text-muted-foreground text-xs font-semibold tracking-wider uppercase">
            <tr>
              <th scope="col" className="px-4 py-3 whitespace-nowrap">
                Time
              </th>
              <th scope="col" className="px-4 py-3 whitespace-nowrap">
                Node
              </th>
              <th scope="col" className="px-4 py-3 whitespace-nowrap">
                Metric
              </th>
              <th scope="col" className="px-4 py-3 whitespace-nowrap">
                Schema
              </th>
              <th scope="col" className="px-4 py-3 whitespace-nowrap">
                Payload
              </th>
            </tr>
          </thead>
          <tbody className="divide-border divide-y">
            {shown.map((record) => {
              const expanded = expandedId === record.id;
              const detailId = `telemetry-detail-${record.id}`;
              const rawJson = JSON.stringify(record.payload);

              return (
                <Fragment key={record.id}>
                  <tr className="hover:bg-highlight/20 transition-colors">
                    <td className="px-4 py-3 whitespace-nowrap">
                      <button
                        type="button"
                        onClick={() =>
                          setExpandedId(expanded ? null : record.id)
                        }
                        aria-expanded={expanded}
                        aria-controls={detailId}
                        className="flex items-center gap-2"
                      >
                        <ChevronRight
                          aria-hidden="true"
                          className={`size-3.5 shrink-0 transition-transform ${expanded ? "rotate-90" : ""}`}
                        />
                        <LocalDateTime value={record.recorded_at} />
                      </button>
                    </td>
                    <td className="px-4 py-3 font-mono text-xs whitespace-nowrap">
                      {record.node_device_id}
                    </td>
                    <td className="px-4 py-3 whitespace-nowrap">
                      {record.metric_name}
                    </td>
                    <td className="px-4 py-3 whitespace-nowrap">
                      {record.payload_schema_name} · v
                      {record.payload_schema_version}
                    </td>
                    <td className="px-4 py-3 font-mono text-xs whitespace-nowrap">
                      {rawJson}
                    </td>
                  </tr>
                  {expanded ? (
                    <tr id={detailId}>
                      <td colSpan={5} className="bg-muted/50 px-4 py-4">
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
      {visible < records.length ? (
        <div className="flex justify-center">
          <Button
            type="button"
            variant="secondary"
            onClick={() =>
              setVisible((current) =>
                Math.min(records.length, current + BATCH_SIZE),
              )
            }
          >
            Show {Math.min(BATCH_SIZE, records.length - visible)} more
          </Button>
        </div>
      ) : null}
    </div>
  );
}
