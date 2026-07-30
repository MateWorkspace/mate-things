import { NextRequest } from "next/server";
import { describe, expect, it } from "vitest";

import { GET } from "./route";

const FUTURE_JWT = "e30.eyJleHAiOjQxMDI0NDQ4MDB9.signature";

describe("invalid-session transition", () => {
  it("clears both httpOnly session cookies before a same-origin login redirect", () => {
    const request = new NextRequest(
      "http://0.0.0.0:3000/auth/invalid-session",
      {
        headers: {
          cookie: [
            `mate_access_token=${FUTURE_JWT}`,
            `mate_refresh_token=${FUTURE_JWT}`,
          ].join("; "),
        },
      },
    );

    const response = GET(request);
    const setCookies = response.headers.getSetCookie();

    expect(response.headers.get("location")).toBe("/login?sessionInvalid=1");
    expect(setCookies).toHaveLength(2);
    expect(setCookies).toEqual(
      expect.arrayContaining([
        expect.stringMatching(
          /^mate_access_token=; Path=\/; Expires=Thu, 01 Jan 1970 00:00:00 GMT/,
        ),
        expect.stringMatching(
          /^mate_refresh_token=; Path=\/; Expires=Thu, 01 Jan 1970 00:00:00 GMT/,
        ),
      ]),
    );
  });
});
