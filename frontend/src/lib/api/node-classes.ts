import "server-only";

import type { ActionResponse } from "@/lib/api/actions";
import { apiFetch, buildQuery } from "@/lib/api/client";
import { collectAllPages } from "@/lib/api/collect-all-pages";
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

export async function listAllNodeClasses(
  query: Omit<PageQuery, "page"> = {},
): Promise<NodeClassResponse[]> {
  return collectAllPages({
    fetchPage: async (page) => {
      const result = await listNodeClasses({ ...query, page });
      return {
        data: result.data,
        page: result.page.page,
        limit: result.page.limit,
        total: result.page.total_items,
      };
    },
    keyOf: (nodeClass) => nodeClass.id,
  });
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

export async function listAllNodeClassActions(
  query: Omit<ListNodeClassActionsQuery, "page"> = {},
): Promise<NodeClassActionDetailResponse[]> {
  return collectAllPages({
    fetchPage: async (page) => {
      const result = await listNodeClassActions({ ...query, page });
      return {
        data: result.data,
        page: result.page.page,
        limit: result.page.limit,
        total: result.page.total_items,
      };
    },
    keyOf: (item) => item.node_class_action.id,
  });
}
