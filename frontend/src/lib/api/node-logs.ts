"use server";

import { apiFetch, buildQuery } from "@/lib/api/client";
import type { CountDataResponse, CountResponse } from "@/lib/api/types";

export type NodeLogLevel = "NONE" | "ERROR" | "WARN" | "INFO" | "DEBUG";

export interface NodeLogResponse {
  id: number;
  node_device_id: string;
  level: NodeLogLevel;
  tag: string;
  message: string;
  logged_at: string;
  created_at: string;
}

export interface NodeLogFilterQuery {
  /** ISO 8601 timestamps. */
  logged_at_start?: string;
  logged_at_end?: string;
  node_device_id?: string;
  level?: NodeLogLevel;
}

export async function listNodeLogs(
  query: NodeLogFilterQuery = {},
): Promise<CountDataResponse<NodeLogResponse>> {
  return apiFetch(`/node-logs${buildQuery(query)}`);
}

export async function deleteNodeLogs(
  query: NodeLogFilterQuery = {},
): Promise<CountResponse> {
  return apiFetch(`/node-logs${buildQuery(query)}`, { method: "DELETE" });
}
