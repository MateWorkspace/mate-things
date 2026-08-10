import "server-only";

import { apiFetch, buildQuery } from "@/lib/api/client";
import type { PermissionResponse } from "@/lib/api/permissions";
import type {
  AuditFields,
  IdResponse,
  PageDataResponse,
  PageQuery,
} from "@/lib/api/types";

export interface RoleResponse extends AuditFields {
  id: string;
  name: string;
  description: string;
  is_default: boolean;
  preferences: Record<string, unknown>;
}

export interface CreateRoleRequest {
  name: string;
  description?: string;
}

export interface UpdateRoleRequest {
  name?: string;
  description?: string;
}

export async function listRoles(
  query: PageQuery = {},
): Promise<PageDataResponse<RoleResponse>> {
  return apiFetch(`/admin/roles${buildQuery(query)}`);
}

export async function getDefaultRole(): Promise<RoleResponse> {
  return apiFetch("/admin/roles/default");
}

export async function getRoleByName(name: string): Promise<RoleResponse> {
  return apiFetch(`/admin/roles/by-name/${encodeURIComponent(name)}`);
}

export async function getRolePermissions(
  id: string,
): Promise<PermissionResponse[]> {
  return apiFetch(`/admin/roles/${id}/permissions`);
}

export async function getRoleById(id: string): Promise<RoleResponse> {
  return apiFetch(`/admin/roles/${id}`);
}

export async function createRole(
  request: CreateRoleRequest,
): Promise<IdResponse> {
  return apiFetch("/admin/roles", { method: "POST", body: request });
}

export async function updateRole(
  id: string,
  request: UpdateRoleRequest,
): Promise<void> {
  return apiFetch(`/admin/roles/${id}`, { method: "PATCH", body: request });
}

export async function setDefaultRole(id: string): Promise<void> {
  return apiFetch(`/admin/roles/${id}/default`, { method: "PATCH" });
}

export async function deleteRole(id: string): Promise<void> {
  return apiFetch(`/admin/roles/${id}`, { method: "DELETE" });
}

export async function assignRolePermission(
  roleId: string,
  permissionId: string,
): Promise<IdResponse> {
  return apiFetch(`/admin/roles/${roleId}/permissions/${permissionId}`, {
    method: "POST",
  });
}

export async function revokeRolePermission(
  roleId: string,
  permissionId: string,
): Promise<void> {
  return apiFetch(`/admin/roles/${roleId}/permissions/${permissionId}`, {
    method: "DELETE",
  });
}
