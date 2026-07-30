import { refresh } from "next/cache";
import { beforeEach, describe, expect, it, vi } from "vitest";

import { updateProfile, updateProfilePassword } from "@/lib/api/profile";
import { requireSessionContext } from "@/lib/session";
import { formData } from "@/test/form-data";
import { USER } from "@/test/fixtures";

import {
  changePasswordAction,
  saveProfileAction,
  type FormActionState,
} from "./profile-actions";

vi.mock("server-only", () => ({}));

vi.mock("next/cache", () => ({
  refresh: vi.fn(),
}));

vi.mock("@/lib/api/profile", () => ({
  updateProfile: vi.fn(),
  updateProfilePassword: vi.fn(),
}));

vi.mock("@/lib/session", () => ({
  requireSessionContext: vi.fn(),
}));

const IDLE_STATE: FormActionState = { status: "idle" };

function permit(...permissions: string[]) {
  vi.mocked(requireSessionContext).mockResolvedValue({
    user: USER,
    permissions: new Set(permissions),
  });
}

describe("profile actions", () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it("trims profile fields, updates the profile, and refreshes the route", async () => {
    permit("profile:set");

    const result = await saveProfileAction(
      IDLE_STATE,
      formData({
        name: "  Alex Morgan  ",
        username: "  alex  ",
        bio: "  Fleet operator  ",
      }),
    );

    expect(updateProfile).toHaveBeenCalledWith({
      name: "Alex Morgan",
      username: "alex",
      bio: "Fleet operator",
    });
    expect(refresh).toHaveBeenCalledOnce();
    expect(result.status).toBe("success");
  });

  it("rejects blank required profile fields before mutation", async () => {
    permit("profile:set");

    const result = await saveProfileAction(
      IDLE_STATE,
      formData({ name: " ", username: " ", bio: " Biography " }),
    );

    expect(result).toMatchObject({
      status: "error",
      fieldErrors: {
        name: expect.any(String),
        username: expect.any(String),
      },
    });
    expect(updateProfile).not.toHaveBeenCalled();
    expect(refresh).not.toHaveBeenCalled();
  });

  it("rechecks profile:set before updating a profile", async () => {
    permit("profile:get");

    const result = await saveProfileAction(
      IDLE_STATE,
      formData({ name: "Alex Morgan", username: "alex", bio: "" }),
    );

    expect(result).toMatchObject({
      status: "error",
      title: "Permission denied",
    });
    expect(updateProfile).not.toHaveBeenCalled();
  });

  it("rejects a password confirmation mismatch before mutation", async () => {
    permit("profile_security:set");

    const result = await changePasswordAction(
      IDLE_STATE,
      formData({
        current_password: "current-secret",
        new_password: "new-secret",
        confirm_password: "different-secret",
      }),
    );

    expect(result).toMatchObject({
      status: "error",
      fieldErrors: { confirm_password: expect.any(String) },
    });
    expect(updateProfilePassword).not.toHaveBeenCalled();
  });

  it("rechecks profile_security:set before changing a password", async () => {
    permit("profile:get");

    const result = await changePasswordAction(
      IDLE_STATE,
      formData({
        current_password: "current-secret",
        new_password: "new-secret",
        confirm_password: "new-secret",
      }),
    );

    expect(result).toMatchObject({
      status: "error",
      title: "Permission denied",
    });
    expect(updateProfilePassword).not.toHaveBeenCalled();
  });

  it("changes a password after confirmation and authorization", async () => {
    permit("profile_security:set");

    const result = await changePasswordAction(
      IDLE_STATE,
      formData({
        current_password: "current-secret",
        new_password: "new-secret",
        confirm_password: "new-secret",
      }),
    );

    expect(updateProfilePassword).toHaveBeenCalledWith({
      current_password: "current-secret",
      new_password: "new-secret",
    });
    expect(result.status).toBe("success");
  });
});
