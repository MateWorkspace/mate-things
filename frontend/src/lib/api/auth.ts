"use server";

import { cookies } from "next/headers";

import { IS_PRODUCTION } from "@/config/env";
import { apiFetch } from "@/lib/api/client";
import type { PermissionResponse } from "@/lib/api/permissions";
import type { RoleResponse } from "@/lib/api/roles";
import type { UserResponse } from "@/lib/api/users";
import {
  ACCESS_TOKEN_COOKIE,
  REFRESH_TOKEN_COOKIE,
} from "@/lib/session/cookies";
import { decodeJwtExpiry } from "@/lib/session/jwt";

export interface LoginRequest {
  username: string;
  password: string;
}

export interface RefreshRequest {
  refresh_token: string;
}

export interface LoginResponse {
  user: UserResponse;
  role: RoleResponse;
  permissions: PermissionResponse[];
  access_token: string;
  refresh_token: string;
}

async function persistSession(session: LoginResponse): Promise<void> {
  const cookieStore = await cookies();

  // Each cookie's expiry mirrors its own JWT's real `exp` claim rather
  // than a guessed duration - correct regardless of how BE_TOKEN_ACCESS_
  // DURATION/BE_TOKEN_REFRESH_DURATION are configured for a given
  // deployment. Falls back to a session cookie (no `expires`) only if the
  // token is somehow unparsable.
  const accessExpires = decodeJwtExpiry(session.access_token) ?? undefined;
  const refreshExpires = decodeJwtExpiry(session.refresh_token) ?? undefined;

  cookieStore.set(ACCESS_TOKEN_COOKIE, session.access_token, {
    httpOnly: true,
    secure: IS_PRODUCTION,
    sameSite: "lax",
    path: "/",
    expires: accessExpires,
  });
  cookieStore.set(REFRESH_TOKEN_COOKIE, session.refresh_token, {
    httpOnly: true,
    secure: IS_PRODUCTION,
    sameSite: "lax",
    path: "/",
    expires: refreshExpires,
  });
}

/**
 * Logs in against the backend and sets the httpOnly access/refresh-token
 * cookies on success. The only caller in this whole layer that doesn't
 * attach an Authorization header first - there's no token yet.
 */
export async function login(request: LoginRequest): Promise<LoginResponse> {
  const session = await apiFetch<LoginResponse>("/auth/login", {
    method: "POST",
    body: request,
    skipAuth: true,
  });

  await persistSession(session);

  return session;
}

/**
 * Exchanges the refresh token cookie for a new access/refresh token pair
 * and re-persists both cookies.
 */
export async function refresh(): Promise<LoginResponse> {
  const cookieStore = await cookies();
  const refreshToken = cookieStore.get(REFRESH_TOKEN_COOKIE)?.value;

  if (!refreshToken) {
    throw new Error("No refresh token cookie present");
  }

  const session = await apiFetch<LoginResponse>("/auth/refresh", {
    method: "POST",
    body: { refresh_token: refreshToken } satisfies RefreshRequest,
    skipAuth: true,
  });

  await persistSession(session);

  return session;
}

/** Clears the session cookies. There is no backend logout endpoint. */
export async function logout(): Promise<void> {
  const cookieStore = await cookies();
  cookieStore.delete(ACCESS_TOKEN_COOKIE);
  cookieStore.delete(REFRESH_TOKEN_COOKIE);
}
