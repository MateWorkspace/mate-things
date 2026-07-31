"use server";

import type { ScopedDeleteState } from "@/components/records/ScopedDeleteDialog";
import { ApiError } from "@/lib/api/client";
import {
  deleteNodeLogs,
  type NodeLogFilterQuery,
  type NodeLogLevel,
} from "@/lib/api/node-logs";
import { parseRecordFilters } from "@/lib/record-filters";
import { requireSessionContext } from "@/lib/session";

const LEVELS = new Set<NodeLogLevel>([
  "NONE",
  "ERROR",
  "WARN",
  "INFO",
  "DEBUG",
]);

export async function deleteNodeLogsAction(
  _state: ScopedDeleteState,
  formData: FormData,
): Promise<ScopedDeleteState> {
  const session = await requireSessionContext();
  if (!session.permissions.has("node_log:remove")) {
    return {
      status: "error",
      message: "You do not have permission to delete node logs.",
    };
  }
  const parsed = parseRecordFilters(
    Object.fromEntries(formData.entries()) as Record<string, string>,
  );
  if (parsed.error) return { status: "error", message: parsed.error };
  const rawLevel = String(formData.get("level") ?? "");
  const level = LEVELS.has(rawLevel as NodeLogLevel)
    ? (rawLevel as NodeLogLevel)
    : undefined;
  const query: NodeLogFilterQuery = {
    logged_at_start: parsed.filters.start,
    logged_at_end: parsed.filters.end,
    node_device_id: parsed.filters.nodeDeviceId,
    level,
  };
  const active = Object.values(query).some(Boolean);
  const expected = active ? "DELETE" : "DELETE ALL";
  if (String(formData.get("confirmation")) !== expected)
    return { status: "error", message: `Type ${expected} to confirm.` };
  try {
    const result = await deleteNodeLogs(query);
    return {
      status: "success",
      count: result.count,
      message: `${result.count} node log records deleted.`,
    };
  } catch (error) {
    return {
      status: "error",
      message:
        error instanceof ApiError
          ? error.message
          : "Deletion failed. Please try again.",
    };
  }
}
