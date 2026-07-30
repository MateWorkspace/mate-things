import type { Metadata } from "next";

import ScopedDeleteDialog from "@/components/records/ScopedDeleteDialog";
import RecordWindow from "@/components/records/RecordWindow";
import RefreshBoundary from "@/components/refresh/RefreshBoundary";
import PageHeader from "@/components/ui/page-header";
import { EmptyState } from "@/components/ui/states";
import { listActionLogs } from "@/lib/api/action-logs";
import { activeFilterEntries, parseRecordFilters } from "@/lib/record-filters";
import { requirePermission } from "@/lib/session";

import ActionLogCard from "./_components/ActionLogCard";
import ActionLogFilters from "./_components/ActionLogFilters";
import { deleteActionHistoryAction } from "./_lib/actions";

export const metadata: Metadata = { title: "Action History — Mate Things" };
type RawSearchParams = Record<string, string | string[] | undefined>;

export default async function ActionHistoryPage({
  searchParams,
}: {
  searchParams: Promise<RawSearchParams>;
}) {
  const [raw, { permissions }] = await Promise.all([
    searchParams,
    requirePermission("action_log:get"),
  ]);
  const parsed = parseRecordFilters(raw);
  const query = {
    executed_at_start: parsed.filters.start,
    executed_at_end: parsed.filters.end,
    action_id: parsed.filters.actionId,
    node_id: parsed.filters.nodeId,
  };
  const result = parsed.error
    ? { data: [], total_items: 0 }
    : await listActionLogs(query);
  const records = parsed.filters.executionId
    ? result.data.filter(
        (log) => log.execution_id === parsed.filters.executionId,
      )
    : result.data;
  const deletionFilters = activeFilterEntries({
    start: parsed.filters.start,
    end: parsed.filters.end,
    action_id: parsed.filters.actionId,
    node_id: parsed.filters.nodeId,
  });
  const latest = records[0]?.created_at;

  return (
    <main className="mx-auto w-full max-w-7xl space-y-6 px-4 py-6 sm:px-6 lg:px-8">
      <PageHeader
        title="Action History"
        description="Inspect execution outcomes and payloads across the fleet."
        actions={
          permissions.has("action_log:remove") &&
          !parsed.filters.executionId ? (
            <ScopedDeleteDialog
              action={deleteActionHistoryAction}
              filters={deletionFilters}
              label="action history"
            />
          ) : undefined
        }
      />
      <ActionLogFilters
        start={parsed.filters.start}
        end={parsed.filters.end}
        actionId={parsed.filters.actionId}
        nodeId={parsed.filters.nodeId}
        executionId={parsed.filters.executionId}
      />
      {parsed.error ? (
        <EmptyState title="Check the time range" description={parsed.error} />
      ) : (
        <>
          <div className="flex flex-wrap justify-between gap-3 text-sm">
            <p>
              <strong>
                {parsed.filters.executionId
                  ? records.length
                  : result.total_items}
              </strong>{" "}
              records
            </p>
            <p className="text-muted-foreground">
              Use a bounded time range for faster operational review.
            </p>
          </div>
          <RefreshBoundary updatedAt={latest}>
            {records.length ? (
              <RecordWindow
                records={records}
                getKey={(record) => record.id}
                renderRecord={(record) => <ActionLogCard log={record} />}
              />
            ) : (
              <EmptyState
                title="No action history found"
                description="Change the active filters or wait for a new action execution."
              />
            )}
          </RefreshBoundary>
        </>
      )}
    </main>
  );
}
