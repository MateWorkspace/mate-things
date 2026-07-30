"use server";

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

    for (const nodeClass of result.data) {
      if (!seenIds.has(nodeClass.id)) {
        seenIds.add(nodeClass.id);
        nodeClasses.push(nodeClass);
      }
    }

    const responseLimit = Math.max(1, result.page.limit);
    const totalPages = Math.ceil(result.page.total_items / responseLimit);
    if (page >= totalPages || result.data.length === 0) {
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
