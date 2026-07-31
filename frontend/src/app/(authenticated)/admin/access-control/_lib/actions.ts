"use server";

import { ApiError } from "@/lib/api/client";
import {
  createPermission,
  deletePermission,
  updatePermission,
} from "@/lib/api/permissions";
import {
  assignRolePermission,
  createRole,
  deleteRole,
  getRolePermissions,
  revokeRolePermission,
  setDefaultRole,
  updateRole,
} from "@/lib/api/roles";
import { requireSessionContext } from "@/lib/session";

import type { AccessActionState } from "./state";

function denied(permission: string): AccessActionState {
  return {
    status: "error",
    title: "Permission denied",
    message: `${permission} is required for this change.`,
  };
}
function failure(error: unknown): AccessActionState {
  return {
    status: "error",
    title: error instanceof ApiError ? error.title : "Something went wrong",
    message:
      error instanceof ApiError
        ? error.message
        : "Please try again.",
  };
}
function text(data: FormData, name: string): string {
  return String(data.get(name) ?? "").trim();
}

export async function saveRoleAction(
  _previous: AccessActionState,
  data: FormData,
): Promise<AccessActionState> {
  const session = await requireSessionContext();
  const id = text(data, "role_id");
  const permission = id ? "role:set" : "role:add";
  if (!session.permissions.has(permission)) return denied(permission);
  const name = text(data, "name");
  const description = text(data, "description");
  if (!name)
    return {
      status: "error",
      title: "Name required",
      message: "Enter a role name.",
      fieldErrors: { name: "This field is required." },
    };
  try {
    if (id) await updateRole(id, { name, description });
    else await createRole({ name, description });
    return {
      status: "success",
      title: id ? "Role updated" : "Role created",
      message: "Role details were saved.",
    };
  } catch (error) {
    return failure(error);
  }
}

export async function removeRoleAction(
  _previous: AccessActionState,
  data: FormData,
): Promise<AccessActionState> {
  const session = await requireSessionContext();
  if (!session.permissions.has("role:remove")) return denied("role:remove");
  const id = text(data, "role_id");
  const name = text(data, "role_name");
  if (!id || text(data, "confirmation") !== name)
    return {
      status: "error",
      title: "Role name does not match",
      message: "Enter the exact role name.",
      fieldErrors: { confirmation: "Confirmation does not match." },
    };
  try {
    await deleteRole(id);
    return {
      status: "success",
      title: "Role deleted",
      message: `${name} was removed.`,
    };
  } catch (error) {
    return failure(error);
  }
}

export async function setDefaultRoleAction(
  _previous: AccessActionState,
  data: FormData,
): Promise<AccessActionState> {
  const session = await requireSessionContext();
  if (!session.permissions.has("role:set")) return denied("role:set");
  const id = text(data, "role_id");
  const name = text(data, "role_name");
  if (!id || text(data, "confirmation") !== name)
    return {
      status: "error",
      title: "Confirmation required",
      message: "Enter the exact role name to make it the default.",
      fieldErrors: { confirmation: "Confirmation does not match." },
    };
  try {
    await setDefaultRole(id);
    return {
      status: "success",
      title: "Default role updated",
      message: `${name} is now assigned to new users by default.`,
    };
  } catch (error) {
    return failure(error);
  }
}

export async function savePermissionAction(
  _previous: AccessActionState,
  data: FormData,
): Promise<AccessActionState> {
  const session = await requireSessionContext();
  const id = text(data, "permission_id");
  const permission = id ? "permission:set" : "permission:add";
  if (!session.permissions.has(permission)) return denied(permission);
  const name = text(data, "name");
  const description = text(data, "description");
  if (!name)
    return {
      status: "error",
      title: "Name required",
      message: "Enter a permission name.",
      fieldErrors: { name: "This field is required." },
    };
  try {
    if (id) await updatePermission(id, { name, description });
    else await createPermission({ name, description });
    return {
      status: "success",
      title: id ? "Permission updated" : "Permission created",
      message: "Permission details were saved.",
    };
  } catch (error) {
    return failure(error);
  }
}

export async function removePermissionAction(
  _previous: AccessActionState,
  data: FormData,
): Promise<AccessActionState> {
  const session = await requireSessionContext();
  if (!session.permissions.has("permission:remove"))
    return denied("permission:remove");
  const id = text(data, "permission_id");
  const name = text(data, "permission_name");
  if (!id || text(data, "confirmation") !== name)
    return {
      status: "error",
      title: "Permission name does not match",
      message: "Enter the exact permission name.",
      fieldErrors: { confirmation: "Confirmation does not match." },
    };
  try {
    await deletePermission(id);
    return {
      status: "success",
      title: "Permission deleted",
      message: `${name} was removed.`,
    };
  } catch (error) {
    return failure(error);
  }
}

export async function updateRoleAssignmentsAction(
  _previous: AccessActionState,
  data: FormData,
): Promise<AccessActionState> {
  const session = await requireSessionContext();
  const roleId = text(data, "role_id");
  if (!roleId || !session.permissions.has("role_permission:get"))
    return denied("role_permission:get");
  const current = new Set(
    (await getRolePermissions(roleId)).map((item) => item.id),
  );
  const desired = data.getAll("permission_ids").map(String);
  const desiredSet = new Set(desired);
  const assign = desired.filter((id) => !current.has(id));
  const revoke = [...current].filter((id) => !desiredSet.has(id));
  if (assign.length && !session.permissions.has("role_permission:add"))
    return denied("role_permission:add");
  if (revoke.length && !session.permissions.has("role_permission:remove"))
    return denied("role_permission:remove");
  const results = await Promise.allSettled([
    ...assign.map((id) => assignRolePermission(roleId, id)),
    ...revoke.map((id) => revokeRolePermission(roleId, id)),
  ]);
  const failures = results.filter(
    (result) => result.status === "rejected",
  ).length;
  if (failures) {
    return {
      status: "error",
      title: "Assignments partially updated",
      message: `${results.length - failures} changes succeeded and ${failures} failed. The current authoritative assignments were reloaded.`,
    };
  }
  return {
    status: "success",
    title: "Assignments updated",
    message: `${results.length} permission changes saved.`,
  };
}
