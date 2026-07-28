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
