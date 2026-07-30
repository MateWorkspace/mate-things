import type { Metadata } from "next";
import { notFound } from "next/navigation";

import ActionForm from "@/app/(authenticated)/actions/_components/ActionForm";
import DispatchActionDialog from "@/app/(authenticated)/actions/_components/DispatchActionDialog";
import Card from "@/components/ui/card";
import PageHeader from "@/components/ui/page-header";
import { EmptyState } from "@/components/ui/states";
import { getActionById } from "@/lib/api/actions";
import { ApiError } from "@/lib/api/client";
import { listAllNodeClasses } from "@/lib/api/node-classes";
import { listNodes } from "@/lib/api/nodes";
import { listAllPayloadSchemas } from "@/lib/api/payload-schemas";
import { requirePermission } from "@/lib/session";

export const metadata: Metadata = { title: "Action details — Mate Things" };

export default async function ActionDetailPage({
  params,
}: {
  params: Promise<{ id: string }>;
}) {
  const [{ id }, { permissions }] = await Promise.all([
    params,
    requirePermission("action:get"),
  ]);
  let action;
  try {
    action = await getActionById(id);
  } catch (error) {
    if (error instanceof ApiError && error.status === 404) notFound();
    throw error;
  }
  const [nodeClasses, schemas, nodesPage] = await Promise.all([
    permissions.has("node_class:get")
      ? listAllNodeClasses()
      : Promise.resolve([]),
    permissions.has("payload_schema:get")
      ? listAllPayloadSchemas()
      : Promise.resolve([]),
    permissions.has("node:get") && permissions.has("action:dispatch")
      ? listNodes({ page: 1, limit: 100, node_class_id: action.node_class_id })
      : Promise.resolve({
          data: [],
          page: { page: 1, limit: 100, total_items: 0 },
        }),
  ]);
  const canEdit =
    permissions.has("action:set") &&
    permissions.has("node_class:get") &&
    permissions.has("payload_schema:get");

  return (
    <main className="mx-auto w-full max-w-5xl space-y-6 px-4 py-6 sm:px-6 lg:px-8">
      <PageHeader
        title={action.name}
        description={action.description || "Fleet action definition"}
        actions={
          <div className="flex flex-wrap gap-2">
            {permissions.has("action:dispatch") &&
            permissions.has("node:get") ? (
              <DispatchActionDialog action={action} nodes={nodesPage.data} />
            ) : null}
            {canEdit || permissions.has("action:remove") ? (
              <ActionForm
                action={action}
                nodeClasses={nodeClasses}
                schemas={schemas}
                canEdit={canEdit}
                canDelete={permissions.has("action:remove")}
              />
            ) : null}
          </div>
        }
      />
      <Card>
        <dl className="grid gap-4 sm:grid-cols-2">
          <div>
            <dt className="text-muted-foreground text-sm">Node class</dt>
            <dd className="mt-1 font-medium">
              {nodeClasses.find((item) => item.id === action.node_class_id)
                ?.name ?? action.node_class_id}
            </dd>
          </div>
          <div>
            <dt className="text-muted-foreground text-sm">Payload schema</dt>
            <dd className="mt-1 font-medium">
              {action.payload_schema_name} · version{" "}
              {action.payload_schema_version}
            </dd>
          </div>
          <div>
            <dt className="text-muted-foreground text-sm">Created</dt>
            <dd className="mt-1">
              <time dateTime={action.created_at}>
                {new Date(action.created_at).toLocaleString()}
              </time>
            </dd>
          </div>
          <div>
            <dt className="text-muted-foreground text-sm">Definition ID</dt>
            <dd className="mt-1 font-mono text-xs break-all">{action.id}</dd>
          </div>
        </dl>
      </Card>
      {permissions.has("action:dispatch") && !permissions.has("node:get") ? (
        <EmptyState
          title="Node access required"
          description="node:get permission is required to choose a compatible dispatch target."
        />
      ) : null}
    </main>
  );
}
