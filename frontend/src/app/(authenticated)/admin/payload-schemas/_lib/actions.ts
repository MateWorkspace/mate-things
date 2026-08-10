"use server";

import { redirect } from "next/navigation";

import { ApiError } from "@/lib/api/client";
import {
  createPayloadSchema,
  deletePayloadSchema,
  updatePayloadSchema,
} from "@/lib/api/payload-schemas";
import { requireSessionContext } from "@/lib/session";
import { requiredString } from "@/lib/forms/parse";

import type { PayloadSchemaActionState } from "./state";

function denied(permission: string): PayloadSchemaActionState {
  return {
    status: "error",
    title: "Permission denied",
    message: `${permission} is required for this change.`,
  };
}
function failure(error: unknown): PayloadSchemaActionState {
  return {
    status: "error",
    title: error instanceof ApiError ? error.title : "Something went wrong",
    message: error instanceof ApiError ? error.message : "Please try again.",
  };
}
function value(data: FormData, name: string): string {
  return String(data.get(name) ?? "").trim();
}
function isoDate(
  raw: string,
  field: string,
  errors: Record<string, string>,
): string | undefined {
  if (!raw) return undefined;
  const parsed = new Date(raw);
  if (Number.isNaN(parsed.valueOf())) {
    errors[field] = "Enter a valid ISO date and time.";
    return undefined;
  }
  return parsed.toISOString();
}
function definition(
  raw: string,
  errors: Record<string, string>,
): Record<string, unknown> | undefined {
  try {
    const parsed: unknown = JSON.parse(raw);
    if (parsed === null || Array.isArray(parsed) || typeof parsed !== "object")
      throw new Error();
    return parsed as Record<string, unknown>;
  } catch {
    errors.definition = "Definition must be a valid JSON object.";
    return undefined;
  }
}

async function saveSchema(
  data: FormData,
  operation: { kind: "create" } | { kind: "update"; id: string },
): Promise<PayloadSchemaActionState> {
  const session = await requireSessionContext();
  const permission =
    operation.kind === "update" ? "payload_schema:set" : "payload_schema:add";
  if (!session.permissions.has(permission)) return denied(permission);
  const errors: Record<string, string> = {};
  const name = value(data, "name");
  const rawVersion = value(data, "version");
  const version = Number(rawVersion);
  if (!name) errors.name = "This field is required.";
  if (
    !/^\d+$/.test(rawVersion) ||
    !Number.isSafeInteger(version) ||
    version < 1
  )
    errors.version = "Version must be a positive integer.";
  const parsedDefinition = definition(value(data, "definition"), errors);
  const validFrom = isoDate(value(data, "valid_from"), "valid_from", errors);
  const validTo = isoDate(value(data, "valid_to"), "valid_to", errors);
  if (validFrom && validTo && new Date(validTo) <= new Date(validFrom))
    errors.valid_to = "Valid to must be later than valid from.";
  if (Object.keys(errors).length || !parsedDefinition)
    return {
      status: "error",
      title: "Check the schema",
      message: "Correct the highlighted fields.",
      fieldErrors: errors,
    };
  try {
    if (operation.kind === "update")
      await updatePayloadSchema(operation.id, {
        name,
        version,
        definition: parsedDefinition,
        valid_from: validFrom,
        valid_to: validTo,
      });
    else
      await createPayloadSchema({
        name,
        version,
        definition: parsedDefinition,
        valid_from: validFrom,
        valid_to: validTo,
      });
    return {
      status: "success",
      title: operation.kind === "update" ? "Schema updated" : "Schema created",
      message: `${name} v${version} was saved.`,
    };
  } catch (error) {
    return failure(error);
  }
}

export async function createPayloadSchemaAction(
  _previous: PayloadSchemaActionState,
  data: FormData,
): Promise<PayloadSchemaActionState> {
  return saveSchema(data, { kind: "create" });
}
export async function updatePayloadSchemaAction(
  expectedIdOrPrevious: string | PayloadSchemaActionState,
  previousOrData: PayloadSchemaActionState | FormData,
  boundData?: FormData,
): Promise<PayloadSchemaActionState> {
  const expectedId =
    typeof expectedIdOrPrevious === "string" ? expectedIdOrPrevious.trim() : "";
  const data =
    boundData ??
    (previousOrData instanceof FormData ? previousOrData : new FormData());
  const submittedId = requiredString(
    data,
    "payload_schema_id",
    "A payload schema ID is required.",
  );
  if (!expectedId || !submittedId.ok) {
    return {
      status: "error",
      title: "Cannot update schema",
      message: "A payload schema ID is required.",
      fieldErrors: {
        payload_schema_id: submittedId.ok
          ? "A trusted payload schema ID is required."
          : submittedId.error,
      },
    };
  }
  if (submittedId.value !== expectedId) {
    return {
      status: "error",
      title: "Cannot update schema",
      message: "The payload schema identity does not match this update.",
      fieldErrors: {
        payload_schema_id: "The payload schema identity does not match.",
      },
    };
  }

  return saveSchema(data, { kind: "update", id: expectedId });
}
export async function deletePayloadSchemaAction(
  _previous: PayloadSchemaActionState,
  data: FormData,
): Promise<PayloadSchemaActionState> {
  const session = await requireSessionContext();
  if (!session.permissions.has("payload_schema:remove"))
    return denied("payload_schema:remove");
  const id = value(data, "payload_schema_id");
  const identity = value(data, "identity");
  if (!id || value(data, "confirmation") !== identity)
    return {
      status: "error",
      title: "Schema identity does not match",
      message: "Enter the exact name and version.",
      fieldErrors: { confirmation: "Confirmation does not match." },
    };
  try {
    await deletePayloadSchema(id);
  } catch (error) {
    return failure(error);
  }
  redirect("/admin/payload-schemas");
}
