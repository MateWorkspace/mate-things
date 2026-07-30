"use client";

import { useState, type ReactNode } from "react";

import Button from "@/components/ui/button";

interface RecordWindowProps<T> {
  records: readonly T[];
  renderRecord: (record: T, index: number) => ReactNode;
  getKey: (record: T) => string | number;
  batchSize?: number;
}

export default function RecordWindow<T>({
  records,
  renderRecord,
  getKey,
  batchSize = 100,
}: RecordWindowProps<T>) {
  const [visible, setVisible] = useState(batchSize);
  const shown = records.slice(0, visible);

  return (
    <div className="space-y-5">
      <div className="grid gap-4 md:grid-cols-2 xl:grid-cols-3">
        {shown.map((record, index) => (
          <div key={getKey(record)}>{renderRecord(record, index)}</div>
        ))}
      </div>
      {visible < records.length ? (
        <div className="flex justify-center">
          <Button
            type="button"
            variant="secondary"
            onClick={() =>
              setVisible((current) =>
                Math.min(records.length, current + batchSize),
              )
            }
          >
            Show {Math.min(batchSize, records.length - visible)} more
          </Button>
        </div>
      ) : null}
    </div>
  );
}
