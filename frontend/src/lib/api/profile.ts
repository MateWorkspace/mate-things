import "server-only";

import { apiFetch } from "@/lib/api/client";
import type { PermissionResponse } from "@/lib/api/permissions";
import type { UserResponse } from "@/lib/api/users";

export interface UpdateProfileRequest {
  name?: string;
  bio?: string;
  username?: string;
}

export interface UpdateProfilePasswordRequest {
  current_password: string;
  new_password: string;
}

/** The currently authenticated user, resolved from the access token. */
export async function getProfile(): Promise<UserResponse> {
  return apiFetch("/profile");
}

export async function getProfilePermissions(): Promise<PermissionResponse[]> {
  return apiFetch("/profile/permissions");
}

export async function updateProfile(
  request: UpdateProfileRequest,
): Promise<void> {
  return apiFetch("/profile", { method: "PATCH", body: request });
}

export async function updateProfilePassword(
  request: UpdateProfilePasswordRequest,
): Promise<void> {
  return apiFetch("/profile/password", { method: "PATCH", body: request });
}
