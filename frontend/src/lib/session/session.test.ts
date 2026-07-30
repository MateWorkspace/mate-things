import { beforeEach, describe, expect, it, vi } from "vitest";

import { ApiError } from "@/lib/api/client";
import { getProfile, getProfilePermissions } from "@/lib/api/profile";
import { USER } from "@/test/fixtures";

import {
  getSession,
  requireAnyPermission,
  requirePermission,
  requireSessionContext,
} from "./session";

const cookieState = vi.hoisted(() => ({
  accessToken: "e30.eyJleHAiOjQxMDI0NDQ4MDB9.signature" as string | undefined,
}));

const navigation = vi.hoisted(() => ({
  forbidden: vi.fn((): never => {
    throw new Error("NEXT_FORBIDDEN");
  }),
  redirect: vi.fn((location: string): never => {
    throw new Error(`NEXT_REDIRECT:${location}`);
  }),
}));

vi.mock("server-only", () => ({}));

vi.mock("next/headers", () => ({
  cookies: vi.fn().mockResolvedValue({
    get: vi.fn((name: string) =>
      name === "mate_access_token" && cookieState.accessToken
        ? { value: cookieState.accessToken }
        : undefined,
    ),
  }),
}));

vi.mock("next/navigation", () => navigation);

vi.mock("@/lib/api/profile", () => ({
  getProfile: vi.fn(),
  getProfilePermissions: vi.fn(),
}));

describe("session authorization", () => {
  beforeEach(() => {
    vi.clearAllMocks();
    cookieState.accessToken = "e30.eyJleHAiOjQxMDI0NDQ4MDB9.signature";
    vi.mocked(getProfile).mockResolvedValue(USER);
    vi.mocked(getProfilePermissions).mockResolvedValue([
      {
        id: "permission-1",
        name: "node:get",
        description: "View nodes",
        preferences: {},
        created_at: "2026-07-30T00:00:00Z",
      },
    ]);
  });

  it("treats a request without an access cookie as anonymous", async () => {
    cookieState.accessToken = undefined;

    await expect(getSession()).resolves.toBeNull();
    expect(getProfile).not.toHaveBeenCalled();
  });

  it("routes a backend-rejected future-exp token through invalid-session cleanup", async () => {
    vi.mocked(getProfile).mockRejectedValue(
      new ApiError(401, "Unauthorized", "Token was rejected"),
    );

    await expect(requireSessionContext()).rejects.toThrow(
      "NEXT_REDIRECT:/auth/invalid-session",
    );
    expect(navigation.redirect).toHaveBeenCalledWith("/auth/invalid-session");
    expect(navigation.forbidden).not.toHaveBeenCalled();
  });

  it("renders forbidden when the backend rejects profile access", async () => {
    vi.mocked(getProfile).mockRejectedValue(
      new ApiError(403, "Access Denied", "Permission revoked"),
    );

    await expect(requireSessionContext()).rejects.toThrow("NEXT_FORBIDDEN");
    expect(navigation.forbidden).toHaveBeenCalledOnce();
    expect(navigation.redirect).not.toHaveBeenCalled();
  });

  it("renders forbidden when the backend rejects profile-permission access", async () => {
    vi.mocked(getProfilePermissions).mockRejectedValue(
      new ApiError(403, "Access Denied", "Permission revoked"),
    );

    await expect(requireSessionContext()).rejects.toThrow("NEXT_FORBIDDEN");
    expect(navigation.forbidden).toHaveBeenCalledOnce();
    expect(navigation.redirect).not.toHaveBeenCalled();
  });

  it("requires an exact page permission", async () => {
    await expect(requirePermission("node:get")).resolves.toMatchObject({
      user: USER,
    });

    await expect(requirePermission("node:set")).rejects.toThrow(
      "NEXT_FORBIDDEN",
    );
  });

  it("requires at least one page permission", async () => {
    await expect(
      requireAnyPermission(["firmware:get", "node:get"]),
    ).resolves.toMatchObject({ user: USER });

    await expect(
      requireAnyPermission(["firmware:get", "action:get"]),
    ).rejects.toThrow("NEXT_FORBIDDEN");
  });
});
