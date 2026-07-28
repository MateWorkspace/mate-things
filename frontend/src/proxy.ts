import { NextResponse } from "next/server";
import type { NextRequest } from "next/server";

import { decodeJwtExpiry } from "@/lib/jwt";

// Duplicated from src/lib/api/client.ts deliberately: Proxy has its own
// cookie API (request.cookies / response.cookies) distinct from
// next/headers's cookies(), and its own runtime constraints, so this stays
// a small, self-contained refresh call rather than importing the
// 'use server' src/lib/api/auth.ts action - see this feature's plan for
// why Proxy, not a Server Component or client.ts, is where refresh has to
// happen.
const ACCESS_TOKEN_COOKIE = "mate_access_token";
const REFRESH_TOKEN_COOKIE = "mate_refresh_token";
const API_VERSION_PATH = "/api/v1";

// Refresh this far ahead of actual expiry, so a request that's mid-flight
// when the token would otherwise lapse still succeeds.
const REFRESH_BUFFER_MS = 10_000;

const PROTECTED_PREFIXES = ["/dashboard"];
const AUTH_ONLY_ROUTES = ["/login"];

interface RefreshedTokens {
  access_token: string;
  refresh_token: string;
}

function isExpiredOrNear(expiry: Date | null): boolean {
  if (!expiry) {
    return true;
  }
  return expiry.getTime() - Date.now() <= REFRESH_BUFFER_MS;
}

async function tryRefresh(refreshToken: string): Promise<RefreshedTokens | null> {
  const base = process.env.API_BASE_URL;
  if (!base) {
    return null;
  }

  try {
    const response = await fetch(
      `${base.replace(/\/$/, "")}${API_VERSION_PATH}/auth/refresh`,
      {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ refresh_token: refreshToken }),
      },
    );

    if (!response.ok) {
      return null;
    }

    return (await response.json()) as RefreshedTokens;
  } catch {
    return null;
  }
}

function setSessionCookies(response: NextResponse, tokens: RefreshedTokens): void {
  const secure = process.env.NODE_ENV === "production";

  response.cookies.set(ACCESS_TOKEN_COOKIE, tokens.access_token, {
    httpOnly: true,
    secure,
    sameSite: "lax",
    path: "/",
    expires: decodeJwtExpiry(tokens.access_token) ?? undefined,
  });
  response.cookies.set(REFRESH_TOKEN_COOKIE, tokens.refresh_token, {
    httpOnly: true,
    secure,
    sameSite: "lax",
    path: "/",
    expires: decodeJwtExpiry(tokens.refresh_token) ?? undefined,
  });
}

function isProtectedRoute(pathname: string): boolean {
  return PROTECTED_PREFIXES.some(
    (prefix) => pathname === prefix || pathname.startsWith(`${prefix}/`),
  );
}

export async function proxy(request: NextRequest) {
  const { pathname } = request.nextUrl;

  let accessToken = request.cookies.get(ACCESS_TOKEN_COOKIE)?.value;
  const refreshToken = request.cookies.get(REFRESH_TOKEN_COOKIE)?.value;

  // Proactive refresh: this is the one place in the app that can durably
  // persist a refreshed token (Server Components can only read cookies,
  // never set them - see this feature's plan). Only costs a network call
  // when the access token is actually missing or near expiry; a request
  // with a still-fresh token skips this entirely.
  let refreshedTokens: RefreshedTokens | null = null;
  if (
    refreshToken &&
    isExpiredOrNear(accessToken ? decodeJwtExpiry(accessToken) : null)
  ) {
    refreshedTokens = await tryRefresh(refreshToken);
    accessToken = refreshedTokens?.access_token;
  }

  const hasSession = Boolean(
    accessToken && !isExpiredOrNear(decodeJwtExpiry(accessToken)),
  );

  let response: NextResponse;
  if (isProtectedRoute(pathname) && !hasSession) {
    response = NextResponse.redirect(new URL("/login", request.nextUrl));
  } else if (AUTH_ONLY_ROUTES.includes(pathname) && hasSession) {
    response = NextResponse.redirect(new URL("/dashboard", request.nextUrl));
  } else {
    response = NextResponse.next();
  }

  if (refreshedTokens) {
    setSessionCookies(response, refreshedTokens);
  }

  return response;
}

// Skip Next internals and static files - everything else (including "/",
// which isn't in PROTECTED_PREFIXES/AUTH_ONLY_ROUTES and so just passes
// through, still getting the proactive refresh above) runs through proxy.
export const config = {
  matcher: [
    "/((?!_next/static|_next/image|favicon.ico|.*\\.(?:svg|png|jpg|jpeg|gif|webp|ico)$).*)",
  ],
};
