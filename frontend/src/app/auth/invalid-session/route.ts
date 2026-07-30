import type { NextRequest } from "next/server";
import { NextResponse } from "next/server";

import { SESSION_COOKIE_NAMES } from "@/lib/session/cookies";

export function GET(request: NextRequest): NextResponse {
  const response = NextResponse.redirect(
    new URL("/login?sessionInvalid=1", request.url),
  );

  for (const cookieName of SESSION_COOKIE_NAMES) {
    response.cookies.delete(cookieName);
  }

  return response;
}
