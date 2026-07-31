"use server";

import type { ScopedDeleteState } from "@/components/records/ScopedDeleteDialog";
import { ApiError } from "@/lib/api/client";
import {
  deleteTelemetryRecords,
  type TelemetryFilterQuery,
} from "@/lib/api/telemetry";
import { parseRecordFilters } from "@/lib/record-filters";
import { requireSessionContext } from "@/lib/session";

export async function deleteTelemetryAction(
  _state: ScopedDeleteState,
  formData: FormData,
): Promise<ScopedDeleteState> {
  const session = await requireSessionContext();
  if (!session.permissions.has("telemetry_record:remove")) {
    return {
      status: "error",
      message: "You do not have permission to delete telemetry.",
    };
  }
  const parsed = parseRecordFilters(
    Object.fromEntries(formData.entries()) as Record<string, string>,
  );
  if (parsed.error) return { status: "error", message: parsed.error };
  const query: TelemetryFilterQuery = {
    recorded_at_start: parsed.filters.start,
    recorded_at_end: parsed.filters.end,
    node_device_id: parsed.filters.nodeDeviceId,
    metric_name: parsed.filters.metricName,
    payload_schema_name: parsed.filters.payloadSchemaName,
    payload_schema_version: parsed.filters.payloadSchemaVersion,
  };
  const active = Object.values(query).some(
    (value) => value !== undefined && value !== "",
  );
  const expected = active ? "DELETE" : "DELETE ALL";
  if (String(formData.get("confirmation")) !== expected)
    return { status: "error", message: `Type ${expected} to confirm.` };
  try {
    const result = await deleteTelemetryRecords(query);
    return {
      status: "success",
      count: result.count,
      message: `${result.count} telemetry records deleted.`,
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
