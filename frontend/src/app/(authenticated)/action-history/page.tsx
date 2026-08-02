import type { Metadata } from "next";
import { redirect } from "next/navigation";

import Pagination from "@/components/collection/Pagination";
import ScopedDeleteDialog from "@/components/records/ScopedDeleteDialog";
import RefreshBoundary from "@/components/refresh/RefreshBoundary";
import PageHeader from "@/components/ui/page-header";
import { EmptyState } from "@/components/ui/states";
import { getActionById } from "@/lib/api/actions";
import { listActionLogs } from "@/lib/api/action-logs";
import { getNodeById } from "@/lib/api/nodes";
import {
  getOutOfRangePageRedirect,
  parsePageQuery,
} from "@/lib/collection-query";
import { activeFilterEntries } from "@/lib/record-filters";
import { requirePermission } from "@/lib/session";

import ActionHistoryTable from "./_components/ActionHistoryTable";
import ActionLogFilters from "./_components/ActionLogFilters";
import { deleteActionHistoryAction } from "./_lib/actions";
import { parseActionHistoryFilters } from "./_lib/filters";

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
  const pageQuery = parsePageQuery(raw);
  const parsed = parseActionHistoryFilters(raw);

  const [actionDefault, nodeDefault] = await Promise.all([
    parsed.filters.actionId
      ? getActionById(parsed.filters.actionId).catch(() => null)
      : Promise.resolve(null),
    parsed.filters.nodeId
      ? getNodeById(parsed.filters.nodeId).catch(() => null)
      : Promise.resolve(null),
  ]);

  const result = parsed.error
    ? {
        data: [],
        page: { page: pageQuery.page, limit: pageQuery.limit, total_items: 0 },
      }
    : await listActionLogs({
        page: pageQuery.page,
        limit: pageQuery.limit,
        executed_at_start: parsed.filters.start,
        executed_at_end: parsed.filters.end,
        action_id: parsed.filters.actionId,
        node_id: parsed.filters.nodeId,
        status: parsed.filters.status,
      });

  if (!parsed.error) {
    const redirectTarget = getOutOfRangePageRedirect(
      "/action-history",
      raw,
      result.page,
    );
    if (redirectTarget) {
      redirect(redirectTarget);
    }
  }

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
    status: parsed.filters.status,
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
        actionName={actionDefault?.name}
        nodeId={parsed.filters.nodeId}
        nodeName={nodeDefault?.name}
        status={parsed.filters.status}
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
                  : result.page.total_items}
              </strong>{" "}
              records
            </p>
            <p className="text-muted-foreground">
              Use a bounded time range for faster operational review.
            </p>
          </div>
          <RefreshBoundary updatedAt={latest}>
            {records.length ? (
              <ActionHistoryTable records={records} />
            ) : (
              <EmptyState
                title="No action history found"
                description="Change the active filters or wait for a new action execution."
              />
            )}
          </RefreshBoundary>
          {!parsed.filters.executionId ? (
            <div className="border-border border-t pt-5">
              <Pagination
                page={result.page}
                pathname="/action-history"
                searchParams={raw}
              />
            </div>
          ) : null}
        </>
      )}
    </main>
  );
}
