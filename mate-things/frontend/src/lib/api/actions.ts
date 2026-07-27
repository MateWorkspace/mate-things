"use server";

import type { ActionLogResponse } from "@/lib/api/action-logs";
import { apiFetch, buildQuery } from "@/lib/api/client";
import type { AuditFields, IdResponse, PageDataResponse, PageQuery } from "@/lib/api/types";

export interface ActionResponse extends AuditFields {
  id: string;
  node_class_id: string;
  name: string;
  description: string;
  payload_schema_name: string;
  payload_schema_version: number;
  preferences: Record<string, unknown>;
}

export interface CreateActionRequest {
  node_class_id: string;
  name: string;
  description?: string;
  payload_schema_name: string;
  payload_schema_version: number;
}

export interface UpdateActionRequest {
  node_class_id?: string;
  name?: string;
  description?: string;
  payload_schema_name?: string;
  payload_schema_version?: number;
}

export interface DispatchActionRequest {
  node_id: string;
  payload: Record<string, unknown>;
  /** ISO 8601 timestamp; defaults to now on the backend if omitted. */
  executed_at?: string;
}

export interface ListActionsQuery extends PageQuery {
  node_class_id?: string;
  payload_schema_name?: string;
  payload_schema_version?: number;
}

export async function listActions(
  query: ListActionsQuery = {},
): Promise<PageDataResponse<ActionResponse>> {
  return apiFetch(`/actions${buildQuery(query)}`);
}

export async function getActionByName(name: string): Promise<ActionResponse> {
  return apiFetch(`/actions/by-name/${encodeURIComponent(name)}`);
}

export async function getActionById(id: string): Promise<ActionResponse> {
  return apiFetch(`/actions/${id}`);
}

export async function createAction(
  request: CreateActionRequest,
): Promise<IdResponse> {
  return apiFetch("/actions", { method: "POST", body: request });
}

export async function updateAction(
  id: string,
  request: UpdateActionRequest,
): Promise<void> {
  return apiFetch(`/actions/${id}`, { method: "PATCH", body: request });
}

export async function deleteAction(id: string): Promise<void> {
  return apiFetch(`/actions/${id}`, { method: "DELETE" });
}

/** Dispatches the action to a node over MQTT; returns the created action-log row. */
export async function dispatchAction(
  id: string,
  request: DispatchActionRequest,
): Promise<ActionLogResponse> {
  return apiFetch(`/actions/${id}/dispatch`, {
    method: "POST",
    body: request,
  });
}
