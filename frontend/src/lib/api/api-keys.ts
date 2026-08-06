"use server";

import { apiFetch, buildQuery } from "@/lib/api/client";
import type { PageDataResponse, PageQuery } from "@/lib/api/types";

export interface ApiKeyResponse {
  id: string;
  user_id: string;
  user_name: string;
  user_username: string;
  key_last_four: string;
  expires_at?: string;
  revoked_at?: string;
  created_at: string;
  updated_at?: string;
}

export interface ApiKeySecretResponse {
  key: string;
}

export type ApiKeyStatus = "active" | "inactive" | "any";

export interface ListApiKeysQuery extends PageQuery {
  status?: ApiKeyStatus;
}

export async function listApiKeys(
  query: ListApiKeysQuery = {},
): Promise<PageDataResponse<ApiKeyResponse>> {
  return apiFetch(`/admin/api-keys${buildQuery(query)}`);
}

export async function createApiKey(request: {
  user_id: string;
  expires_at?: string;
}): Promise<ApiKeySecretResponse> {
  return apiFetch("/admin/api-keys", { method: "POST", body: request });
}

export async function regenerateApiKey(
  id: string,
  request: { expires_at?: string },
): Promise<ApiKeySecretResponse> {
  return apiFetch(`/admin/api-keys/${id}/regenerate`, {
    method: "PATCH",
    body: request,
  });
}

export async function revokeApiKey(id: string): Promise<void> {
  return apiFetch(`/admin/api-keys/${id}/revoke`, { method: "PATCH" });
}

export async function deleteApiKey(id: string): Promise<void> {
  return apiFetch(`/admin/api-keys/${id}`, { method: "DELETE" });
}
