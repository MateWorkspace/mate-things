"use server";

import { redirect } from "next/navigation";

import { ApiError } from "@/lib/api/client";
import {
  assignNodeClassAction,
  createNodeClass,
  deleteNodeClass,
  getNodeClassActions,
  revokeNodeClassAction,
  updateNodeClass,
} from "@/lib/api/node-classes";
import { requireSessionContext } from "@/lib/session";
import type { ActionState } from "@/lib/forms/action-state";

export type FormActionState = ActionState<string>;

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

export const EMPTY_NODE_CLASS_ASSIGNMENT_STATE: AssignmentActionState = {
  status: "idle",
  appliedIds: [],
  failed: [],
};

function permissionDenied(): FormActionState {
  return {
    status: "error",
    title: "Permission denied",
    message: "You do not have permission to make this change.",
  };
}

function actionError(error: unknown): FormActionState {
  if (error instanceof ApiError) {
    return {
      status: "error",
      title: error.title,
      message: error.message,
    };
  }

  return {
    status: "error",
    title: "Something went wrong",
    message: "Please try again.",
  };
}

function assignmentError(error: unknown): AssignmentResult {
  const state = actionError(error);
  return {
    status: "error",
    title: state.title ?? "Something went wrong",
    message: state.message ?? "Please try again.",
    appliedIds: [],
    failed: [],
  };
}

function assignmentPermissionDenied(): AssignmentResult {
  const state = permissionDenied();
  return {
    status: "error",
    title: state.title ?? "Permission denied",
    message: state.message ?? "You do not have permission to make this change.",
    appliedIds: [],
    failed: [],
  };
}

function failedMessage(error: unknown): string {
  return error instanceof ApiError ? error.message : "Please try again.";
}

function requiredFieldErrors(
  fields: Record<string, string>,
): Record<string, string> {
  return Object.fromEntries(
    Object.entries(fields)
      .filter(([, value]) => !value)
      .map(([field]) => [field, "This field is required."]),
  );
}

export async function createNodeClassAction(
  _previousState: FormActionState,
  formData: FormData,
): Promise<FormActionState> {
  const session = await requireSessionContext();
  if (!session.permissions.has("node_class:add")) {
    return permissionDenied();
  }

  const name = String(formData.get("name") ?? "").trim();
  const description = String(formData.get("description") ?? "").trim();
  const fieldErrors = requiredFieldErrors({ name });

  if (Object.keys(fieldErrors).length > 0) {
    return {
      status: "error",
      title: "Check the node class",
      message: "Complete the required field.",
      fieldErrors,
    };
  }

  try {
    await createNodeClass({ name, description });
  } catch (error) {
    return actionError(error);
  }

  return {
    status: "success",
    title: "Node class created",
    message: `${name} is ready to use.`,
  };
}

export async function updateNodeClassAction(
  _previousState: FormActionState,
  formData: FormData,
): Promise<FormActionState> {
  const session = await requireSessionContext();
  if (!session.permissions.has("node_class:set")) {
    return permissionDenied();
  }

  const nodeClassId = String(formData.get("node_class_id") ?? "").trim();
  const name = String(formData.get("name") ?? "").trim();
  const description = String(formData.get("description") ?? "").trim();
  const fieldErrors = requiredFieldErrors({
    node_class_id: nodeClassId,
    name,
  });

  if (Object.keys(fieldErrors).length > 0) {
    return {
      status: "error",
      title: "Check the node class",
      message: "Complete the required fields.",
      fieldErrors,
    };
  }

  try {
    await updateNodeClass(nodeClassId, { name, description });
  } catch (error) {
    return actionError(error);
  }

  return {
    status: "success",
    title: "Node class updated",
    message: "The node class details have been saved.",
  };
}

export async function deleteNodeClassAction(
  _previousState: FormActionState,
  formData: FormData,
): Promise<FormActionState> {
  const session = await requireSessionContext();
  if (!session.permissions.has("node_class:remove")) {
    return permissionDenied();
  }

  const nodeClassId = String(formData.get("node_class_id") ?? "").trim();
  const nodeClassName = String(formData.get("node_class_name") ?? "").trim();
  const confirmation = String(formData.get("confirmation") ?? "");
  const fieldErrors = requiredFieldErrors({
    node_class_id: nodeClassId,
    node_class_name: nodeClassName,
  });

  if (Object.keys(fieldErrors).length > 0) {
    return {
      status: "error",
      title: "Check the node class",
      message: "The node class identity is incomplete. Refresh and try again.",
      fieldErrors,
    };
  }

  // This typed-name check is a UI safety interlock, not authorization.
  // Authorization is the remove permission above plus the backend DELETE
  // endpoint. Avoiding a class read keeps remove-only custom roles functional.
  if (confirmation !== nodeClassName) {
    return {
      status: "error",
      title: "Node class name does not match",
      message: "Enter the exact class name before deleting it.",
      fieldErrors: {
        confirmation: "The confirmation must exactly match the class name.",
      },
    };
  }

  try {
    await deleteNodeClass(nodeClassId);
  } catch (error) {
    return actionError(error);
  }

  redirect("/node-classes");
}

export async function updateNodeClassActionsAction(
  _previousState: AssignmentActionState,
  formData: FormData,
): Promise<AssignmentResult> {
  const session = await requireSessionContext();
  const nodeClassId = String(formData.get("node_class_id") ?? "").trim();
  if (!nodeClassId || !session.permissions.has("node_class_action:get")) {
    return assignmentPermissionDenied();
  }

  let current: Set<string>;
  try {
    current = new Set(
      (await getNodeClassActions(nodeClassId)).map((item) => item.id),
    );
  } catch (error) {
    return assignmentError(error);
  }
  const desired = formData.getAll("action_ids").map(String);
  const desiredSet = new Set(desired);
  const assign = desired.filter((id) => !current.has(id));
  const revoke = [...current].filter((id) => !desiredSet.has(id));

  if (assign.length && !session.permissions.has("node_class_action:add")) {
    return assignmentPermissionDenied();
  }
  if (revoke.length && !session.permissions.has("node_class_action:remove")) {
    return assignmentPermissionDenied();
  }

  const operations = [
    ...assign.map((id) => ({
      id,
      apply: () => assignNodeClassAction(nodeClassId, id),
    })),
    ...revoke.map((id) => ({
      id,
      apply: () => revokeNodeClassAction(nodeClassId, id),
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
    message: `${results.length} action changes saved.`,
    appliedIds,
    failed,
  };
}
