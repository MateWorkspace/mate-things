/**
 * Pure, dependency-free JWT payload decoding - no signature verification.
 * That's the backend's job whenever the token is actually used against an
 * API call; this only reads the `exp` claim for optimistic freshness
 * checks (proxy.ts deciding whether to refresh, auth.ts setting a
 * cookie's real expiry). Safe to import from both proxy.ts (its own
 * module, Node.js runtime) and Server Actions/Components.
 */
export function decodeJwtExpiry(token: string): Date | null {
  const parts = token.split(".");
  if (parts.length !== 3) {
    return null;
  }

  try {
    const payloadJson = Buffer.from(parts[1], "base64url").toString("utf-8");
    const payload = JSON.parse(payloadJson) as { exp?: unknown };

    if (typeof payload.exp !== "number") {
      return null;
    }

    return new Date(payload.exp * 1000);
  } catch {
    return null;
  }
}
