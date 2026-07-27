"use server";

import { apiFetch } from "@/lib/api/client";

export type PreferencesResource =
  | "action"
  | "firmware"
  | "node"
  | "node_class"
  | "payload_schema"
  | "permission"
  | "role"
  | "user";

/**
 * Sets arbitrary key/value preferences on any preference-bearing resource
 * (each of the response types elsewhere in this layer that has a
 * `preferences` field). `resource` must be one of the backend's supported
 * values - anything else is rejected as a validation error.
 */
export async function updatePreferences(
  resource: PreferencesResource,
  id: string,
  preferences: Record<string, unknown>,
): Promise<void> {
  return apiFetch(`/preferences/${resource}/${id}`, {
    method: "PATCH",
    body: { preferences },
  });
}
