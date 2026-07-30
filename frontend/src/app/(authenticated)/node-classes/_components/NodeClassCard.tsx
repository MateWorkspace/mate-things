import Link from "next/link";

import ResourceCard from "@/components/collection/ResourceCard";
import type { NodeClassResponse } from "@/lib/api/node-classes";

import NodeClassForm from "./NodeClassForm";

interface NodeClassCardProps {
  firmwareCount?: number;
  nodeClass: NodeClassResponse;
  nodeCount?: number;
  permissions: readonly string[];
}

const DATE_FORMATTER = new Intl.DateTimeFormat("en", {
  dateStyle: "medium",
  timeStyle: "short",
  timeZone: "UTC",
});

function CountItem({ label, value }: { label: string; value: number }) {
  return (
    <div className="border-border bg-muted rounded-xl border px-3 py-2">
      <dt className="text-muted-foreground text-xs font-semibold tracking-wider uppercase">
        {label}
      </dt>
      <dd className="mt-1 text-lg font-semibold">{value}</dd>
    </div>
  );
}

export default function NodeClassCard({
  firmwareCount,
  nodeClass,
  nodeCount,
  permissions,
}: NodeClassCardProps) {
  const canEdit = new Set(permissions).has("node_class:set");
  const freshness = nodeClass.updated_at ?? nodeClass.created_at;

  return (
    <ResourceCard
      title={nodeClass.name}
      actions={
        canEdit ? (
          <NodeClassForm nodeClass={nodeClass} canDelete={false} />
        ) : null
      }
    >
      <p className="text-foreground/70 line-clamp-3 min-h-[3.75rem]">
        {nodeClass.description || "No description provided."}
      </p>

      {nodeCount !== undefined || firmwareCount !== undefined ? (
        <dl className="mt-4 grid grid-cols-2 gap-2">
          {nodeCount !== undefined ? (
            <CountItem label="Nodes" value={nodeCount} />
          ) : null}
          {firmwareCount !== undefined ? (
            <CountItem label="Firmware" value={firmwareCount} />
          ) : null}
        </dl>
      ) : null}

      <dl className="mt-4">
        <div className="flex items-start justify-between gap-4">
          <dt className="text-muted-foreground">
            {nodeClass.updated_at ? "Updated" : "Created"}
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
        href={`/node-classes/${nodeClass.id}`}
      >
        View details
      </Link>
    </ResourceCard>
  );
}
