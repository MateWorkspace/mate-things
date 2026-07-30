"use server";

import { refresh } from "next/cache";

import type { ScopedDeleteState } from "@/components/records/ScopedDeleteDialog";
import {
  deleteActionLogs,
  type ActionLogFilterQuery,
} from "@/lib/api/action-logs";
import { ApiError } from "@/lib/api/client";
import { parseRecordFilters } from "@/lib/record-filters";
import { requireSessionContext } from "@/lib/session";

export async function deleteActionHistoryAction(
  _state: ScopedDeleteState,
  formData: FormData,
): Promise<ScopedDeleteState> {
  const session = await requireSessionContext();
  if (!session.permissions.has("action_log:remove")) {
    return {
      status: "error",
      message: "You do not have permission to delete action history.",
    };
  }
  const raw = Object.fromEntries(formData.entries()) as Record<string, string>;
  const parsed = parseRecordFilters(raw);
  if (parsed.error) return { status: "error", message: parsed.error };
  const filters: ActionLogFilterQuery = {
    executed_at_start: parsed.filters.start,
    executed_at_end: parsed.filters.end,
    action_id: parsed.filters.actionId,
    node_id: parsed.filters.nodeId,
  };
  const active = Object.values(filters).some(Boolean);
  const expected = active ? "DELETE" : "DELETE ALL";
  if (String(formData.get("confirmation")) !== expected) {
    return { status: "error", message: `Type ${expected} to confirm.` };
  }
  try {
    const result = await deleteActionLogs(filters);
    refresh();
    return {
      status: "success",
      count: result.count,
      message: `${result.count} action history records deleted.`,
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
