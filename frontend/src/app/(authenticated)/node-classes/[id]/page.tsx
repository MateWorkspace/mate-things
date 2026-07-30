import type { ReactNode } from "react";
import type { Metadata } from "next";
import Link from "next/link";
import { notFound } from "next/navigation";

import Card from "@/components/ui/card";
import PageHeader from "@/components/ui/page-header";
import StatusBadge from "@/components/ui/status-badge";
import { listActions, type ActionResponse } from "@/lib/api/actions";
import { ApiError } from "@/lib/api/client";
import {
  listFirmwaresByNodeClassId,
  type FirmwareResponse,
} from "@/lib/api/firmwares";
import { getNodeClassById } from "@/lib/api/node-classes";
import { listNodes, type NodeResponse } from "@/lib/api/nodes";
import { requirePermission } from "@/lib/session";

import NodeClassForm from "../_components/NodeClassForm";

export const metadata: Metadata = {
  title: "Node Class Details — Mate Things",
};

interface NodeClassDetailPageProps {
  params: Promise<{ id: string }>;
}

const DATE_FORMATTER = new Intl.DateTimeFormat("en", {
  dateStyle: "medium",
  timeStyle: "short",
  timeZone: "UTC",
});

function formatBytes(size: number): string {
  if (size < 1024) {
    return `${size} B`;
  }

  return `${(size / 1024).toFixed(size % 1024 === 0 ? 0 : 1)} KB`;
}

function RelationshipSection({
  children,
  count,
  emptyCopy,
  href,
  linkLabel,
  title,
}: {
  children: ReactNode;
  count: number;
  emptyCopy: string;
  href: string;
  linkLabel: string;
  title: string;
}) {
  return (
    <section aria-labelledby={`${title.toLowerCase()}-heading`}>
      <div className="mb-3 flex flex-wrap items-end justify-between gap-3">
        <div>
          <h2
            id={`${title.toLowerCase()}-heading`}
            className="font-display text-primary text-2xl tracking-wide"
          >
            {title}
          </h2>
          <p className="text-muted-foreground mt-1 text-sm">
            {count} associated {count === 1 ? "resource" : "resources"}
          </p>
        </div>
        <Link
          className="text-primary focus-visible:ring-focus rounded-lg text-sm font-semibold underline-offset-4 hover:underline focus-visible:ring-2 focus-visible:outline-none"
          href={href}
        >
          {linkLabel}
        </Link>
      </div>
      {count > 0 ? (
        <div className="grid grid-cols-1 gap-3 md:grid-cols-2 xl:grid-cols-3">
          {children}
        </div>
      ) : (
        <p className="border-border bg-muted rounded-xl border border-dashed p-4 text-sm">
          {emptyCopy}
        </p>
      )}
    </section>
  );
}

function RelationshipLink({
  children,
  href,
  label,
}: {
  children: ReactNode;
  href: string;
  label: string;
}) {
  return (
    <Card className="flex min-h-32 flex-col justify-between gap-3 p-4">
      {children}
      <Link
        aria-label={label}
        className="text-primary focus-visible:ring-focus w-fit rounded-lg text-sm font-semibold underline-offset-4 hover:underline focus-visible:ring-2 focus-visible:outline-none"
        href={href}
      >
        View details
      </Link>
    </Card>
  );
}

function FirmwareRelationship({ firmware }: { firmware: FirmwareResponse }) {
  return (
    <RelationshipLink
      href={`/firmware/${firmware.id}`}
      label={`View ${firmware.name}`}
    >
      <div>
        <h3 className="font-semibold">{firmware.name}</h3>
        <p className="text-muted-foreground mt-1 text-sm">
          {formatBytes(firmware.size)} ·{" "}
          <span className="font-mono">{firmware.checksum.slice(0, 12)}</span>
        </p>
      </div>
    </RelationshipLink>
  );
}

function NodeRelationship({ node }: { node: NodeResponse }) {
  return (
    <RelationshipLink href={`/nodes/${node.id}`} label={`View ${node.name}`}>
      <div>
        <div className="flex flex-wrap items-start justify-between gap-2">
          <h3 className="font-semibold">{node.name}</h3>
          <StatusBadge variant={node.is_connected ? "success" : "critical"}>
            {node.is_connected ? "Connected" : "Disconnected"}
          </StatusBadge>
        </div>
        <p className="text-muted-foreground mt-2 font-mono text-sm">
          {node.device_id}
        </p>
      </div>
    </RelationshipLink>
  );
}

