"use client";

import { useState, type ReactNode } from "react";

import Button from "@/components/ui/button";

interface RecordWindowProps {
  children: readonly ReactNode[];
  batchSize?: number;
}

export default function RecordWindow({
  children,
  batchSize = 100,
}: RecordWindowProps) {
  const [visible, setVisible] = useState(batchSize);
  const shown = children.slice(0, visible);

  return (
    <div className="space-y-5">
      <div className="grid gap-4 md:grid-cols-2 xl:grid-cols-3">{shown}</div>
      {visible < children.length ? (
        <div className="flex justify-center">
          <Button
            type="button"
            variant="secondary"
            onClick={() =>
              setVisible((current) =>
                Math.min(children.length, current + batchSize),
              )
            }
          >
            Show {Math.min(batchSize, children.length - visible)} more
          </Button>
        </div>
      ) : null}
    </div>
  );
}
