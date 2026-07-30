import "server-only";

import { redirect } from "next/navigation";
import { cache } from "react";

import { ApiError } from "@/lib/api/client";
import { getProfile, getProfilePermissions } from "@/lib/api/profile";
import type { UserResponse } from "@/lib/api/users";
import type { PermissionName } from "@/lib/permissions";

export interface SessionContext {
  user: UserResponse;
  permissions: ReadonlySet<PermissionName>;
}

export const getSession = cache(async (): Promise<UserResponse | null> => {
  try {
    return await getProfile();
  } catch (err) {
    if (err instanceof ApiError && err.status === 401) {
      return null;
    }
    throw err;
  }
});

export async function requireSession(): Promise<UserResponse> {
  const session = await getSession();

  if (!session) {
    redirect("/login");
  }

  return session;
}

export const getSessionContext = cache(
  async (): Promise<SessionContext | null> => {
    const user = await getSession();

    if (!user) {
      return null;
    }

    const permissions = await getProfilePermissions();

    return {
      user,
      permissions: new Set(permissions.map((permission) => permission.name)),
    };
  },
);

export async function requireSessionContext(): Promise<SessionContext> {
  const session = await getSessionContext();

  if (!session) {
    redirect("/login");
  }

  return session;
}
