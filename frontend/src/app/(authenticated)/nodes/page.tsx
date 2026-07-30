import type { Metadata } from "next";

import CollectionToolbar from "@/components/collection/CollectionToolbar";
import Pagination from "@/components/collection/Pagination";
import PageHeader from "@/components/ui/page-header";
import {
  listFirmwares,
  listNodeClasses,
  listNodes,
  type ListNodesQuery,
} from "@/lib/api";
import { parsePageQuery } from "@/lib/collection-query";
import { requirePermission } from "@/lib/session";

import NodeCard from "./_components/NodeCard";
import NodeCollectionEmptyState from "./_components/NodeCollectionEmptyState";
import NodeFilters, { type ConnectionFilter } from "./_components/NodeFilters";

export const metadata: Metadata = {
  title: "Nodes — Mate Things",
};

type RawSearchParams = Record<string, string | string[] | undefined>;

interface NodesPageProps {
  searchParams: Promise<RawSearchParams>;
}

function firstString(value: string | string[] | undefined): string | undefined {
  const result = Array.isArray(value) ? value[0] : value;
  return result?.trim() || undefined;
}

function parseConnectionFilter(
  value: string | string[] | undefined,
): ConnectionFilter {
  const parsed = firstString(value);
  return parsed === "connected" || parsed === "disconnected" ? parsed : "all";
}

export default async function NodesPage({ searchParams }: NodesPageProps) {
  const { permissions } = await requirePermission("node:get");
  const rawSearchParams = await searchParams;
  const pageQuery = parsePageQuery(rawSearchParams);
  const nodeClassId = firstString(rawSearchParams.node_class_id);
  const firmwareId = firstString(rawSearchParams.firmware_id);
  const connection = parseConnectionFilter(rawSearchParams.connection);
  const query: ListNodesQuery = {
    ...pageQuery,
    node_class_id: nodeClassId,
    firmware_id: firmwareId,
  };
  const canReadClasses = permissions.has("node_class:get");
  const canReadFirmware = permissions.has("firmware:get");

  const [nodes, classes, firmwares] = await Promise.all([
    listNodes(query),
    canReadClasses ? listNodeClasses({ limit: 48 }) : null,
    canReadFirmware ? listFirmwares({ limit: 48 }) : null,
  ]);

  const classOptions =
    classes?.data.map(({ id, name }) => ({ id, name })) ?? [];
  const firmwareOptions =
    firmwares?.data.map(({ id, name }) => ({ id, name })) ?? [];
  const classNames = new Map(
    classOptions.map((nodeClass) => [nodeClass.id, nodeClass.name]),
  );
  const firmwareNames = new Map(
    firmwareOptions.map((firmware) => [firmware.id, firmware.name]),
  );
  const visibleNodes = nodes.data.filter((node) => {
    if (connection === "connected") {
      return node.is_connected;
    }

    if (connection === "disconnected") {
      return !node.is_connected;
    }

    return true;
  });
  const paginationParams: RawSearchParams = {
    limit: String(pageQuery.limit),
    search: pageQuery.search,
    node_class_id: nodeClassId,
    firmware_id: firmwareId,
    connection: connection === "all" ? undefined : connection,
  };
  const cardPermissions = [...permissions];
  const filterKey = JSON.stringify([
    pageQuery.search,
    nodeClassId,
    firmwareId,
    connection,
    pageQuery.limit,
  ]);

  return (
    <main className="mx-auto w-full max-w-7xl space-y-6 px-4 py-6 sm:px-6 lg:px-8">
      <PageHeader
        title="Nodes fleet"
        description="Monitor self-registered devices, connection health, class assignments, and deployed firmware."
      />

      <CollectionToolbar filterTitle="Filter nodes">
        <NodeFilters
          key={filterKey}
          classes={classOptions}
          connection={connection}
          firmwareId={firmwareId}
          firmwares={firmwareOptions}
          limit={pageQuery.limit}
          nodeClassId={nodeClassId}
          search={pageQuery.search}
          showClassFilter={canReadClasses}
          showFirmwareFilter={canReadFirmware}
        />
      </CollectionToolbar>

      <div className="px-1">
        <p className="text-muted-foreground text-xs font-semibold tracking-wider uppercase">
          {connection === "all"
            ? `Showing ${nodes.data.length} of ${nodes.page.total_items} nodes`
            : `Showing ${visibleNodes.length} of ${nodes.data.length} nodes on this page · ${nodes.page.total_items} total`}
        </p>
      </div>

      {visibleNodes.length > 0 ? (
        <section
          aria-label="Node fleet"
          className="grid grid-cols-1 gap-5 sm:grid-cols-2 xl:grid-cols-3 2xl:grid-cols-4"
        >
          {visibleNodes.map((node) => (
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
          connection={connection}
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
