"use server";

import { redirect } from "next/navigation";

import { ApiError } from "@/lib/api/client";
import {
  createInfraredDeviceType,
  createInfraredState,
  listInfraredStates,
  startInfraredRecordSession,
  type InfraredStateResponse,
  type InfraredStateType,
  type StartRecordSessionRequest,
} from "@/lib/api/infrared";
import { requireSessionContext } from "@/lib/session";

import type {
  NewRecordSessionActionState,
  WizardCreateActionState,
} from "./state";

function permissionDenied(): NewRecordSessionActionState {
  return {
    status: "error",
    title: "Permission denied",
    message: "You do not have permission to make this change.",
  };
}

function actionError(error: unknown): NewRecordSessionActionState {
  if (error instanceof ApiError) {
    return { status: "error", title: error.title, message: error.message };
  }
  return {
    status: "error",
    title: "Something went wrong",
    message: "Please try again.",
  };
}

export async function listStatesForDeviceTypeAction(
  deviceTypeId: string,
): Promise<InfraredStateResponse[]> {
  const { permissions } = await requireSessionContext();
  if (!permissions.has("infrared_reference:get") || !deviceTypeId) {
    return [];
  }
  return listInfraredStates(deviceTypeId);
}

export async function startRecordSessionAction(
  _previousState: NewRecordSessionActionState,
  formData: FormData,
): Promise<NewRecordSessionActionState> {
  const { permissions } = await requireSessionContext();
  if (!permissions.has("infrared_record_session:add")) {
    return permissionDenied();
  }

  let payload: StartRecordSessionRequest;
  try {
    payload = JSON.parse(
      String(formData.get("payload") ?? ""),
    ) as StartRecordSessionRequest;
  } catch {
    return {
      status: "error",
      title: "Something went wrong",
      message: "Could not read the form. Please try again.",
    };
  }

  if (
    !payload.node_id ||
    !payload.infrared_device_type_id ||
    !payload.brand?.trim() ||
    !payload.model?.trim() ||
    !payload.definitions?.length
  ) {
    return {
      status: "error",
      title: "Check the form",
      message: "Fill in every required field before starting.",
    };
  }

  let created;
  try {
    created = await startInfraredRecordSession(payload);
  } catch (error) {
    return actionError(error);
  }

  redirect(`/apps/infrared/record/${created.id}`);
}

export async function createDeviceTypeForWizardAction(
  _previousState: WizardCreateActionState,
  formData: FormData,
): Promise<WizardCreateActionState> {
  const { permissions } = await requireSessionContext();
  if (!permissions.has("infrared_reference:add")) {
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

  let created;
  try {
    created = await createInfraredDeviceType(name);
  } catch (error) {
    return actionError(error);
  }

  return {
    status: "success",
    title: "Device type created",
    message: `${name} is ready to use.`,
    created: { id: created.id, name },
  };
}

export async function createStateForWizardAction(
  _previousState: WizardCreateActionState,
  formData: FormData,
): Promise<WizardCreateActionState> {
  const { permissions } = await requireSessionContext();
  if (!permissions.has("infrared_reference:add")) {
    return permissionDenied();
  }

  const deviceTypeId = String(formData.get("device_type_id") ?? "").trim();
  const name = String(formData.get("name") ?? "").trim();
  const type = String(formData.get("type") ?? "");

  const fieldErrors: Record<string, string> = {};
  if (!name) fieldErrors.name = "This field is required.";
  if (type !== "ENUM" && type !== "RANGE") {
    fieldErrors.type = "Choose a type.";
  }
  if (Object.keys(fieldErrors).length) {
    return {
      status: "error",
      title: "Check the state",
      message: "Complete the required fields.",
      fieldErrors,
    };
  }

  let created;
  try {
    created = await createInfraredState(deviceTypeId, {
      name,
      type: type as InfraredStateType,
    });
  } catch (error) {
    return actionError(error);
  }

  return {
    status: "success",
    title: "State created",
    message: `${name} is ready to use.`,
    created: { id: created.id, name },
  };
}
