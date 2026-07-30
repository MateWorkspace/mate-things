"use server";

import { refresh } from "next/cache";
import { redirect } from "next/navigation";

import {
  createAction,
  deleteAction,
  dispatchAction,
  updateAction,
} from "@/lib/api/actions";
import { ApiError } from "@/lib/api/client";
import { requireSessionContext } from "@/lib/session";

export interface ActionFormState {
  status: "idle" | "success" | "error";
  title?: string;
  message?: string;
  executionId?: string;
  fieldErrors?: Record<string, string>;
}

function denied(): ActionFormState {
  return {
    status: "error",
    title: "Permission denied",
    message: "You do not have permission to perform this action.",
  };
}

function failure(error: unknown): ActionFormState {
  if (error instanceof ApiError) {
    return {
      status: "error",
      title: error.title,
      message: error.details?.trim() || error.message,
    };
  }
  return {
    status: "error",
    title: "Something went wrong",
    message: "Please try again.",
  };
}

function values(formData: FormData) {
  return {
    id: String(formData.get("action_id") ?? "").trim(),
    name: String(formData.get("name") ?? "").trim(),
    description: String(formData.get("description") ?? "").trim(),
    nodeClassId: String(formData.get("node_class_id") ?? "").trim(),
    schemaName: String(formData.get("payload_schema_name") ?? "").trim(),
    schemaVersion: Number(formData.get("payload_schema_version")),
  };
}

function validateAction(
  input: ReturnType<typeof values>,
): ActionFormState | null {
  const fieldErrors: Record<string, string> = {};
  if (!input.name) fieldErrors.name = "Name is required.";
  if (!input.nodeClassId) fieldErrors.node_class_id = "Node class is required.";
  if (!input.schemaName)
    fieldErrors.payload_schema_name = "Payload schema is required.";
  if (!Number.isInteger(input.schemaVersion) || input.schemaVersion <= 0)
    fieldErrors.payload_schema_version = "Select a valid schema version.";
  return Object.keys(fieldErrors).length
    ? {
        status: "error",
        title: "Check the action",
        message: "Complete the required fields.",
        fieldErrors,
      }
    : null;
}

export async function createActionFormAction(
  _state: ActionFormState,
  formData: FormData,
): Promise<ActionFormState> {
  const session = await requireSessionContext();
  if (!session.permissions.has("action:add")) return denied();
  const input = values(formData);
  const invalid = validateAction(input);
  if (invalid) return invalid;
  try {
    await createAction({
      name: input.name,
      description: input.description,
      node_class_id: input.nodeClassId,
      payload_schema_name: input.schemaName,
      payload_schema_version: input.schemaVersion,
    });
    refresh();
    return {
      status: "success",
      title: "Action created",
      message: `${input.name} is ready to dispatch.`,
    };
  } catch (error) {
    return failure(error);
  }
}

export async function updateActionFormAction(
  _state: ActionFormState,
  formData: FormData,
): Promise<ActionFormState> {
  const session = await requireSessionContext();
  if (!session.permissions.has("action:set")) return denied();
  const input = values(formData);
  const invalid = validateAction(input);
  if (invalid || !input.id)
    return (
      invalid ?? {
        status: "error",
        title: "Action unavailable",
        message: "Refresh and try again.",
      }
    );
  try {
    await updateAction(input.id, {
      name: input.name,
      description: input.description,
      node_class_id: input.nodeClassId,
      payload_schema_name: input.schemaName,
      payload_schema_version: input.schemaVersion,
    });
    refresh();
    return {
      status: "success",
      title: "Action updated",
      message: "The action definition has been saved.",
    };
  } catch (error) {
    return failure(error);
  }
}

export async function deleteActionFormAction(
  _state: ActionFormState,
  formData: FormData,
): Promise<ActionFormState> {
  const session = await requireSessionContext();
  if (!session.permissions.has("action:remove")) return denied();
  const id = String(formData.get("action_id") ?? "").trim();
  const name = String(formData.get("action_name") ?? "").trim();
  const confirmation = String(formData.get("confirmation") ?? "");
  if (!id || confirmation !== name) {
    return {
      status: "error",
      title: "Action name does not match",
      message: "Enter the exact action name before deleting it.",
      fieldErrors: { confirmation: "The confirmation does not match." },
    };
  }
  try {
    await deleteAction(id);
  } catch (error) {
    return failure(error);
  }
  redirect("/actions");
}

export async function dispatchActionFormAction(
  _state: ActionFormState,
  formData: FormData,
): Promise<ActionFormState> {
  const session = await requireSessionContext();
  if (!session.permissions.has("action:dispatch")) return denied();
  const actionId = String(formData.get("action_id") ?? "").trim();
  const nodeId = String(formData.get("node_id") ?? "").trim();
  const rawPayload = String(formData.get("payload") ?? "").trim();
  const executedAt = String(formData.get("executed_at") ?? "").trim();
  const fieldErrors: Record<string, string> = {};
  if (!actionId) fieldErrors.action_id = "Action is required.";
  if (!nodeId) fieldErrors.node_id = "Compatible node is required.";
  let payload: Record<string, unknown> = {};
  try {
    const parsed: unknown = JSON.parse(rawPayload || "{}");
    if (!parsed || Array.isArray(parsed) || typeof parsed !== "object") {
      throw new Error();
    }
    payload = parsed as Record<string, unknown>;
  } catch {
    fieldErrors.payload = "Payload must be a valid JSON object.";
  }
  let executedAtIso: string | undefined;
  if (executedAt) {
    const date = new Date(executedAt);
    if (Number.isNaN(date.getTime())) {
      fieldErrors.executed_at = "Execution time is invalid.";
    } else {
      executedAtIso = date.toISOString();
    }
  }
  if (Object.keys(fieldErrors).length) {
    return {
      status: "error",
      title: "Check the dispatch",
      message: "Correct the highlighted fields.",
      fieldErrors,
    };
  }
  try {
    const log = await dispatchAction(actionId, {
      node_id: nodeId,
      payload,
      executed_at: executedAtIso,
    });
    refresh();
    return {
      status: "success",
      title: "Action dispatched",
      message: "The execution is now visible in Action History.",
      executionId: log.execution_id,
    };
  } catch (error) {
    return failure(error);
  }
}