function ActionRelationship({ action }: { action: ActionResponse }) {
  return (
    <RelationshipLink
      href={`/actions/${action.id}`}
      label={`View ${action.name}`}
    >
      <div>
        <h3 className="font-semibold">{action.name}</h3>
        <p className="text-muted-foreground mt-1 line-clamp-2 text-sm">
          {action.description || "No description provided."}
        </p>
        <p className="text-muted-foreground mt-2 text-xs">
          Schema: {action.payload_schema_name} v{action.payload_schema_version}
        </p>
      </div>
    </RelationshipLink>
  );
}

export default async function NodeClassDetailPage({
  params,
}: NodeClassDetailPageProps) {
  const [{ id }, { permissions }] = await Promise.all([
    params,
    requirePermission("node_class:get"),
  ]);
  const canReadFirmware = permissions.has("firmware:get");
  const canReadNodes = permissions.has("node:get");
  const canReadActions = permissions.has("action:get");
  const nodeClassRequest = getNodeClassById(id).catch((error: unknown) => {
    if (error instanceof ApiError && error.status === 404) {
      return null;
    }

    throw error;
  });

  const [nodeClass, firmwares, nodes, actions] = await Promise.all([
    nodeClassRequest,
    canReadFirmware
      ? listFirmwaresByNodeClassId(id, { limit: 12 })
      : Promise.resolve(null),
    canReadNodes
      ? listNodes({ node_class_id: id, limit: 12 })
      : Promise.resolve(null),
    canReadActions
      ? listActions({ node_class_id: id, limit: 12 })
      : Promise.resolve(null),
  ]);

  if (!nodeClass) {
    notFound();
  }

  const freshness = nodeClass.updated_at ?? nodeClass.created_at;
  const canEdit = permissions.has("node_class:set");
  const canDelete = permissions.has("node_class:remove");

  return (
    <main className="mx-auto w-full max-w-7xl space-y-8 px-4 py-6 sm:px-6 lg:px-8">
      <PageHeader
        title={nodeClass.name}
        description={
          nodeClass.description || "No description has been provided."
        }
        actions={
          canEdit || canDelete ? (
            <NodeClassForm
              canDelete={canDelete}
              canEdit={canEdit}
              nodeClass={nodeClass}
            />
          ) : undefined
        }
      />

      <Card className="p-5">
        <h2 className="font-display text-primary text-xl tracking-wide">
          Class overview
        </h2>
        <dl className="mt-4 grid grid-cols-1 gap-4 text-sm sm:grid-cols-2 lg:grid-cols-4">
          <div>
            <dt className="text-muted-foreground">Class ID</dt>
            <dd className="mt-1 font-mono break-all">{nodeClass.id}</dd>
          </div>
          <div>
            <dt className="text-muted-foreground">
              {nodeClass.updated_at ? "Updated" : "Created"}
            </dt>
            <dd className="mt-1 font-medium">
              <time dateTime={freshness}>
                {DATE_FORMATTER.format(new Date(freshness))}
              </time>
            </dd>
          </div>
          {nodes ? (
            <div>
              <dt className="text-muted-foreground">Nodes</dt>
              <dd className="mt-1 font-medium">{nodes.page.total_items}</dd>
            </div>
          ) : null}
          {firmwares ? (
            <div>
              <dt className="text-muted-foreground">Firmware</dt>
              <dd className="mt-1 font-medium">{firmwares.page.total_items}</dd>
            </div>
          ) : null}
        </dl>
      </Card>

      {nodes ? (
        <RelationshipSection
          count={nodes.page.total_items}
          emptyCopy="No nodes are assigned to this class."
          href={`/nodes?node_class_id=${encodeURIComponent(id)}`}
          linkLabel="View all nodes"
          title="Nodes"
        >
          {nodes.data.map((node) => (
            <NodeRelationship key={node.id} node={node} />
          ))}
        </RelationshipSection>
      ) : null}

      {firmwares ? (
        <RelationshipSection
          count={firmwares.page.total_items}
          emptyCopy="No firmware belongs to this class."
          href={`/firmware?node_class_id=${encodeURIComponent(id)}`}
          linkLabel="View all firmware"
          title="Firmware"
        >
          {firmwares.data.map((firmware) => (
            <FirmwareRelationship key={firmware.id} firmware={firmware} />
          ))}
        </RelationshipSection>
      ) : null}

      {actions ? (
        <RelationshipSection
          count={actions.page.total_items}
          emptyCopy="No actions belong to this class."
          href={`/actions?node_class_id=${encodeURIComponent(id)}`}
          linkLabel="View all actions"
          title="Actions"
        >
          {actions.data.map((action) => (
            <ActionRelationship key={action.id} action={action} />
          ))}
        </RelationshipSection>
      ) : null}
    </main>
  );
}
