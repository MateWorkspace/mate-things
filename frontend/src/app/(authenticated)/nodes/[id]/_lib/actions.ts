"use server";

import { refresh } from "next/cache";
import { redirect } from "next/navigation";

import { ApiError } from "@/lib/api/client";
import { setNodeConfig } from "@/lib/api/node-config";
import { deleteNode, getNodeById, updateNode } from "@/lib/api/nodes";
import { requireSessionContext } from "@/lib/session";

export interface NodeActionState {
  status: "idle" | "success" | "error";
  title?: string;
  message?: string;
  fieldErrors?: Record<string, string>;
}

const UINT32_MAX = 4_294_967_295;

function permissionDenied(): NodeActionState {
  return {
    status: "error",
    title: "Permission denied",
    message: "You do not have permission to make this change.",
  };
}

function actionError(error: unknown): NodeActionState {
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

function requiredFields(
  fields: Record<string, string>,
): Record<string, string> {
  const errors: Record<string, string> = {};

  for (const [field, value] of Object.entries(fields)) {
    if (!value) {
      errors[field] = "This field is required.";
    }
  }

  return errors;
}

export async function saveNodeConfigAction(
  _previousState: NodeActionState,
  formData: FormData,
): Promise<NodeActionState> {
  const session = await requireSessionContext();
  if (!session.permissions.has("node_config:set")) {
    return permissionDenied();
  }

  const nodeId = String(formData.get("node_id") ?? "").trim();
  const key = String(formData.get("key") ?? "").trim();
  const valueType = String(formData.get("value_type") ?? "").trim();
  const valueEntry = formData.get("value");
  const value = typeof valueEntry === "string" ? valueEntry : "";
  const fieldErrors = requiredFields({
    node_id: nodeId,
    key,
    value_type: valueType,
  });

  if (typeof valueEntry !== "string") {
    fieldErrors.value = "This field is required.";
  } else if (valueType === "uint32") {
    if (!/^(0|[1-9]\d*)$/.test(value) || Number(value) > UINT32_MAX) {
      fieldErrors.value = "Enter a whole number from 0 to 4294967295.";
    }
  } else if (valueType === "bool") {
    if (value !== "true" && value !== "false") {
      fieldErrors.value = 'Choose either "true" or "false".';
    }
  } else if (valueType !== "string") {
    fieldErrors.value_type = "This configuration type is not supported.";
  }

  if (Object.keys(fieldErrors).length > 0) {
    return {
      status: "error",
      title: "Check the configuration value",
      message: "Complete the required field.",
      fieldErrors,
    };
  }

  try {
    await setNodeConfig(nodeId, key, value);
    refresh();
  } catch (error) {
    return actionError(error);
  }

  return {
    status: "success",
    title: "Configuration updated",
    message: `${key} was sent to the node.`,
  };
}

export async function saveNodeAction(
  _previousState: NodeActionState,
  formData: FormData,
): Promise<NodeActionState> {
  const session = await requireSessionContext();
  if (!session.permissions.has("node:set")) {
    return permissionDenied();
  }

  const nodeId = String(formData.get("node_id") ?? "").trim();
  const name = String(formData.get("name") ?? "").trim();
  const description = String(formData.get("description") ?? "").trim();
  const fieldErrors = requiredFields({
    node_id: nodeId,
    name,
  });

  if (Object.keys(fieldErrors).length > 0) {
    return {
      status: "error",
      title: "Check the node details",
      message: "Complete the required fields.",
      fieldErrors,
    };
  }

  try {
    await updateNode(nodeId, {
      name,
      description,
    });
    refresh();
  } catch (error) {
    return actionError(error);
  }

  return {
    status: "success",
    title: "Node updated",
    message: "The node details have been saved.",
  };
}

export async function deleteNodeAction(
  _previousState: NodeActionState,
  formData: FormData,
): Promise<NodeActionState> {
  const session = await requireSessionContext();
  if (!session.permissions.has("node:remove")) {
    return permissionDenied();
  }

  const nodeId = String(formData.get("node_id") ?? "").trim();
  const confirmation = String(formData.get("confirmation") ?? "");

  if (!nodeId) {
    return {
      status: "error",
      title: "Check the node",
      message: "A node ID is required.",
      fieldErrors: { node_id: "This field is required." },
    };
  }

  try {
    const node = await getNodeById(nodeId);

    if (confirmation !== node.device_id) {
      return {
        status: "error",
        title: "Device ID does not match",
        message: "Enter the exact device ID before deleting this node.",
        fieldErrors: {
          confirmation: "The confirmation must exactly match the device ID.",
        },
      };
    }

    await deleteNode(nodeId);
  } catch (error) {
    return actionError(error);
  }

  redirect("/nodes");
}
