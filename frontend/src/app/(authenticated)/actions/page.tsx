import type { Metadata } from "next";
import { redirect } from "next/navigation";

import ActionCard from "@/app/(authenticated)/actions/_components/ActionCard";
import ActionFilters from "@/app/(authenticated)/actions/_components/ActionFilters";
import ActionForm from "@/app/(authenticated)/actions/_components/ActionForm";
import Pagination from "@/components/collection/Pagination";
import PageHeader from "@/components/ui/page-header";
import { EmptyState } from "@/components/ui/states";
import { listActions } from "@/lib/api/actions";
import { getNodeClassById } from "@/lib/api/node-classes";
import { listAllPayloadSchemas } from "@/lib/api/payload-schemas";
import { getOptionalById } from "@/lib/api/optional";
import {
  getOutOfRangePageRedirect,
  parsePageQuery,
} from "@/lib/collection-query";
import { requirePermission } from "@/lib/session";

export const metadata: Metadata = { title: "Actions — Mate Things" };
type RawSearchParams = Record<string, string | string[] | undefined>;

export default async function ActionsPage({
  searchParams,
}: {
  searchParams: Promise<RawSearchParams>;
}) {
  const [raw, { permissions }] = await Promise.all([
    searchParams,
    requirePermission("action:get"),
  ]);
  const pageQuery = parsePageQuery(raw);
  const nodeClassId =
    typeof raw.node_class_id === "string" ? raw.node_class_id.trim() : "";
  const canReadNodeClasses = permissions.has("node_class:get");
  const [result, schemas, nodeClassDefault] = await Promise.all([
    listActions({
      ...pageQuery,
      node_class_id: nodeClassId || undefined,
    }),
    permissions.has("payload_schema:get")
      ? listAllPayloadSchemas()
      : Promise.resolve([]),
    canReadNodeClasses && nodeClassId
      ? getOptionalById(() => getNodeClassById(nodeClassId))
      : null,
  ]);
  const redirectTarget = getOutOfRangePageRedirect(
    "/actions",
    raw,
    result.page,
  );
  if (redirectTarget) {
    redirect(redirectTarget);
  }
  const canCreate =
    permissions.has("action:add") && permissions.has("payload_schema:get");

  return (
    <main className="mx-auto w-full max-w-7xl space-y-6 px-4 py-6 sm:px-6 lg:px-8">
      <PageHeader
        title="Actions"
        description="Define fleet commands, match them to compatible node classes, and dispatch them safely."
        actions={canCreate ? <ActionForm schemas={schemas} /> : undefined}
      />
      <ActionFilters
        canReadNodeClasses={canReadNodeClasses}
        search={pageQuery.search}
        nodeClassId={nodeClassId}
        nodeClassName={nodeClassDefault?.name}
      />
      <p className="text-muted-foreground text-sm">
        {result.page.total_items} action definitions
      </p>
      {result.data.length ? (
        <div className="grid gap-4 md:grid-cols-2 xl:grid-cols-3">
          {result.data.map((action) => (
            <ActionCard key={action.id} action={action} />
          ))}
        </div>
      ) : (
        <EmptyState
          title={
            pageQuery.search || nodeClassId
              ? "No matching actions"
              : "No actions yet"
          }
          description={
            pageQuery.search || nodeClassId
              ? "Clear or change the filters to broaden the results."
              : "Create an action definition when your role permits it."
          }
        />
      )}
      <Pagination page={result.page} pathname="/actions" searchParams={raw} />
    </main>
  );
}
