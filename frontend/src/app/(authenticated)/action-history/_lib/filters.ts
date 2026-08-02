import type { ActionStatus } from "@/lib/api/action-logs";

export interface ActionHistoryFilters {
  start?: string;
  end?: string;
  actionId?: string;
  nodeId?: string;
  status?: ActionStatus;
  executionId?: string;
}

export interface ParsedActionHistoryFilters {
  filters: ActionHistoryFilters;
  error?: string;
}

type RawActionHistoryParams = Record<string, string | string[] | undefined>;

const VALID_STATUSES: readonly ActionStatus[] = [
  "UNEXECUTED",
  "UNRESPONDED",
  "FAILED",
  "SUCCESS",
];

function first(value: string | string[] | undefined): string {
  return (Array.isArray(value) ? value[0] : (value ?? "")).trim();
}

function iso(
  value: string,
  field: string,
): { value?: string; error?: string } {
  if (!value) return {};
  const date = new Date(value);
  if (Number.isNaN(date.getTime())) {
    return { error: `${field} must be a valid date and time.` };
  }
  return { value: date.toISOString() };
}

export function parseActionHistoryFilters(
  raw: RawActionHistoryParams,
): ParsedActionHistoryFilters {
  const start = iso(first(raw.start), "Start");
  const end = iso(first(raw.end), "End");

  let error = start.error ?? end.error;
  if (
    !error &&
    start.value &&
    end.value &&
    new Date(start.value).getTime() > new Date(end.value).getTime()
  ) {
    error = "End must be after start.";
  }

  const statusRaw = first(raw.status).toUpperCase();
  const status = VALID_STATUSES.includes(statusRaw as ActionStatus)
    ? (statusRaw as ActionStatus)
    : undefined;
  if (!error && statusRaw && !status) {
    error = "Select a valid status.";
  }

  return {
    filters: {
      start: start.value,
      end: end.value,
      actionId: first(raw.action_id) || undefined,
      nodeId: first(raw.node_id) || undefined,
      status,
      executionId: first(raw.execution_id) || undefined,
    },
    error,
  };
}
