import { CloudUpload, Settings } from "lucide-react";
import Link from "next/link";

import ResourceCard from "@/components/collection/ResourceCard";
import LocalDateTime from "@/components/ui/local-date-time";
import StatusBadge from "@/components/ui/status-badge";
import type { NodeResponse } from "@/lib/api";

interface NodeCardProps {
  node: NodeResponse;
  permissions: readonly string[];
  nodeClassName?: string;
  firmwareName?: string;
}

function DetailRow({
  label,
  children,
}: {
  label: string;
  children: React.ReactNode;
}) {
  return (
    <div className="flex items-start justify-between gap-4">
      <dt className="text-muted-foreground">{label}</dt>
      <dd className="min-w-0 text-right font-medium break-words">{children}</dd>
    </div>
  );
}

export default function NodeCard({
  firmwareName,
  node,
  nodeClassName,
  permissions,
}: NodeCardProps) {
  const permissionSet = new Set(permissions);
  const freshness = node.updated_at ?? node.created_at;

  return (
    <ResourceCard
      title={node.name}
      summary={
        <StatusBadge variant={node.is_connected ? "success" : "critical"}>
          {node.is_connected ? "Connected" : "Disconnected"}
        </StatusBadge>
      }
    >
      <dl className="space-y-2.5">
        <DetailRow label="Device ID">
          <span className="font-mono">{node.device_id}</span>
        </DetailRow>
        <DetailRow label="Class">{nodeClassName ?? "Unresolved"}</DetailRow>
        <DetailRow label="Firmware">{firmwareName ?? "Unresolved"}</DetailRow>
        <DetailRow label={node.updated_at ? "Updated" : "Registered"}>
          <LocalDateTime value={freshness} />
        </DetailRow>
      </dl>

      <p className="text-foreground/70 mt-4 line-clamp-3 min-h-[3.75rem]">
        {node.description || "No description provided."}
      </p>

      <div className="mt-5 flex flex-wrap items-center gap-2">
        <Link
          className="bg-primary text-surface focus-visible:ring-focus focus-visible:ring-offset-background inline-flex min-h-11 flex-1 items-center justify-center rounded-xl px-4 py-2.5 text-sm font-semibold transition-opacity hover:opacity-90 focus-visible:ring-2 focus-visible:ring-offset-2 focus-visible:outline-none"
          href={`/nodes/${node.id}`}
        >
          View details
        </Link>
        {permissionSet.has("node:set") ? (
          <Link
            aria-label={`Edit ${node.name}`}
            className="border-border text-primary hover:bg-highlight/40 focus-visible:ring-focus focus-visible:ring-offset-background inline-flex size-11 items-center justify-center rounded-xl border transition-colors focus-visible:ring-2 focus-visible:ring-offset-2 focus-visible:outline-none"
            href={`/nodes/${node.id}?tab=overview`}
          >
            <Settings aria-hidden="true" className="size-4" />
          </Link>
        ) : null}
        {permissionSet.has("ota:dispatch") ? (
          <Link
            aria-label={`Dispatch OTA to ${node.name}`}
            className="border-border text-primary hover:bg-highlight/40 focus-visible:ring-focus focus-visible:ring-offset-background inline-flex size-11 items-center justify-center rounded-xl border transition-colors focus-visible:ring-2 focus-visible:ring-offset-2 focus-visible:outline-none"
            href={`/nodes/${node.id}?tab=firmware`}
          >
            <CloudUpload aria-hidden="true" className="size-4" />
          </Link>
        ) : null}
      </div>
    </ResourceCard>
  );
}
