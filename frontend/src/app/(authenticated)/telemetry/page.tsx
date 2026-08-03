import type { Metadata } from "next";

import ScopedDeleteDialog from "@/components/records/ScopedDeleteDialog";
import RefreshBoundary from "@/components/refresh/RefreshBoundary";
import PageHeader from "@/components/ui/page-header";
import { EmptyState } from "@/components/ui/states";
import { getNodeByDeviceId } from "@/lib/api/nodes";
import { listTelemetryRecords } from "@/lib/api/telemetry";
import { activeFilterEntries, parseRecordFilters } from "@/lib/record-filters";
import { requirePermission } from "@/lib/session";

import TelemetryFilters from "./_components/TelemetryFilters";
import TelemetryTable from "./_components/TelemetryTable";
import { deleteTelemetryAction } from "./_lib/actions";

export const metadata: Metadata = { title: "Telemetry — Mate Things" };
type RawSearchParams = Record<string, string | string[] | undefined>;

export default async function TelemetryPage({
  searchParams,
}: {
  searchParams: Promise<RawSearchParams>;
}) {
  const [raw, { permissions }] = await Promise.all([
    searchParams,
    requirePermission("telemetry_record:get"),
  ]);
  const parsed = parseRecordFilters(raw);
  const nodeDefault =
    parsed.error || !parsed.filters.nodeDeviceId
      ? null
      : await getNodeByDeviceId(parsed.filters.nodeDeviceId).catch(() => null);
  const query = {
    recorded_at_start: parsed.filters.start,
    recorded_at_end: parsed.filters.end,
    node_device_id: parsed.filters.nodeDeviceId,
    metric_name: parsed.filters.metricName,
  };
  const result = parsed.error
    ? { data: [], total_items: 0 }
    : await listTelemetryRecords(query);
  const deletionFilters = activeFilterEntries({
    start: parsed.filters.start,
    end: parsed.filters.end,
    node_device_id: parsed.filters.nodeDeviceId,
    metric_name: parsed.filters.metricName,
  });

  return (
    <main className="mx-auto w-full max-w-7xl space-y-6 px-4 py-6 sm:px-6 lg:px-8">
      <PageHeader
        title="Telemetry"
        description="Explore device measurements and inspect their structured payloads."
        actions={
          permissions.has("telemetry_record:remove") ? (
            <ScopedDeleteDialog
              action={deleteTelemetryAction}
              filters={deletionFilters}
              label="telemetry"
            />
          ) : undefined
        }
      />
      <TelemetryFilters
        start={parsed.filters.start}
        end={parsed.filters.end}
        nodeDeviceId={parsed.filters.nodeDeviceId}
        nodeName={nodeDefault?.name}
        metricName={parsed.filters.metricName}
      />
      {parsed.error ? (
        <EmptyState title="Check the filters" description={parsed.error} />
      ) : (
        <>
          <div className="flex flex-wrap justify-between gap-3 text-sm">
            <p>
              <strong>{result.total_items}</strong> records
            </p>
          </div>
          <RefreshBoundary updatedAt={result.data[0]?.created_at}>
            {result.data.length ? (
              <TelemetryTable records={result.data} />
            ) : (
              <EmptyState
                title="No telemetry found"
                description="Change the filters or wait for the node to report new measurements."
              />
            )}
          </RefreshBoundary>
        </>
      )}
    </main>
  );
}
