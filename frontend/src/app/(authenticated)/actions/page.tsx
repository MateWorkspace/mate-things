import type { Metadata } from "next";

import ActionCard from "@/app/(authenticated)/actions/_components/ActionCard";
import ActionForm from "@/app/(authenticated)/actions/_components/ActionForm";
import Pagination from "@/components/collection/Pagination";
import Input from "@/components/ui/input";
import PageHeader from "@/components/ui/page-header";
import { EmptyState } from "@/components/ui/states";
import { listActions } from "@/lib/api/actions";
import { listAllPayloadSchemas } from "@/lib/api/payload-schemas";
import { parsePageQuery } from "@/lib/collection-query";
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
  const schemaName =
    typeof raw.payload_schema_name === "string"
      ? raw.payload_schema_name.trim()
      : "";
  const [result, schemas] = await Promise.all([
    listActions({
      ...pageQuery,
      node_class_id: nodeClassId || undefined,
      payload_schema_name: schemaName || undefined,
    }),
    permissions.has("payload_schema:get")
      ? listAllPayloadSchemas()
      : Promise.resolve([]),
  ]);
  const canCreate =
    permissions.has("action:add") && permissions.has("payload_schema:get");

  return (
    <main className="mx-auto w-full max-w-7xl space-y-6 px-4 py-6 sm:px-6 lg:px-8">
      <PageHeader
        title="Actions"
        description="Define fleet commands, match them to compatible node classes, and dispatch them safely."
        actions={canCreate ? <ActionForm schemas={schemas} /> : undefined}
      />
      <form className="border-border bg-muted grid gap-3 rounded-2xl border p-4 sm:grid-cols-[1fr_1fr_auto]">
        <Input
          name="node_class_id"
          defaultValue={nodeClassId}
          placeholder="Node class ID"
          aria-label="Filter by node class ID"
        />
        <Input
          name="payload_schema_name"
          defaultValue={schemaName}
          placeholder="Payload schema"
          aria-label="Filter by payload schema"
        />
        <button className="bg-primary text-surface rounded-xl px-4 py-2.5 text-sm font-semibold">
          Apply
        </button>
      </form>
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
            nodeClassId || schemaName ? "No matching actions" : "No actions yet"
          }
          description={
            nodeClassId || schemaName
              ? "Clear or change the filters to broaden the results."
              : "Create an action definition when your role permits it."
          }
        />
      )}
      <Pagination page={result.page} pathname="/actions" searchParams={raw} />
    </main>
  );
}
