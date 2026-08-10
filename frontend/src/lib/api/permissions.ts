import "server-only";

import { apiFetch, buildQuery } from "@/lib/api/client";
import { collectAllPages } from "@/lib/api/collect-all-pages";
import type {
  AuditFields,
  IdResponse,
  PageDataResponse,
  PageQuery,
} from "@/lib/api/types";

export interface PermissionResponse extends AuditFields {
  id: string;
  name: string;
  description: string;
  preferences: Record<string, unknown>;
}

export interface CreatePermissionRequest {
  name: string;
  description?: string;
}

export interface UpdatePermissionRequest {
  name?: string;
  description?: string;
}

export async function listPermissions(
  query: PageQuery = {},
): Promise<PageDataResponse<PermissionResponse>> {
  return apiFetch(`/admin/permissions${buildQuery(query)}`);
}

export async function listAllPermissions(
  query: Omit<PageQuery, "page"> = {},
): Promise<PermissionResponse[]> {
  return collectAllPages({
    fetchPage: async (page) => {
      const result = await listPermissions({ ...query, page });
      return {
        data: result.data,
        page: result.page.page,
        limit: result.page.limit,
        total: result.page.total_items,
      };
    },
    keyOf: (permission) => permission.id,
  });
}

export async function getPermissionByName(
  name: string,
): Promise<PermissionResponse> {
  return apiFetch(`/admin/permissions/by-name/${encodeURIComponent(name)}`);
}

export async function getPermissionById(
  id: string,
): Promise<PermissionResponse> {
  return apiFetch(`/admin/permissions/${id}`);
}

export async function createPermission(
  request: CreatePermissionRequest,
): Promise<IdResponse> {
  return apiFetch("/admin/permissions", { method: "POST", body: request });
}

export async function updatePermission(
  id: string,
  request: UpdatePermissionRequest,
): Promise<void> {
  return apiFetch(`/admin/permissions/${id}`, {
    method: "PATCH",
    body: request,
  });
}

export async function deletePermission(id: string): Promise<void> {
  return apiFetch(`/admin/permissions/${id}`, { method: "DELETE" });
}
