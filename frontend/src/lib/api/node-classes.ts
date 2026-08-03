"use server";

import type { ActionResponse } from "@/lib/api/actions";
import { apiFetch, buildQuery } from "@/lib/api/client";
import type {
  AuditFields,
  IdResponse,
  PageDataResponse,
  PageQuery,
} from "@/lib/api/types";

export interface NodeClassResponse extends AuditFields {
  id: string;
  name: string;
  description: string;
  preferences: Record<string, unknown>;
}

export interface CreateNodeClassRequest {
  name: string;
  description?: string;
}

export interface UpdateNodeClassRequest {
  name?: string;
  description?: string;
}

export async function listNodeClasses(
  query: PageQuery = {},
): Promise<PageDataResponse<NodeClassResponse>> {
  return apiFetch(`/node-classes${buildQuery(query)}`);
}

const NODE_CLASS_OPTION_PAGE_LIMIT = 48;

export async function listAllNodeClasses(): Promise<NodeClassResponse[]> {
  const nodeClasses: NodeClassResponse[] = [];
  const seenIds = new Set<string>();
  let page = 1;

  while (true) {
    const result = await listNodeClasses({
      page,
      limit: NODE_CLASS_OPTION_PAGE_LIMIT,
    });

    let added = 0;
    for (const nodeClass of result.data) {
      if (!seenIds.has(nodeClass.id)) {
        seenIds.add(nodeClass.id);
        nodeClasses.push(nodeClass);
        added += 1;
      }
    }

    const responseLimit =
      Number.isSafeInteger(result.page.limit) && result.page.limit > 0
        ? result.page.limit
        : NODE_CLASS_OPTION_PAGE_LIMIT;
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

  return nodeClasses;
}

export async function getNodeClassByName(
  name: string,
): Promise<NodeClassResponse> {
  return apiFetch(`/node-classes/by-name/${encodeURIComponent(name)}`);
}

export async function getNodeClassById(id: string): Promise<NodeClassResponse> {
  return apiFetch(`/node-classes/${id}`);
}

export async function createNodeClass(
  request: CreateNodeClassRequest,
): Promise<IdResponse> {
  return apiFetch("/node-classes", { method: "POST", body: request });
}

export async function updateNodeClass(
  id: string,
  request: UpdateNodeClassRequest,
): Promise<void> {
  return apiFetch(`/node-classes/${id}`, { method: "PATCH", body: request });
}

export async function deleteNodeClass(id: string): Promise<void> {
  return apiFetch(`/node-classes/${id}`, { method: "DELETE" });
}

export interface NodeClassActionResponse {
  id: string;
  node_class_id: string;
  action_id: string;
  created_at: string;
  created_by?: string;
}

export interface NodeClassActionDetailResponse {
  node_class_action: NodeClassActionResponse;
  node_class: NodeClassResponse;
  action: ActionResponse;
}

export interface ListNodeClassActionsQuery extends PageQuery {
  node_class_id?: string;
  action_id?: string;
}

export async function getNodeClassActions(
  nodeClassId: string,
): Promise<ActionResponse[]> {
  return apiFetch(`/node-classes/${nodeClassId}/actions`);
}

export async function assignNodeClassAction(
  nodeClassId: string,
  actionId: string,
): Promise<IdResponse> {
  return apiFetch(`/node-classes/${nodeClassId}/actions/${actionId}`, {
    method: "POST",
  });
}

export async function revokeNodeClassAction(
  nodeClassId: string,
  actionId: string,
): Promise<void> {
  return apiFetch(`/node-classes/${nodeClassId}/actions/${actionId}`, {
    method: "DELETE",
  });
}

export async function listNodeClassActions(
  query: ListNodeClassActionsQuery = {},
): Promise<PageDataResponse<NodeClassActionDetailResponse>> {
  return apiFetch(`/node-class-actions${buildQuery(query)}`);
}
