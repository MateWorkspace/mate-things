import { refresh } from "next/cache";
import { beforeEach, describe, expect, it, vi } from "vitest";

import { ApiError } from "@/lib/api/client";
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
    vi.resetAllMocks();
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

  it("checks authentication before rejecting invalid profile fields", async () => {
    const sessionError = new Error("authentication redirect");
    vi.mocked(requireSessionContext).mockRejectedValue(sessionError);

    await expect(
      saveProfileAction(
        IDLE_STATE,
        formData({ name: " ", username: " ", bio: "" }),
      ),
    ).rejects.toBe(sessionError);

    expect(requireSessionContext).toHaveBeenCalledOnce();
    expect(updateProfile).not.toHaveBeenCalled();
  });

  it("checks profile:set before rejecting invalid profile fields", async () => {
    permit("profile:get");

    const result = await saveProfileAction(
      IDLE_STATE,
      formData({ name: " ", username: " ", bio: "" }),
    );

    expect(requireSessionContext).toHaveBeenCalledOnce();
    expect(result).toMatchObject({
      status: "error",
      title: "Permission denied",
    });
    expect(updateProfile).not.toHaveBeenCalled();
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

  it("shows authoritative profile validation details from a 400 response", async () => {
    permit("profile:set");
    vi.mocked(updateProfile).mockRejectedValue(
      new ApiError(
        400,
        "Invalid Format",
        "Unable to update your profile.",
        "username may only contain letters and numbers",
      ),
    );

    const result = await saveProfileAction(
      IDLE_STATE,
      formData({ name: "Alex Morgan", username: "alex!", bio: "" }),
    );

    expect(result).toMatchObject({
      status: "error",
      title: "Invalid Format",
      message: "username may only contain letters and numbers",
    });
  });

  it("does not expose internal profile error details", async () => {
    permit("profile:set");
    vi.mocked(updateProfile).mockRejectedValue(
      new ApiError(
        500,
        "Internal Server Error",
        "Unable to update your profile.",
        "postgres: connection refused at 10.0.0.5",
      ),
    );

    const result = await saveProfileAction(
      IDLE_STATE,
      formData({ name: "Alex Morgan", username: "alex", bio: "" }),
    );

    expect(result).toMatchObject({
      status: "error",
      title: "Internal Server Error",
      message: "Unable to update your profile.",
    });
    expect(JSON.stringify(result)).not.toContain("10.0.0.5");
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

  it("checks profile_security:set before rejecting a password mismatch", async () => {
    permit("profile:get");

    const result = await changePasswordAction(
      IDLE_STATE,
      formData({
        current_password: "current-secret",
        new_password: "new-secret",
        confirm_password: "different-secret",
      }),
    );

    expect(requireSessionContext).toHaveBeenCalledOnce();
    expect(result).toMatchObject({
      status: "error",
      title: "Permission denied",
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

  it("shows authoritative password validation details from a 400 response", async () => {
    permit("profile_security:set");
    vi.mocked(updateProfilePassword).mockRejectedValue(
      new ApiError(
        400,
        "Invalid Format",
        "Unable to change your password.",
        "new_password must be at least 8 characters",
      ),
    );

    const result = await changePasswordAction(
      IDLE_STATE,
      formData({
        current_password: "current-secret",
        new_password: "short",
        confirm_password: "short",
      }),
    );

    expect(result).toMatchObject({
      status: "error",
      title: "Invalid Format",
      message: "new_password must be at least 8 characters",
    });
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
