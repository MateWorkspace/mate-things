"use server";

import { apiFetch, buildQuery } from "@/lib/api/client";
import type { CountResponse, PageDataResponse, PageQuery } from "@/lib/api/types";

export type ActionStatus = "UNEXECUTED" | "UNRESPONDED" | "FAILED" | "SUCCESS";

export interface ActionLogResponse {
  id: number;
  execution_id: string;
  action_id: string;
  action_name: string;
  node_id?: string;
  node_device_id?: string;
  node_name?: string;
  action_status: ActionStatus;
  action_message?: string;
  payload: Record<string, unknown>;
  executed_at: string;
  created_at: string;
}

export interface ActionLogFilterQuery extends PageQuery {
  /** ISO 8601 timestamps. */
  executed_at_start?: string;
  executed_at_end?: string;
  action_id?: string;
  node_id?: string;
  status?: ActionStatus;
}

export async function listActionLogs(
  query: ActionLogFilterQuery = {},
): Promise<PageDataResponse<ActionLogResponse>> {
  return apiFetch(`/action-logs${buildQuery(query)}`);
}

export async function deleteActionLogs(
  query: ActionLogFilterQuery = {},
): Promise<CountResponse> {
  return apiFetch(`/action-logs${buildQuery(query)}`, { method: "DELETE" });
}
