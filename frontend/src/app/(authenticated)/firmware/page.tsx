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
import { listFirmwares } from "@/lib/api/firmwares";
import { listAllNodeClasses } from "@/lib/api/node-classes";
import {
  getOutOfRangePageRedirect,
  parsePageQuery,
} from "@/lib/collection-query";
import { requirePermission } from "@/lib/session";

import FirmwareCard from "./_components/FirmwareCard";
import FirmwareForm from "./_components/FirmwareForm";

export const metadata: Metadata = {
  title: "Firmware — Mate Things",
  description:
    "Manage versioned device firmware, binaries, compatibility, and configuration schemas.",
};

type RawSearchParams = Record<string, string | string[] | undefined>;

interface FirmwarePageProps {
  searchParams: Promise<RawSearchParams>;
}

function firstString(value: string | string[] | undefined): string | undefined {
  return Array.isArray(value) ? value[0] : value;
}

export default async function FirmwarePage({
  searchParams,
}: FirmwarePageProps) {
  const [{ permissions }, rawSearchParams] = await Promise.all([
    requirePermission("firmware:get"),
    searchParams,
  ]);
  const query = parsePageQuery(rawSearchParams);
  const nodeClassId = firstString(rawSearchParams.node_class_id)?.trim();
  const canReadClasses = permissions.has("node_class:get");
  const [firmwares, nodeClasses] = await Promise.all([
    listFirmwares({ ...query, node_class_id: nodeClassId || undefined }),
    canReadClasses ? listAllNodeClasses() : Promise.resolve(null),
  ]);
  const redirectTarget = getOutOfRangePageRedirect(
    "/firmware",
    rawSearchParams,
    firmwares.page,
  );
  if (redirectTarget) {
    redirect(redirectTarget);
  }
  const nodeClassOptions =
    nodeClasses?.map(({ id, name }) => ({ id, name })) ?? [];
  const classNames = new Map(
    nodeClassOptions.map((nodeClass) => [nodeClass.id, nodeClass.name]),
  );
  const firstPageHref = `/firmware?limit=${query.limit}`;

  return (
    <main className="mx-auto w-full max-w-7xl space-y-6 px-4 py-6 sm:px-6 lg:px-8">
      <PageHeader
        title="Firmware"
        description="Manage versioned device binaries and their configuration schemas."
        actions={
          permissions.has("firmware:add") ? (
            <FirmwareForm nodeClasses={nodeClassOptions} />
          ) : undefined
        }
      />

      <CollectionToolbar filterTitle="Filter firmware">
        <form
          action="/firmware"
          method="get"
          className="grid gap-3 sm:grid-cols-[minmax(0,1fr)_minmax(12rem,18rem)_auto] sm:items-end"
        >
          <input type="hidden" name="limit" value={query.limit} />
          <label>
            <span className="text-foreground/80 mb-1.5 block text-sm font-medium">
              Search firmware
            </span>
            <Input
              key={query.search ?? ""}
              name="search"
              type="search"
              defaultValue={query.search}
              placeholder="Search by firmware name"
            />
          </label>
          {canReadClasses ? (
            <label>
              <span className="text-foreground/80 mb-1.5 block text-sm font-medium">
                Node class
              </span>
              <select
                name="node_class_id"
                defaultValue={nodeClassId ?? ""}
                className="border-control-border bg-background text-foreground focus-visible:border-focus focus-visible:ring-focus min-h-11 w-full rounded-xl border px-3.5 py-2.5 text-sm focus-visible:ring-2 focus-visible:outline-none"
              >
                <option value="">All node classes</option>
                {nodeClassOptions.map((nodeClass) => (
                  <option key={nodeClass.id} value={nodeClass.id}>
                    {nodeClass.name}
                  </option>
                ))}
              </select>
            </label>
          ) : null}
          <Button type="submit" className="min-h-11 gap-2">
            <Search aria-hidden="true" className="size-4" />
            Apply
          </Button>
        </form>
      </CollectionToolbar>

      <p className="text-muted-foreground px-1 text-xs font-semibold tracking-wider uppercase">
        Showing {firmwares.data.length} of {firmwares.page.total_items} firmware
        records
      </p>

      {firmwares.data.length > 0 ? (
        <section
          aria-label="Firmware collection"
          className="grid grid-cols-1 gap-5 sm:grid-cols-2 xl:grid-cols-3"
        >
          {firmwares.data.map((firmware) => (
            <FirmwareCard
              key={firmware.id}
              firmware={firmware}
              nodeClasses={nodeClassOptions}
              nodeClassName={classNames.get(firmware.node_class_id)}
              permissions={[...permissions]}
            />
          ))}
        </section>
      ) : query.search || nodeClassId ? (
        <EmptyState
          title="No matching firmware"
          description="No firmware matches the current filters."
          action={
            <Link
              href={firstPageHref}
              className="bg-primary text-surface focus-visible:ring-focus focus-visible:ring-offset-background inline-flex min-h-11 items-center justify-center rounded-xl px-4 py-2.5 text-sm font-semibold focus-visible:ring-2 focus-visible:ring-offset-2 focus-visible:outline-none"
            >
              Clear filters
            </Link>
          }
        />
      ) : (
        <EmptyState
          title="No firmware yet"
          description="Upload a firmware binary to start managing compatible device releases."
        />
      )}

      <div className="border-border border-t pt-5">
        <Pagination
          page={firmwares.page}
          pathname="/firmware"
          searchParams={{
            limit: String(query.limit),
            search: query.search,
            node_class_id: nodeClassId,
          }}
        />
      </div>
    </main>
  );
}
