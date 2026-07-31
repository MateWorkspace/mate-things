"use server";

import { ApiError } from "@/lib/api/client";
import {
  updatePreferences,
  type PreferencesResource,
} from "@/lib/api/preferences";
import { requireSessionContext } from "@/lib/session";

import type { PreferencesActionState } from "./preferences-state";

export async function savePreferencesAction(
  _previous: PreferencesActionState,
  formData: FormData,
): Promise<PreferencesActionState> {
  const session = await requireSessionContext();
  if (!session.permissions.has("preferences:set")) {
    return {
      status: "error",
      message: "You do not have permission to update preferences.",
    };
  }

  const resource = String(
    formData.get("resource") ?? "",
  ) as PreferencesResource;
  const id = String(formData.get("id") ?? "").trim();
  const raw = String(formData.get("preferences") ?? "");
  const allowed: PreferencesResource[] = [
    "action",
    "firmware",
    "node",
    "node_class",
    "payload_schema",
    "permission",
    "role",
    "user",
  ];

  if (!id || !allowed.includes(resource)) {
    return { status: "error", message: "The resource identity is invalid." };
  }

  let preferences: Record<string, unknown>;
  try {
    const parsed: unknown = JSON.parse(raw);
    if (
      parsed === null ||
      Array.isArray(parsed) ||
      typeof parsed !== "object"
    ) {
      throw new Error("Preferences must be a JSON object.");
    }
    preferences = parsed as Record<string, unknown>;
  } catch (error) {
    return {
      status: "error",
      message: error instanceof Error ? error.message : "Enter valid JSON.",
      fieldErrors: { preferences: "Preferences must be a valid JSON object." },
    };
  }

  try {
    await updatePreferences(resource, id, preferences);
    return { status: "success", message: "Preferences saved." };
  } catch (error) {
    return {
      status: "error",
      message:
        error instanceof ApiError
          ? error.message
          : "Unable to save preferences. Please try again.",
    };
  }
}
