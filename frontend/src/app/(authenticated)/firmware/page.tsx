import type { Metadata } from "next";
import { redirect } from "next/navigation";

import Pagination from "@/components/collection/Pagination";
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
import FirmwareFilters from "./_components/FirmwareFilters";
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

      <FirmwareFilters
        canReadNodeClasses={canReadClasses}
        search={query.search}
        nodeClassId={nodeClassId}
        nodeClassName={nodeClassId ? classNames.get(nodeClassId) : undefined}
      />

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
          description="Clear or change the filters to broaden the results."
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
