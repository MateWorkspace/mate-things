"use server";

import { ApiError } from "@/lib/api/client";
import {
  createApiKey,
  deleteApiKey,
  regenerateApiKey,
  revokeApiKey,
} from "@/lib/api/api-keys";
import { requireSessionContext } from "@/lib/session";

import type { ApiKeyActionState } from "./state";

function denied(): ApiKeyActionState {
  return {
    status: "error",
    title: "Permission denied",
    message: "You do not have permission to make this change.",
  };
}

function failure(error: unknown): ApiKeyActionState {
  return {
    status: "error",
    title: error instanceof ApiError ? error.title : "Something went wrong",
    message: error instanceof ApiError ? error.message : "Please try again.",
  };
}

function value(data: FormData, name: string): string {
  return String(data.get(name) ?? "").trim();
}

function isoExpiresAt(
  raw: string,
  errors: Record<string, string>,
): string | undefined {
  if (!raw) return undefined;
  const parsed = new Date(raw);
  if (Number.isNaN(parsed.valueOf())) {
    errors.expires_at = "Enter a valid date and time.";
    return undefined;
  }
  return parsed.toISOString();
}

export async function generateApiKeyAction(
  _previous: ApiKeyActionState,
  formData: FormData,
): Promise<ApiKeyActionState> {
  const session = await requireSessionContext();
  if (!session.permissions.has("api_key:add")) return denied();

  const userId = value(formData, "user_id");
  const errors: Record<string, string> = {};
  if (!userId) errors.user_id = "Select a user.";
  const expiresAt = isoExpiresAt(value(formData, "expires_at"), errors);
  if (Object.keys(errors).length) {
    return {
      status: "error",
      title: "Check the form",
      message: "Complete the required fields.",
      fieldErrors: errors,
    };
  }

  try {
    const result = await createApiKey({ user_id: userId, expires_at: expiresAt });
    return {
      status: "success",
      title: "API key generated",
      message: "Copy this key now - it won't be shown again.",
      key: result.key,
    };
  } catch (error) {
    return failure(error);
  }
}

export async function regenerateApiKeyAction(
  _previous: ApiKeyActionState,
  formData: FormData,
): Promise<ApiKeyActionState> {
  const session = await requireSessionContext();
  if (!session.permissions.has("api_key:set")) return denied();

  const id = value(formData, "id");
  const errors: Record<string, string> = {};
  const expiresAt = isoExpiresAt(value(formData, "expires_at"), errors);
  if (!id || Object.keys(errors).length) {
    return {
      status: "error",
      title: "Check the form",
      message: "Complete the required fields.",
      fieldErrors: errors,
    };
  }

  try {
    const result = await regenerateApiKey(id, { expires_at: expiresAt });
    return {
      status: "success",
      title: "API key regenerated",
      message: "Copy this key now - it won't be shown again.",
      key: result.key,
    };
  } catch (error) {
    return failure(error);
  }
}

export async function revokeApiKeyAction(
  _previous: ApiKeyActionState,
  formData: FormData,
): Promise<ApiKeyActionState> {
  const session = await requireSessionContext();
  if (!session.permissions.has("api_key:set")) return denied();

  const id = value(formData, "id");
  if (!id) return failure(new Error("Missing API key id"));

  try {
    await revokeApiKey(id);
    return {
      status: "success",
      title: "API key revoked",
      message: "This key can no longer authenticate. Regenerate it to reactivate.",
    };
  } catch (error) {
    return failure(error);
  }
}

export async function deleteApiKeyAction(
  _previous: ApiKeyActionState,
  formData: FormData,
): Promise<ApiKeyActionState> {
  const session = await requireSessionContext();
  if (!session.permissions.has("api_key:remove")) return denied();

  const id = value(formData, "id");
  const username = value(formData, "username");
  const confirmation = value(formData, "confirmation");
  if (!id || !username || confirmation !== username) {
    return {
      status: "error",
      title: "Username does not match",
      message: "Enter the exact username before deleting this API key.",
      fieldErrors: {
        confirmation: "The confirmation must exactly match the username.",
      },
    };
  }

  try {
    await deleteApiKey(id);
    return {
      status: "success",
      title: "API key deleted",
      message: "This key has been permanently removed.",
    };
  } catch (error) {
    return failure(error);
  }
}
