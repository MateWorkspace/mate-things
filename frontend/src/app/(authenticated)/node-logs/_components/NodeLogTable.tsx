"use client";

import { useState } from "react";

import Button from "@/components/ui/button";
import LocalDateTime from "@/components/ui/local-date-time";
import StatusBadge from "@/components/ui/status-badge";
import type { NodeLogResponse } from "@/lib/api/node-logs";

import { NODE_LOG_LEVEL_LABELS } from "../_lib/status";

const BATCH_SIZE = 100;

export default function NodeLogTable({
  records,
}: {
  records: readonly NodeLogResponse[];
}) {
  const [visible, setVisible] = useState(BATCH_SIZE);
  const shown = records.slice(0, visible);

  return (
    <div className="space-y-4">
      <div className="border-border overflow-x-auto rounded-2xl border">
        <table className="w-full text-left text-sm">
          <thead className="bg-muted text-muted-foreground text-xs font-semibold tracking-wider uppercase">
            <tr>
              <th scope="col" className="px-4 py-3 whitespace-nowrap">
                Timestamp
              </th>
              <th scope="col" className="px-4 py-3 whitespace-nowrap">
                Node
              </th>
              <th scope="col" className="px-4 py-3 whitespace-nowrap">
                Level
              </th>
              <th scope="col" className="px-4 py-3 whitespace-nowrap">
                Tag
              </th>
              <th scope="col" className="px-4 py-3 whitespace-nowrap">
                Message
              </th>
            </tr>
          </thead>
          <tbody className="divide-border divide-y">
            {shown.map((record) => {
              const level = NODE_LOG_LEVEL_LABELS[record.level];

              return (
                <tr
                  key={record.id}
                  className="hover:bg-highlight/20 transition-colors"
                >
                  <td className="px-4 py-3 whitespace-nowrap">
                    <LocalDateTime value={record.logged_at} />
                  </td>
                  <td className="px-4 py-3 font-mono text-xs whitespace-nowrap">
                    {record.node_device_id}
                  </td>
                  <td className="px-4 py-3 whitespace-nowrap">
                    <StatusBadge variant={level.variant}>
                      {level.label}
                    </StatusBadge>
                  </td>
                  <td className="px-4 py-3 whitespace-nowrap">
                    {record.tag || "untagged"}
                  </td>
                  <td className="px-4 py-3 whitespace-nowrap">
                    {record.message}
                  </td>
                </tr>
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
