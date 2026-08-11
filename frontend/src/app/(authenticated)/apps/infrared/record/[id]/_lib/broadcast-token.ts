"use server";

import { cookies } from "next/headers";

import { requireSessionContext } from "@/lib/session";
import { ACCESS_TOKEN_COOKIE } from "@/lib/session/cookies";

// The WebSocket handshake can't set an Authorization header, so the token
// has to travel as a query param the client can read - this is the one
// deliberate, narrow exception to "client JS never sees the access token".
// It's re-callable on every (re)connect attempt so a refreshed cookie is
// always picked up.
export async function getBroadcastToken(): Promise<string | null> {
  await requireSessionContext();
  const cookieStore = await cookies();
  return cookieStore.get(ACCESS_TOKEN_COOKIE)?.value ?? null;
}
