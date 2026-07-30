"use server";

import { refresh } from "next/cache";
import { redirect } from "next/navigation";

import { ApiError } from "@/lib/api/client";
import {
  createFirmware,
  deleteFirmware,
  getFirmwareBinaryUrlById,
  listAvailableFirmwaresByNodeId,
  replaceFirmwareBinary,
  updateFirmware,
  type FirmwareConfigSchemaItem,
  type FirmwareResponse,
} from "@/lib/api/firmwares";
import { dispatchOtaByNodeId } from "@/lib/api/ota";
import { requireSessionContext } from "@/lib/session";

export interface FormActionState {
  status: "idle" | "success" | "error";
  title?: string;
  message?: string;
  fieldErrors?: Record<string, string>;
}

export interface DispatchOtaInput {
  nodeId: string;
  firmwareId: string;
}

const CONFIG_VALUE_TYPES = new Set(["string", "uint32", "bool"]);
const OTA_PAGE_LIMIT = 48;

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
      message:
        (error.status === 400 ? error.details?.trim() : undefined) ||
        error.message,
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

function firmwareFile(formData: FormData): {
  file?: File | Blob;
  error?: string;
} {
  const value = formData.get("file");

  if (!(value instanceof Blob) || value.size === 0) {
    return { error: "Select a non-empty firmware binary." };
  }

  return { file: value };
}

function configSchema(formData: FormData): {
  schema?: FirmwareConfigSchemaItem[];
  error?: string;
} {
  const keys = formData
    .getAll("schema_key")
    .map((value) => String(value).trim());
  const valueTypes = formData
    .getAll("schema_value_type")
    .map((value) => String(value).trim());

  if (keys.length !== valueTypes.length) {
    return { error: "Each configuration key needs a value type." };
  }

  const schema: FirmwareConfigSchemaItem[] = [];
  for (const [index, key] of keys.entries()) {
    const valueType = valueTypes[index];
    const emptyRow = !key && !valueType;

    if (emptyRow) {
      continue;
    }

    if (!key || !CONFIG_VALUE_TYPES.has(valueType)) {
      return {
        error:
          "Each configuration row needs a key and a string, uint32, or bool value type.",
      };
    }

    schema.push({ key, value_type: valueType });
  }

  if (new Set(schema.map((item) => item.key)).size !== schema.length) {
    return { error: "Configuration keys must be unique." };
  }

  return { schema };
}

export async function createFirmwareAction(
  _previousState: FormActionState,
  formData: FormData,
): Promise<FormActionState> {
  const session = await requireSessionContext();
  if (!session.permissions.has("firmware:add")) {
    return permissionDenied();
  }

  const nodeClassId = String(formData.get("node_class_id") ?? "").trim();
  const name = String(formData.get("name") ?? "").trim();
  const fieldErrors = requiredFieldErrors({
    node_class_id: nodeClassId,
    name,
  });
  const fileResult = firmwareFile(formData);
  const schemaResult = configSchema(formData);

  if (fileResult.error) {
    fieldErrors.file = fileResult.error;
  }
  if (schemaResult.error) {
    fieldErrors.config_schema = schemaResult.error;
  }

  if (
    Object.keys(fieldErrors).length > 0 ||
    !fileResult.file ||
    !schemaResult.schema
  ) {
    return {
      status: "error",
      title: "Check the firmware",
      message: "Correct the highlighted fields and try again.",
      fieldErrors,
    };
  }

  try {
    await createFirmware(
      nodeClassId,
      name,
      fileResult.file,
      schemaResult.schema,
    );
    refresh();
  } catch (error) {
    return actionError(error);
  }

  return {
    status: "success",
    title: "Firmware uploaded",
    message: `${name} is ready to use.`,
  };
}

export async function updateFirmwareAction(
  _previousState: FormActionState,
  formData: FormData,
): Promise<FormActionState> {
  const session = await requireSessionContext();
  if (!session.permissions.has("firmware:set")) {
    return permissionDenied();
  }

  const firmwareId = String(formData.get("firmware_id") ?? "").trim();
  const nodeClassId = String(formData.get("node_class_id") ?? "").trim();
  const name = String(formData.get("name") ?? "").trim();
  const fieldErrors = requiredFieldErrors({
    firmware_id: firmwareId,
    node_class_id: nodeClassId,
    name,
  });

  if (Object.keys(fieldErrors).length > 0) {
    return {
      status: "error",
      title: "Check the firmware",
      message: "Complete the required fields.",
      fieldErrors,
    };
  }

  try {
    await updateFirmware(firmwareId, {
      node_class_id: nodeClassId,
      name,
    });
    refresh();
  } catch (error) {
    return actionError(error);
  }

  return {
    status: "success",
    title: "Firmware updated",
    message: "The firmware metadata has been saved.",
  };
}

