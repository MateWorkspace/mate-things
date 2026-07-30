import Link from "next/link";

import ResourceCard from "@/components/collection/ResourceCard";
import StatusBadge from "@/components/ui/status-badge";
import type { FirmwareResponse } from "@/lib/api/firmwares";

import { formatBytes } from "../_lib/format";
import FirmwareForm, { type FirmwareNodeClassOption } from "./FirmwareForm";

interface FirmwareCardProps {
  firmware: FirmwareResponse;
  nodeClasses?: readonly FirmwareNodeClassOption[];
  nodeClassName?: string;
  permissions: readonly string[];
}

const DATE_FORMATTER = new Intl.DateTimeFormat("en", {
  dateStyle: "medium",
  timeStyle: "short",
  timeZone: "UTC",
});

export default function FirmwareCard({
  firmware,
  nodeClasses = [],
  nodeClassName,
  permissions,
}: FirmwareCardProps) {
  const freshness = firmware.updated_at ?? firmware.created_at;
  const hasBinary = Boolean(firmware.binary_path) && firmware.size > 0;
  const canEdit = new Set(permissions).has("firmware:set");

  return (
    <ResourceCard
      title={firmware.name}
      summary={nodeClassName ?? `Class ${firmware.node_class_id}`}
      actions={
        <div className="flex flex-col items-end gap-2">
          <StatusBadge variant={hasBinary ? "success" : "warning"}>
            {hasBinary ? "Binary available" : "Binary unavailable"}
          </StatusBadge>
          {canEdit ? (
            <FirmwareForm
              firmware={firmware}
              nodeClasses={nodeClasses}
              canEdit
            />
          ) : null}
        </div>
      }
    >
      <dl className="space-y-3">
        <div className="flex items-start justify-between gap-4">
          <dt className="text-muted-foreground">Size</dt>
          <dd className="font-medium">{formatBytes(firmware.size)}</dd>
        </div>
        <div className="flex items-start justify-between gap-4">
          <dt className="text-muted-foreground">Checksum</dt>
          <dd
            className="max-w-40 truncate font-mono text-xs"
            title={firmware.checksum}
          >
            {firmware.checksum || "Not reported"}
          </dd>
        </div>
        <div className="flex items-start justify-between gap-4">
          <dt className="text-muted-foreground">
            {firmware.updated_at ? "Updated" : "Created"}
          </dt>
          <dd className="text-right font-medium">
            <time dateTime={freshness}>
              {DATE_FORMATTER.format(new Date(freshness))}
            </time>
          </dd>
        </div>
      </dl>

      <Link
        className="bg-primary text-surface focus-visible:ring-focus focus-visible:ring-offset-background mt-5 inline-flex min-h-11 w-full items-center justify-center rounded-xl px-4 py-2.5 text-sm font-semibold transition-opacity hover:opacity-90 focus-visible:ring-2 focus-visible:ring-offset-2 focus-visible:outline-none"
        href={`/firmware/${firmware.id}`}
      >
        View details
      </Link>
    </ResourceCard>
  );
}
