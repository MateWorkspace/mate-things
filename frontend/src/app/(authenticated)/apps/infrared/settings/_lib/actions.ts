"use server";

import { ApiError } from "@/lib/api/client";
import {
  createInfraredDeviceType,
  deleteInfraredDeviceType,
} from "@/lib/api/infrared";
import { requireSessionContext } from "@/lib/session";

import type { InfraredSettingsActionState } from "./state";

function permissionDenied(): InfraredSettingsActionState {
  return {
    status: "error",
    title: "Permission denied",
    message: "You do not have permission to make this change.",
  };
}

function actionError(error: unknown): InfraredSettingsActionState {
  if (error instanceof ApiError) {
    return { status: "error", title: error.title, message: error.message };
  }

  return {
    status: "error",
    title: "Something went wrong",
    message: "Please try again.",
  };
}

export async function createDeviceTypeAction(
  _previousState: InfraredSettingsActionState,
  formData: FormData,
): Promise<InfraredSettingsActionState> {
  const session = await requireSessionContext();
  if (!session.permissions.has("infrared_reference:add")) {
    return permissionDenied();
  }

  const name = String(formData.get("name") ?? "").trim();
  if (!name) {
    return {
      status: "error",
      title: "Check the device type",
      message: "Complete the required field.",
      fieldErrors: { name: "This field is required." },
    };
  }

  try {
    await createInfraredDeviceType(name);
  } catch (error) {
    return actionError(error);
  }

  return {
    status: "success",
    title: "Device type created",
    message: `${name} is ready to use.`,
  };
}

export async function deleteDeviceTypeAction(
  _previousState: InfraredSettingsActionState,
  formData: FormData,
): Promise<InfraredSettingsActionState> {
  const session = await requireSessionContext();
  if (!session.permissions.has("infrared_reference:delete")) {
    return permissionDenied();
  }

  const id = String(formData.get("id") ?? "").trim();
  if (!id) {
    return {
      status: "error",
      title: "Check the device type",
      message: "The device type identity is incomplete. Refresh and try again.",
      fieldErrors: { id: "This field is required." },
    };
  }

  try {
    await deleteInfraredDeviceType(id);
  } catch (error) {
    return actionError(error);
  }

  return {
    status: "success",
    title: "Device type deleted",
    message: "The device type has been removed.",
  };
}
