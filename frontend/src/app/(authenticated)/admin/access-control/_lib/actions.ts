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

export type AssignmentResult = {
  status: "success" | "partial" | "error";
  appliedIds: string[];
  failed: Array<{ id: string; message: string }>;
  title: string;
  message: string;
};

export type AssignmentActionState =
  | AssignmentResult
  | {
      status: "idle";
      appliedIds: string[];
      failed: Array<{ id: string; message: string }>;
    };

export const EMPTY_ROLE_ASSIGNMENT_STATE: AssignmentActionState = {
  status: "idle",
  appliedIds: [],
  failed: [],
};

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
    message: error instanceof ApiError ? error.message : "Please try again.",
  };
}
function text(data: FormData, name: string): string {
  return String(data.get(name) ?? "").trim();
}

function assignmentFailure(error: unknown): AssignmentResult {
  const state = failure(error);
  return {
    status: "error",
    title: state.title ?? "Something went wrong",
    message: state.message ?? "Please try again.",
    appliedIds: [],
    failed: [],
  };
}

function assignmentDenied(permission: string): AssignmentResult {
  const state = denied(permission);
  return {
    status: "error",
    title: state.title ?? "Permission denied",
    message: state.message ?? `${permission} is required for this change.`,
    appliedIds: [],
    failed: [],
  };
}

function failedMessage(error: unknown): string {
  return error instanceof ApiError ? error.message : "Please try again.";
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
  _previous: AssignmentActionState,
  data: FormData,
): Promise<AssignmentResult> {
  const session = await requireSessionContext();
  const roleId = text(data, "role_id");
  if (!roleId || !session.permissions.has("role_permission:get"))
    return assignmentDenied("role_permission:get");
  let current: Set<string>;
  try {
    current = new Set(
      (await getRolePermissions(roleId)).map((item) => item.id),
    );
  } catch (error) {
    return assignmentFailure(error);
  }
  const desired = data.getAll("permission_ids").map(String);
  const desiredSet = new Set(desired);
  const assign = desired.filter((id) => !current.has(id));
  const revoke = [...current].filter((id) => !desiredSet.has(id));
  if (assign.length && !session.permissions.has("role_permission:add"))
    return assignmentDenied("role_permission:add");
  if (revoke.length && !session.permissions.has("role_permission:remove"))
    return assignmentDenied("role_permission:remove");
  const operations = [
    ...assign.map((id) => ({
      id,
      apply: () => assignRolePermission(roleId, id),
    })),
    ...revoke.map((id) => ({
      id,
      apply: () => revokeRolePermission(roleId, id),
    })),
  ];
  const results = await Promise.allSettled(
    operations.map((operation) => operation.apply()),
  );
  const appliedIds: string[] = [];
  const failed: Array<{ id: string; message: string }> = [];
  results.forEach((result, index) => {
    const { id } = operations[index];
    if (result.status === "fulfilled") appliedIds.push(id);
    else failed.push({ id, message: failedMessage(result.reason) });
  });
  if (failed.length) {
    return {
      status: appliedIds.length ? "partial" : "error",
      title: "Assignments partially updated",
      message: appliedIds.length
        ? `${appliedIds.length} changes succeeded and ${failed.length} failed. Failed changes remain selected for retry.`
        : `No changes were applied. ${failed.length} changes failed and remain selected for retry.`,
      appliedIds,
      failed,
    };
  }
  return {
    status: "success",
    title: "Assignments updated",
    message: `${results.length} permission changes saved.`,
    appliedIds,
    failed,
  };
}
