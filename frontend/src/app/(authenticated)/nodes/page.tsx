import type { Metadata } from "next";
import { redirect } from "next/navigation";

import Pagination from "@/components/collection/Pagination";
import PageHeader from "@/components/ui/page-header";
import {
  listAllFirmwares,
  listAllNodeClasses,
  listNodes,
  type ListNodesQuery,
} from "@/lib/api";
import {
  getOutOfRangePageRedirect,
  parsePageQuery,
} from "@/lib/collection-query";
import { requirePermission } from "@/lib/session";

import NodeCard from "./_components/NodeCard";
import NodeCollectionEmptyState from "./_components/NodeCollectionEmptyState";
import NodeFilters from "./_components/NodeFilters";

export const metadata: Metadata = {
  title: "Nodes — Mate Things",
  description:
    "Browse, filter, and monitor self-registered nodes across the device fleet.",
};

type RawSearchParams = Record<string, string | string[] | undefined>;

interface NodesPageProps {
  searchParams: Promise<RawSearchParams>;
}

function firstString(value: string | string[] | undefined): string | undefined {
  const result = Array.isArray(value) ? value[0] : value;
  return result?.trim() || undefined;
}

export default async function NodesPage({ searchParams }: NodesPageProps) {
  const { permissions } = await requirePermission("node:get");
  const rawSearchParams = await searchParams;
  const pageQuery = parsePageQuery(rawSearchParams);
  const nodeClassId = firstString(rawSearchParams.node_class_id);
  const firmwareId = firstString(rawSearchParams.firmware_id);
  const query: ListNodesQuery = {
    ...pageQuery,
    node_class_id: nodeClassId,
    firmware_id: firmwareId,
  };
  const canReadClasses = permissions.has("node_class:get");
  const canReadFirmware = permissions.has("firmware:get");

  const [nodes, classes, firmwares] = await Promise.all([
    listNodes(query),
    canReadClasses ? listAllNodeClasses() : null,
    canReadFirmware ? listAllFirmwares() : null,
  ]);
  const supportedSearchParams: RawSearchParams = {
    page: rawSearchParams.page,
    limit: String(pageQuery.limit),
    search: pageQuery.search,
    node_class_id: nodeClassId,
    firmware_id: firmwareId,
  };
  const redirectTarget = getOutOfRangePageRedirect(
    "/nodes",
    supportedSearchParams,
    nodes.page,
  );
  if (redirectTarget) {
    redirect(redirectTarget);
  }

  const classOptions = classes?.map(({ id, name }) => ({ id, name })) ?? [];
  const firmwareOptions =
    firmwares?.map(({ id, name }) => ({ id, name })) ?? [];
  const classNames = new Map(
    classOptions.map((nodeClass) => [nodeClass.id, nodeClass.name]),
  );
  const firmwareNames = new Map(
    firmwareOptions.map((firmware) => [firmware.id, firmware.name]),
  );
  const paginationParams: RawSearchParams = {
    limit: String(pageQuery.limit),
    search: pageQuery.search,
    node_class_id: nodeClassId,
    firmware_id: firmwareId,
  };
  const cardPermissions = [...permissions];

  return (
    <main className="mx-auto w-full max-w-7xl space-y-6 px-4 py-6 sm:px-6 lg:px-8">
      <PageHeader
        title="Nodes fleet"
        description="Monitor self-registered devices, connection health, class assignments, and deployed firmware."
      />

      <NodeFilters
        search={pageQuery.search}
        nodeClassId={nodeClassId}
        nodeClassName={nodeClassId ? classNames.get(nodeClassId) : undefined}
        firmwareId={firmwareId}
        firmwareName={firmwareId ? firmwareNames.get(firmwareId) : undefined}
        showClassFilter={canReadClasses}
        showFirmwareFilter={canReadFirmware}
      />

      <div className="px-1">
        <p className="text-muted-foreground text-xs font-semibold tracking-wider uppercase">
          Showing {nodes.data.length} of {nodes.page.total_items} nodes
        </p>
      </div>

      {nodes.data.length > 0 ? (
        <section
          aria-label="Node fleet"
          className="grid grid-cols-1 gap-5 sm:grid-cols-2 xl:grid-cols-3 2xl:grid-cols-4"
        >
          {nodes.data.map((node) => (
            <NodeCard
              key={node.id}
              firmwareName={firmwareNames.get(node.firmware_id)}
              node={node}
              nodeClassName={classNames.get(node.node_class_id)}
              permissions={cardPermissions}
            />
          ))}
        </section>
      ) : (
        <NodeCollectionEmptyState
          firmwareId={firmwareId}
          limit={pageQuery.limit}
          nodeClassId={nodeClassId}
          search={pageQuery.search}
          totalItems={nodes.page.total_items}
        />
      )}

      <div className="border-border border-t pt-5">
        <Pagination
          page={nodes.page}
          pathname="/nodes"
          searchParams={paginationParams}
        />
      </div>
    </main>
  );
}
