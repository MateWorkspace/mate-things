import { Search } from "lucide-react";
import type { Metadata } from "next";
import Link from "next/link";
import { redirect } from "next/navigation";

import CollectionToolbar from "@/components/collection/CollectionToolbar";
import Pagination from "@/components/collection/Pagination";
import Button from "@/components/ui/button";
import Input from "@/components/ui/input";
import PageHeader from "@/components/ui/page-header";
import { EmptyState } from "@/components/ui/states";
import {
  getLatestPayloadSchema,
  getPayloadSchemaByNameAndVersion,
  listPayloadSchemas,
} from "@/lib/api/payload-schemas";
import {
  getOutOfRangePageRedirect,
  parsePageQuery,
} from "@/lib/collection-query";
import { requirePermission } from "@/lib/session";

import PayloadSchemaCard from "./_components/PayloadSchemaCard";
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
  const lookupName = String(raw.lookup_name ?? "").trim();
  const lookupVersion = String(raw.lookup_version ?? "").trim();
  if (lookupName) {
    const found = lookupVersion
      ? await getPayloadSchemaByNameAndVersion(
          lookupName,
          Number(lookupVersion),
        )
      : await getLatestPayloadSchema(lookupName);
    redirect(`/admin/payload-schemas/${found.id}`);
  }
  const base = parsePageQuery(raw);
  const rawValidAt = String(raw.valid_at ?? "").trim();
  const validAt =
    rawValidAt && !Number.isNaN(new Date(rawValidAt).valueOf())
      ? new Date(rawValidAt).toISOString()
      : undefined;
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
      <CollectionToolbar filterTitle="Filter payload schemas">
        <form
          action="/admin/payload-schemas"
          className="grid gap-3 md:grid-cols-[1fr_1fr_auto]"
        >
          <Input
            name="search"
            type="search"
            defaultValue={base.search}
            placeholder="Search schema names"
            aria-label="Search schemas"
          />
          <Input
            name="valid_at"
            type="datetime-local"
            defaultValue={rawValidAt}
            aria-label="Valid at"
          />
          <Button type="submit" className="gap-2">
            <Search className="size-4" aria-hidden="true" />
            Filter
          </Button>
        </form>
        <form
          action="/admin/payload-schemas"
          className="border-border mt-4 grid gap-3 border-t pt-4 md:grid-cols-[1fr_10rem_auto]"
        >
          <Input
            name="lookup_name"
            placeholder="Exact schema name"
            aria-label="Schema lookup name"
            required
          />
          <Input
            name="lookup_version"
            type="number"
            min={1}
            placeholder="Latest"
            aria-label="Schema version"
          />
          <Button type="submit" variant="secondary">
            Open exact/latest
          </Button>
        </form>
      </CollectionToolbar>
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
            base.search || validAt
              ? "No matching schemas"
              : "No payload schemas yet"
          }
          description={
            base.search || validAt
              ? "No schema version matches these filters."
              : "Create a versioned schema for validated action payloads."
          }
          action={
            base.search || validAt ? (
              <Link
                href="/admin/payload-schemas"
                className="text-primary font-semibold"
              >
                Clear filters
              </Link>
            ) : undefined
          }
        />
      )}
      <Pagination
        page={result.page}
        pathname="/admin/payload-schemas"
        searchParams={{
          limit: String(base.limit),
          search: base.search,
          valid_at: rawValidAt || undefined,
        }}
      />
    </main>
  );
}
