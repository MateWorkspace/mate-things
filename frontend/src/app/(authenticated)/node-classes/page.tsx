import type { Metadata } from "next";
import Link from "next/link";
import { redirect } from "next/navigation";

import Pagination from "@/components/collection/Pagination";
import PageHeader from "@/components/ui/page-header";
import { EmptyState } from "@/components/ui/states";
import { listNodeClasses } from "@/lib/api/node-classes";
import {
  getOutOfRangePageRedirect,
  parsePageQuery,
} from "@/lib/collection-query";
import { requirePermission } from "@/lib/session";

import NodeClassCard from "./_components/NodeClassCard";
import NodeClassFilters from "./_components/NodeClassFilters";
import NodeClassForm from "./_components/NodeClassForm";

export const metadata: Metadata = {
  title: "Node Classes — Mate Things",
  description:
    "Manage device compatibility groups and their related nodes, firmware, and actions.",
};

type RawSearchParams = Record<string, string | string[] | undefined>;

interface NodeClassesPageProps {
  searchParams: Promise<RawSearchParams>;
}

export default async function NodeClassesPage({
  searchParams,
}: NodeClassesPageProps) {
  const [{ permissions }, rawSearchParams] = await Promise.all([
    requirePermission("node_class:get"),
    searchParams,
  ]);
  const query = parsePageQuery(rawSearchParams);
  const nodeClasses = await listNodeClasses(query);
  const redirectTarget = getOutOfRangePageRedirect(
    "/node-classes",
    rawSearchParams,
    nodeClasses.page,
  );
  if (redirectTarget) {
    redirect(redirectTarget);
  }
  const cardPermissions = [...permissions];
  const firstPageHref = `/node-classes?limit=${query.limit}`;

  return (
    <main className="mx-auto w-full max-w-7xl space-y-6 px-4 py-6 sm:px-6 lg:px-8">
      <PageHeader
        title="Node classes"
        description="Organize compatible devices, firmware, and actions into fleet groups."
        actions={
          permissions.has("node_class:add") ? <NodeClassForm /> : undefined
        }
      />

      <NodeClassFilters search={query.search} />

      <div className="px-1">
        <p className="text-muted-foreground text-xs font-semibold tracking-wider uppercase">
          Showing {nodeClasses.data.length} of {nodeClasses.page.total_items}{" "}
          node classes
        </p>
      </div>

      {nodeClasses.data.length > 0 ? (
        <section
          aria-label="Node class collection"
          className="grid grid-cols-1 gap-5 sm:grid-cols-2 xl:grid-cols-3"
        >
          {nodeClasses.data.map((nodeClass) => (
            <NodeClassCard
              key={nodeClass.id}
              nodeClass={nodeClass}
              permissions={cardPermissions}
            />
          ))}
        </section>
      ) : query.search ? (
        <EmptyState
          title="No matching node classes"
          description="Clear or change the search to broaden the results."
        />
      ) : query.page > 1 ? (
        <EmptyState
          title="No node classes on this page"
          description="This page no longer contains results. Return to the first page to continue browsing."
          action={
            <Link
              className="bg-primary text-surface focus-visible:ring-focus focus-visible:ring-offset-background inline-flex min-h-11 items-center justify-center rounded-xl px-4 py-2.5 text-sm font-semibold transition-opacity hover:opacity-90 focus-visible:ring-2 focus-visible:ring-offset-2 focus-visible:outline-none"
              href={firstPageHref}
            >
              Return to first page
            </Link>
          }
        />
      ) : (
        <EmptyState
          title="No node classes yet"
          description="Create a node class to group compatible devices, firmware, and actions."
        />
      )}

      <div className="border-border border-t pt-5">
        <Pagination
          page={nodeClasses.page}
          pathname="/node-classes"
          searchParams={{
            limit: String(query.limit),
            search: query.search,
          }}
        />
      </div>
    </main>
  );
}
