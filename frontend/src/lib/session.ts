import "server-only";

import { redirect } from "next/navigation";
import { cache } from "react";

import { ApiError } from "@/lib/api/client";
import { getProfile } from "@/lib/api/profile";
import type { UserResponse } from "@/lib/api/users";

/**
 * The "secure" session check (a real backend call via getProfile()) - see
 * frontend/AGENTS.md and this feature's plan for why this exists alongside
 * proxy.ts's cookie/JWT-only "optimistic" check rather than instead of it.
 * Memoized per render pass via React's cache(), matching the Next.js DAL
 * pattern (node_modules/next/dist/docs/.../authentication.md).
 */
export const getSession = cache(async (): Promise<UserResponse | null> => {
  try {
    return await getProfile();
  } catch (err) {
    if (err instanceof ApiError && err.status === 401) {
      return null;
    }
    // A genuine failure (network, 5xx) is not "logged out" - let it
    // surface as a real error rather than silently bouncing to /login.
    throw err;
  }
});

/**
 * Called directly from protected page components (not layouts - Next.js
 * layouts don't re-render on client-side navigation, so an auth check
 * there wouldn't run on every route change; see this feature's plan).
 */
export async function requireSession(): Promise<UserResponse> {
  const session = await getSession();

  if (!session) {
    redirect("/login");
  }

  return session;
}
