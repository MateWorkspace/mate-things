import "server-only";

import { apiFetch, buildQuery } from "@/lib/api/client";
import type { PermissionResponse } from "@/lib/api/permissions";
import type { RoleResponse } from "@/lib/api/roles";
import type { PageDataResponse, PageQuery } from "@/lib/api/types";

export interface RolePermissionResponse {
  id: string;
  role_id: string;
  permission_id: string;
  created_at: string;
  created_by?: string;
}

export interface RolePermissionDetailResponse {
  role_permission: RolePermissionResponse;
  role: RoleResponse;
  permission: PermissionResponse;
}

export interface ListRolePermissionsQuery extends PageQuery {
  role_id?: string;
  permission_id?: string;
}

/**
 * Read-only: assigning/revoking a role-permission pair is done via
 * roles.ts's assignRolePermission/revokeRolePermission, which own the
 * mutating `/admin/roles/:role_id/permissions/:permission_id` routes.
 */
export async function listRolePermissions(
  query: ListRolePermissionsQuery = {},
): Promise<PageDataResponse<RolePermissionDetailResponse>> {
  return apiFetch(`/admin/role-permissions${buildQuery(query)}`);
}

export async function getRolePermissionByPair(
  roleId: string,
  permissionId: string,
): Promise<RolePermissionDetailResponse> {
  return apiFetch(
    `/admin/role-permissions/by-pair${buildQuery({ role_id: roleId, permission_id: permissionId })}`,
  );
}

export async function getRolePermissionById(
  id: string,
): Promise<RolePermissionDetailResponse> {
  return apiFetch(`/admin/role-permissions/${id}`);
}
