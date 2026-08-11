import type { Metadata } from "next";
import { redirect } from "next/navigation";

import Pagination from "@/components/collection/Pagination";
import PageHeader from "@/components/ui/page-header";
import { EmptyState } from "@/components/ui/states";
import { listPayloadSchemas } from "@/lib/api/payload-schemas";
import {
  getOutOfRangePageRedirect,
  parsePageQuery,
} from "@/lib/collection-query";
import { requirePermission } from "@/lib/session";

import PayloadSchemaCard from "./_components/PayloadSchemaCard";
import PayloadSchemaFilters from "./_components/PayloadSchemaFilters";
import PayloadSchemaForm from "./_components/PayloadSchemaForm";

export const metadata: Metadata = { title: "Payload Schemas — Mate Things" };
type Raw = Record<string, string | string[] | undefined>;

export default async function PayloadSchemasPage({
  searchParams,
}: {
  searchParams: Promise<Raw>;
}) {
  const [session, raw] = await Promise.all([
    requirePermission("payload_schema:get"),
    searchParams,
  ]);
  const base = parsePageQuery(raw);
  const stillValid = String(raw.still_valid ?? "") === "1";
  const validAt = stillValid ? new Date().toISOString() : undefined;
  const result = await listPayloadSchemas({ ...base, valid_at: validAt });
  const target = getOutOfRangePageRedirect(
    "/admin/payload-schemas",
    raw,
    result.page,
  );
  if (target) redirect(target);
  return (
    <main className="mx-auto w-full max-w-7xl space-y-6 px-4 py-6 sm:px-6 lg:px-8">
      <PageHeader
        title="Payload Schemas"
        description="Manage versioned JSON contracts and their validity windows."
        actions={
          session.permissions.has("payload_schema:add") ? (
            <PayloadSchemaForm />
          ) : undefined
        }
      />
      <PayloadSchemaFilters search={base.search} stillValid={stillValid} />
      <p className="text-muted-foreground text-xs font-semibold tracking-wider uppercase">
        Showing {result.data.length} of {result.page.total_items} versions
      </p>
      {result.data.length ? (
        <section className="grid gap-5 sm:grid-cols-2 xl:grid-cols-3">
          {result.data.map((schema) => (
            <PayloadSchemaCard key={schema.id} schema={schema} />
          ))}
        </section>
      ) : (
        <EmptyState
          title={
            base.search || stillValid
              ? "No matching schemas"
              : "No payload schemas yet"
          }
          description={
            base.search || stillValid
              ? "Clear or change the filters to broaden the results."
              : "Create a versioned schema for validated action payloads."
          }
        />
      )}
      <Pagination
        page={result.page}
        pathname="/admin/payload-schemas"
        searchParams={{
          limit: String(base.limit),
          search: base.search,
          still_valid: stillValid ? "1" : undefined,
        }}
      />
    </main>
  );
}
