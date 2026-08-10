import "server-only";

import type { ActionLogResponse } from "@/lib/api/action-logs";
import { apiFetch, buildQuery } from "@/lib/api/client";
import type {
  AuditFields,
  IdResponse,
  PageDataResponse,
  PageQuery,
} from "@/lib/api/types";

export interface ActionResponse extends AuditFields {
  id: string;
  name: string;
  description: string;
  payload_schema_name: string;
  payload_schema_version: number;
  preferences: Record<string, unknown>;
  compatible_node_class_count?: number;
}

export interface CreateActionRequest {
  name: string;
  description?: string;
  payload_schema_name: string;
  payload_schema_version: number;
}

export interface UpdateActionRequest {
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

const ACTION_OPTION_PAGE_LIMIT = 48;

export async function listAllActions(): Promise<ActionResponse[]> {
  const actions: ActionResponse[] = [];
  const seenIds = new Set<string>();
  let page = 1;

  while (true) {
    const result = await listActions({
      page,
      limit: ACTION_OPTION_PAGE_LIMIT,
    });

    let added = 0;
    for (const action of result.data) {
      if (!seenIds.has(action.id)) {
        seenIds.add(action.id);
        actions.push(action);
        added += 1;
      }
    }

    const responseLimit =
      Number.isSafeInteger(result.page.limit) && result.page.limit > 0
        ? result.page.limit
        : ACTION_OPTION_PAGE_LIMIT;
    const totalPages = Math.ceil(result.page.total_items / responseLimit);
    if (
      !Number.isSafeInteger(totalPages) ||
      page >= totalPages ||
      result.data.length === 0 ||
      added === 0
    ) {
      break;
    }

    page += 1;
  }

  return actions;
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
