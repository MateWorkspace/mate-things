import type { Metadata } from "next";
import Link from "next/link";
import { redirect } from "next/navigation";

import Pagination from "@/components/collection/Pagination";
import PageHeader from "@/components/ui/page-header";
import { EmptyState } from "@/components/ui/states";
import {
  listInfraredDeviceTypes,
  listInfraredRecordSessions,
} from "@/lib/api/infrared";
import {
  getOutOfRangePageRedirect,
  parsePageQuery,
} from "@/lib/collection-query";
import { requirePermission } from "@/lib/session";

import RecordSessionFilters from "./_components/RecordSessionFilters";
import RecordSessionsTable from "./_components/RecordSessionsTable";
import { parseRecordSessionFilters } from "./_lib/filters";

export const metadata: Metadata = { title: "Infrared Record — Mate Things" };
type RawSearchParams = Record<string, string | string[] | undefined>;

export default async function InfraredRecordPage({
  searchParams,
}: {
  searchParams: Promise<RawSearchParams>;
}) {
  const [raw, { permissions }] = await Promise.all([
    searchParams,
    requirePermission("infrared_record_session:get"),
  ]);
  const pageQuery = parsePageQuery(raw);
  const parsed = parseRecordSessionFilters(raw);
  const canReadReferences = permissions.has("infrared_reference:get");

  const deviceTypes = canReadReferences ? await listInfraredDeviceTypes() : [];

  const result = parsed.error
    ? {
        data: [],
        page: { page: pageQuery.page, limit: pageQuery.limit, total_items: 0 },
      }
    : await listInfraredRecordSessions({
        page: pageQuery.page,
        limit: pageQuery.limit,
        recording_state: parsed.filters.recordingState,
        infrared_device_type_id: parsed.filters.deviceTypeId,
        created_at_start: parsed.filters.start,
        created_at_end: parsed.filters.end,
      });

  if (!parsed.error) {
    const redirectTarget = getOutOfRangePageRedirect(
      "/apps/infrared/record",
      raw,
      result.page,
    );
    if (redirectTarget) {
      redirect(redirectTarget);
    }
  }

  const records = result.data;

  return (
    <main className="mx-auto w-full max-w-7xl space-y-6 px-4 py-6 sm:px-6 lg:px-8">
      <PageHeader
        title="Record"
        description="Identify the device control by capturing its remote's infrared signals."
        actions={
          permissions.has("infrared_record_session:add") ? (
            <Link
              href="/apps/infrared/record/new"
              className="bg-primary text-surface focus-visible:ring-focus focus-visible:ring-offset-background inline-flex min-h-11 w-full items-center justify-center rounded-xl px-4 py-2.5 text-sm font-semibold transition-colors hover:opacity-90 focus-visible:ring-2 focus-visible:ring-offset-2 focus-visible:outline-none sm:w-auto"
            >
              Record New Device
            </Link>
          ) : null
        }
      />
      <RecordSessionFilters
        deviceTypes={deviceTypes}
        start={parsed.filters.start}
        end={parsed.filters.end}
        recordingState={parsed.filters.recordingState}
        deviceTypeId={parsed.filters.deviceTypeId}
      />
      {parsed.error ? (
        <EmptyState
          title="Check the active filters"
          description={parsed.error}
        />
      ) : (
        <>
          <div className="flex flex-wrap justify-between gap-3 text-sm">
            <p>
              <strong>{result.page.total_items}</strong> sessions
            </p>
          </div>
          {records.length ? (
            <RecordSessionsTable records={records} />
          ) : (
            <EmptyState
              title="No record sessions yet"
              description="Start recording a device to teach the Infrared app how to control it."
            />
          )}
          <div className="border-border border-t pt-5">
            <Pagination
              page={result.page}
              pathname="/apps/infrared/record"
              searchParams={raw}
            />
          </div>
        </>
      )}
    </main>
  );
}
