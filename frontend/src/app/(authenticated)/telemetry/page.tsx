import type { Metadata } from "next";

import RecordWindow from "@/components/records/RecordWindow";
import ScopedDeleteDialog from "@/components/records/ScopedDeleteDialog";
import RefreshBoundary from "@/components/refresh/RefreshBoundary";
import PageHeader from "@/components/ui/page-header";
import { EmptyState } from "@/components/ui/states";
import { listTelemetryRecords } from "@/lib/api/telemetry";
import { activeFilterEntries, parseRecordFilters } from "@/lib/record-filters";
import { requirePermission } from "@/lib/session";

import TelemetryCard from "./_components/TelemetryCard";
import TelemetryFilters from "./_components/TelemetryFilters";
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
  const query = {
    recorded_at_start: parsed.filters.start,
    recorded_at_end: parsed.filters.end,
    node_device_id: parsed.filters.nodeDeviceId,
    metric_name: parsed.filters.metricName,
    payload_schema_name: parsed.filters.payloadSchemaName,
    payload_schema_version: parsed.filters.payloadSchemaVersion,
  };
  const result = parsed.error
    ? { data: [], total_items: 0 }
    : await listTelemetryRecords(query);
  const deletionFilters = activeFilterEntries({
    start: parsed.filters.start,
    end: parsed.filters.end,
    node_device_id: parsed.filters.nodeDeviceId,
    metric_name: parsed.filters.metricName,
    payload_schema_name: parsed.filters.payloadSchemaName,
    payload_schema_version: parsed.filters.payloadSchemaVersion,
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
        metricName={parsed.filters.metricName}
        schemaName={parsed.filters.payloadSchemaName}
        schemaVersion={parsed.filters.payloadSchemaVersion}
      />
      {parsed.error ? (
        <EmptyState title="Check the filters" description={parsed.error} />
      ) : (
        <>
          <div className="flex flex-wrap justify-between gap-3 text-sm">
            <p>
              <strong>{result.total_items}</strong> records
            </p>
            <p className="text-muted-foreground">
              Bound the time range to keep high-volume views focused.
            </p>
          </div>
          <RefreshBoundary updatedAt={result.data[0]?.created_at}>
            {result.data.length ? (
              <RecordWindow
                records={result.data}
                getKey={(record) => record.id}
                renderRecord={(record) => <TelemetryCard record={record} />}
              />
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
