import "server-only";

import { redirect } from "next/navigation";
import { cache } from "react";

import { ApiError } from "@/lib/api/client";
import { getProfile } from "@/lib/api/profile";
import type { UserResponse } from "@/lib/api/users";

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