export async function replaceFirmwareBinaryAction(
  _previousState: FormActionState,
  formData: FormData,
): Promise<FormActionState> {
  const session = await requireSessionContext();
  if (!session.permissions.has("firmware:set")) {
    return permissionDenied();
  }

  const firmwareId = String(formData.get("firmware_id") ?? "").trim();
  const fieldErrors = requiredFieldErrors({ firmware_id: firmwareId });
  const fileResult = firmwareFile(formData);
  const schemaResult = configSchema(formData);

  if (fileResult.error) {
    fieldErrors.file = fileResult.error;
  }
  if (schemaResult.error) {
    fieldErrors.config_schema = schemaResult.error;
  }

  if (
    Object.keys(fieldErrors).length > 0 ||
    !fileResult.file ||
    !schemaResult.schema
  ) {
    return {
      status: "error",
      title: "Check the replacement",
      message: "Correct the highlighted fields and try again.",
      fieldErrors,
    };
  }

  try {
    await replaceFirmwareBinary(
      firmwareId,
      fileResult.file,
      schemaResult.schema,
    );
    refresh();
  } catch (error) {
    return actionError(error);
  }

  return {
    status: "success",
    title: "Firmware binary replaced",
    message: "The new binary and configuration schema are active.",
  };
}

export async function deleteFirmwareAction(
  _previousState: FormActionState,
  formData: FormData,
): Promise<FormActionState> {
  const session = await requireSessionContext();
  if (!session.permissions.has("firmware:remove")) {
    return permissionDenied();
  }

  const firmwareId = String(formData.get("firmware_id") ?? "").trim();
  const firmwareName = String(formData.get("firmware_name") ?? "").trim();
  const confirmation = String(formData.get("confirmation") ?? "");
  const fieldErrors = requiredFieldErrors({
    firmware_id: firmwareId,
    firmware_name: firmwareName,
  });

  if (Object.keys(fieldErrors).length > 0) {
    return {
      status: "error",
      title: "Check the firmware",
      message: "The firmware identity is incomplete. Refresh and try again.",
      fieldErrors,
    };
  }

  if (confirmation !== firmwareName) {
    return {
      status: "error",
      title: "Firmware name does not match",
      message: "Enter the exact firmware name before deleting it.",
      fieldErrors: {
        confirmation: "The confirmation must exactly match the firmware name.",
      },
    };
  }

  try {
    await deleteFirmware(firmwareId);
  } catch (error) {
    return actionError(error);
  }

  redirect("/firmware");
}

async function findAvailableFirmware(
  nodeId: string,
  firmwareId: string,
): Promise<FirmwareResponse | undefined> {
  let pageNumber = 1;

  while (true) {
    const result = await listAvailableFirmwaresByNodeId(nodeId, {
      page: pageNumber,
      limit: OTA_PAGE_LIMIT,
    });
    const firmware = result.data.find((item) => item.id === firmwareId);

    if (firmware) {
      return firmware;
    }

    const totalPages = Math.max(
      1,
      Math.ceil(result.page.total_items / Math.max(1, result.page.limit)),
    );
    if (pageNumber >= totalPages || result.data.length === 0) {
      return undefined;
    }

    pageNumber += 1;
  }
}

export async function dispatchOtaAction(
  input: DispatchOtaInput,
): Promise<FormActionState> {
  const session = await requireSessionContext();
  if (!session.permissions.has("ota:dispatch")) {
    return permissionDenied();
  }

  if (
    !input ||
    typeof input.nodeId !== "string" ||
    typeof input.firmwareId !== "string"
  ) {
    return {
      status: "error",
      title: "Check the OTA request",
      message: "Select a node and compatible firmware.",
    };
  }

  const nodeId = input.nodeId.trim();
  const firmwareId = input.firmwareId.trim();
  if (!nodeId || !firmwareId) {
    return {
      status: "error",
      title: "Check the OTA request",
      message: "Select a node and compatible firmware.",
    };
  }

  try {
    const firmware = await findAvailableFirmware(nodeId, firmwareId);
    if (!firmware) {
      return {
        status: "error",
        title: "Firmware not available",
        message:
          "That firmware is not available for this node. Refresh and select a compatible firmware.",
      };
    }

    const firmwareUrl = await getFirmwareBinaryUrlById(firmware.id);
    await dispatchOtaByNodeId(nodeId, {
      firmware_id: firmware.id,
      firmware_url: firmwareUrl,
    });
    refresh();
  } catch (error) {
    return actionError(error);
  }

  return {
    status: "success",
    title: "OTA dispatched",
    message: "The firmware update request was sent to the node.",
  };
}
