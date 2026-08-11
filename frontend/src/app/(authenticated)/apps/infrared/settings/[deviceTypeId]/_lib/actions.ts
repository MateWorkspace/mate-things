"use server";

import { ApiError } from "@/lib/api/client";
import {
  createInfraredState,
  deleteInfraredState,
  type InfraredStateType,
} from "@/lib/api/infrared";
import { requireSessionContext } from "@/lib/session";

import type { InfraredStateActionState } from "./state";

const STATE_TYPES = new Set<InfraredStateType>(["ENUM", "RANGE"]);

function permissionDenied(): InfraredStateActionState {
  return {
    status: "error",
    title: "Permission denied",
    message: "You do not have permission to make this change.",
  };
}

function actionError(error: unknown): InfraredStateActionState {
  if (error instanceof ApiError) {
    return { status: "error", title: error.title, message: error.message };
  }

  return {
    status: "error",
    title: "Something went wrong",
    message: "Please try again.",
  };
}

export async function createStateAction(
  _previousState: InfraredStateActionState,
  formData: FormData,
): Promise<InfraredStateActionState> {
  const session = await requireSessionContext();
  if (!session.permissions.has("infrared_reference:add")) {
    return permissionDenied();
  }

  const deviceTypeId = String(formData.get("device_type_id") ?? "").trim();
  const name = String(formData.get("name") ?? "").trim();
  const rawType = String(formData.get("type") ?? "").trim();
  const type = STATE_TYPES.has(rawType as InfraredStateType)
    ? (rawType as InfraredStateType)
    : undefined;

  const fieldErrors: Record<string, string> = {};
  if (!name) fieldErrors.name = "This field is required.";
  if (!type) fieldErrors.type = "Choose a valid state type.";
  if (!deviceTypeId) {
    return {
      status: "error",
      title: "Check the device type",
      message: "The device type identity is incomplete. Refresh and try again.",
    };
  }
  if (Object.keys(fieldErrors).length > 0) {
    return {
      status: "error",
      title: "Check the state",
      message: "Complete the required fields.",
      fieldErrors,
    };
  }

  try {
    await createInfraredState(deviceTypeId, { name, type: type! });
  } catch (error) {
    return actionError(error);
  }

  return {
    status: "success",
    title: "State created",
    message: `${name} is ready to use.`,
  };
}

export async function deleteStateAction(
  _previousState: InfraredStateActionState,
  formData: FormData,
): Promise<InfraredStateActionState> {
  const session = await requireSessionContext();
  if (!session.permissions.has("infrared_reference:delete")) {
    return permissionDenied();
  }

  const id = String(formData.get("id") ?? "").trim();
  if (!id) {
    return {
      status: "error",
      title: "Check the state",
      message: "The state identity is incomplete. Refresh and try again.",
      fieldErrors: { id: "This field is required." },
    };
  }

  try {
    await deleteInfraredState(id);
  } catch (error) {
    return actionError(error);
  }

  return {
    status: "success",
    title: "State deleted",
    message: "The state has been removed.",
  };
}
