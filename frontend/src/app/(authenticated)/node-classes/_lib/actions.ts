"use server";

import { refresh } from "next/cache";
import { redirect } from "next/navigation";

import { ApiError } from "@/lib/api/client";
import {
  createNodeClass,
  deleteNodeClass,
  updateNodeClass,
} from "@/lib/api/node-classes";
import { requireSessionContext } from "@/lib/session";

export interface FormActionState {
  status: "idle" | "success" | "error";
  title?: string;
  message?: string;
  fieldErrors?: Record<string, string>;
}

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
    refresh();
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
    refresh();
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
