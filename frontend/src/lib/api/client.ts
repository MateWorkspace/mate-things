import "server-only";

import { cookies } from "next/headers";

import { API_BASE_URL } from "@/config/env";
import type { ErrorResponse } from "@/lib/api/types";
import { ACCESS_TOKEN_COOKIE } from "@/lib/session/cookies";

// Server-only: this whole module (and every file under src/lib/api/) never
// runs in the browser, so the backend's address never needs to reach the
// client bundle - a plain env var, not NEXT_PUBLIC_*.
const API_VERSION_PATH = "/api/v1";

export class ApiError extends Error {
  readonly status: number;
  readonly title: string;

  constructor(status: number, title: string, message: string) {
    super(message);
    this.name = "ApiError";
    this.status = status;
    this.title = title;
  }
}

function getBaseUrl(): string {
  return `${API_BASE_URL.replace(/\/$/, "")}${API_VERSION_PATH}`;
}

/**
 * Builds a query string from a plain object, skipping undefined/null
 * values. Reused by every list function's page/limit/search/filter params.
 * Generic over `object` (not `Record<string, unknown>`) so callers can pass
 * a specific query-params interface without an index signature - see
 * https://github.com/microsoft/TypeScript/issues/15300.
 */
export function buildQuery<T extends object>(params: T): string {
  const search = new URLSearchParams();

  for (const [key, value] of Object.entries(params) as [string, unknown][]) {
    if (value === undefined || value === null || value === "") {
      continue;
    }
    search.set(key, String(value));
  }

  const query = search.toString();
  return query ? `?${query}` : "";
}

export interface ApiFetchOptions {
  method?: "GET" | "HEAD" | "POST" | "PATCH" | "PUT" | "DELETE";
  /** Plain object is JSON-encoded; FormData is passed through untouched. */
  body?: unknown;
  /** Skip attaching the Authorization header (only auth.ts's login needs this). */
  skipAuth?: boolean;
  cache?: RequestCache;
}

async function authHeader(skipAuth: boolean | undefined): Promise<HeadersInit> {
  if (skipAuth) {
    return {};
  }

  const cookieStore = await cookies();
  const token = cookieStore.get(ACCESS_TOKEN_COOKIE)?.value;

  return token ? { Authorization: `Bearer ${token}` } : {};
}

/**
 * Low-level request, returning the raw Response with no error-throwing or
 * body parsing. Use this only when a handler's behavior genuinely isn't a
 * plain JSON in/out call - e.g. firmwares.ts's binary GET, which responds
 * with a 302 + Location header instead of a JSON body. Every other caller
 * should use apiFetch below.
 */
export async function apiRequest(
  path: string,
  options: ApiFetchOptions = {},
): Promise<Response> {
  const { method = "GET", body, skipAuth, cache } = options;

  const headers: HeadersInit = {
    ...(await authHeader(skipAuth)),
  };

  let requestBody: BodyInit | undefined;
  if (body instanceof FormData) {
    requestBody = body;
  } else if (body !== undefined) {
    (headers as Record<string, string>)["Content-Type"] = "application/json";
    requestBody = JSON.stringify(body);
  }

  return fetch(`${getBaseUrl()}${path}`, {
    method,
    headers,
    body: requestBody,
    cache,
    redirect: "manual",
  });
}

/**
 * Standard JSON in/out request. Throws ApiError on any non-2xx response;
 * returns undefined for a 204 No Content response.
 */
export async function apiFetch<T>(
  path: string,
  options: ApiFetchOptions = {},
): Promise<T> {
  const response = await apiRequest(path, options);

  if (response.status === 204) {
    return undefined as T;
  }

  if (response.status >= 300 && response.status < 400) {
    // A 3xx here (redirect: 'manual') means a caller used apiFetch on an
    // endpoint that actually redirects (e.g. a firmware binary GET) -
    // that should go through apiRequest directly instead.
    throw new ApiError(
      response.status,
      "Unexpected Redirect",
      `Unexpected redirect response for ${path} - use apiRequest directly for redirect-returning endpoints.`,
    );
  }

  if (!response.ok) {
    const errorBody = (await response
      .json()
      .catch(() => null)) as ErrorResponse | null;

    throw new ApiError(
      response.status,
      errorBody?.error ?? "Something Went Wrong",
      errorBody?.message ?? response.statusText,
    );
  }

  return (await response.json()) as T;
}
