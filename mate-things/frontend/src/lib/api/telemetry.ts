"use server";

import { apiFetch, buildQuery } from "@/lib/api/client";
import type { CountDataResponse, CountResponse } from "@/lib/api/types";

export interface TelemetryRecordResponse {
  id: number;
  node_device_id: string;
  metric_name: string;
  payload_schema_name: string;
  payload_schema_version: number;
  payload: Record<string, unknown>;
  recorded_at: string;
  created_at: string;
}

export interface TelemetryFilterQuery {
  /** ISO 8601 timestamps. */
  recorded_at_start?: string;
  recorded_at_end?: string;
  node_device_id?: string;
  metric_name?: string;
  payload_schema_name?: string;
  payload_schema_version?: number;
}

/** Filter-only, no pagination - see AGENTS.md's data-fetching note on the two list-response shapes. */
export async function listTelemetryRecords(
  query: TelemetryFilterQuery = {},
): Promise<CountDataResponse<TelemetryRecordResponse>> {
  return apiFetch(`/telemetry-records${buildQuery(query)}`);
}

export async function deleteTelemetryRecords(
  query: TelemetryFilterQuery = {},
): Promise<CountResponse> {
  return apiFetch(`/telemetry-records${buildQuery(query)}`, {
    method: "DELETE",
  });
}
