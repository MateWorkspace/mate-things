import type { Metadata } from "next";

import ScopedDeleteDialog from "@/components/records/ScopedDeleteDialog";
import RefreshBoundary from "@/components/refresh/RefreshBoundary";
import PageHeader from "@/components/ui/page-header";
import { EmptyState } from "@/components/ui/states";
import { listNodeLogs, type NodeLogLevel } from "@/lib/api/node-logs";
import { getNodeByDeviceId } from "@/lib/api/nodes";
import { getOptionalById } from "@/lib/api/optional";
import { activeFilterEntries, parseRecordFilters } from "@/lib/record-filters";
import { requirePermission } from "@/lib/session";

import NodeLogFilters from "./_components/NodeLogFilters";
import NodeLogTable from "./_components/NodeLogTable";
import { deleteNodeLogsAction } from "./_lib/actions";

export const metadata: Metadata = { title: "Node Logs — Mate Things" };
type RawSearchParams = Record<string, string | string[] | undefined>;
const LEVELS = new Set<NodeLogLevel>([
  "NONE",
  "ERROR",
  "WARN",
  "INFO",
  "DEBUG",
]);

export default async function NodeLogsPage({
  searchParams,
}: {
  searchParams: Promise<RawSearchParams>;
}) {
  const [raw, { permissions }] = await Promise.all([
    searchParams,
    requirePermission("node_log:get"),
  ]);
  const parsed = parseRecordFilters(raw);
  const rawLevel = typeof raw.level === "string" ? raw.level : "";
  const level = LEVELS.has(rawLevel as NodeLogLevel)
    ? (rawLevel as NodeLogLevel)
    : undefined;
  const levelError =
    rawLevel && !level ? "Choose a supported log level." : undefined;
  const canReadNodes = permissions.has("node:get");
  const nodeDefault =
    parsed.error || levelError || !canReadNodes || !parsed.filters.nodeDeviceId
      ? null
      : await getOptionalById(() =>
          getNodeByDeviceId(parsed.filters.nodeDeviceId!),
        );
  const query = {
    logged_at_start: parsed.filters.start,
    logged_at_end: parsed.filters.end,
    node_device_id: parsed.filters.nodeDeviceId,
    level,
  };
  const result =
    parsed.error || levelError
      ? { data: [], total_items: 0 }
      : await listNodeLogs(query);
  const deletionFilters = activeFilterEntries({
    start: parsed.filters.start,
    end: parsed.filters.end,
    node_device_id: parsed.filters.nodeDeviceId,
    level,
  });

  return (
    <main className="mx-auto w-full max-w-7xl space-y-6 px-4 py-6 sm:px-6 lg:px-8">
      <PageHeader
        title="Node Logs"
        description="Follow device events, warnings, and errors without losing fleet context."
        actions={
          permissions.has("node_log:remove") ? (
            <ScopedDeleteDialog
              action={deleteNodeLogsAction}
              filters={deletionFilters}
              label="node logs"
            />
          ) : undefined
        }
      />
      <NodeLogFilters
        canReadNodes={canReadNodes}
        start={parsed.filters.start}
        end={parsed.filters.end}
        nodeDeviceId={parsed.filters.nodeDeviceId}
        nodeName={nodeDefault?.name}
        level={level}
      />
      {parsed.error || levelError ? (
        <EmptyState
          title="Check the filters"
          description={parsed.error ?? levelError ?? "Invalid filters."}
        />
      ) : (
        <>
          <div className="flex flex-wrap justify-between gap-3 text-sm">
            <p>
              <strong>{result.total_items}</strong> log records
            </p>
          </div>
          <RefreshBoundary
            updatedAt={result.data[0]?.created_at}
            intervalMs={15_000}
          >
            {result.data.length ? (
              <NodeLogTable records={result.data} />
            ) : (
              <EmptyState
                title="No node logs found"
                description="Change the filters or wait for the device to publish a new event."
              />
            )}
          </RefreshBoundary>
        </>
      )}
    </main>
  );
}
