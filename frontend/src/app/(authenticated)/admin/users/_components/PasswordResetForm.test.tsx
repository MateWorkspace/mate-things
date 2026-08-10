import { cleanup, render, screen, within } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { afterEach, beforeEach, expect, it, vi } from "vitest";

import type { UserResponse } from "@/lib/api/users";

import PasswordResetForm from "./PasswordResetForm";

const mocks = vi.hoisted(() => ({ resetUserPasswordAction: vi.fn() }));

vi.mock("../_lib/actions", () => ({
  resetUserPasswordAction: mocks.resetUserPasswordAction,
}));

const userRecord = {
  id: "user-1",
  role_id: "role-1",
  name: "Ada Lovelace",
  bio: "",
  username: "ada",
  preferences: {},
  created_at: "2026-08-10T00:00:00Z",
} as UserResponse;

beforeEach(() => {
  mocks.resetUserPasswordAction.mockReset();
});

afterEach(cleanup);

it("renders password errors, preserves values, and focuses the first invalid field", async () => {
  mocks.resetUserPasswordAction.mockResolvedValue({
    status: "error",
    title: "Check the password",
    message: "Correct the highlighted fields.",
    fieldErrors: {
      password: "Enter a password.",
      password_confirmation: "Passwords do not match.",
    },
  });
  const user = userEvent.setup();
  render(<PasswordResetForm user={userRecord} />);

  await user.click(screen.getByRole("button", { name: "Reset password" }));
  await user.type(screen.getByLabelText("New password"), "draft-secret");
  await user.type(screen.getByLabelText("Confirm password"), "different");
  await user.click(
    within(screen.getByRole("dialog")).getByRole("button", {
      name: "Reset password",
    }),
  );

  expect(await screen.findByText("Enter a password.")).toBeVisible();
  expect(screen.getByText("Passwords do not match.")).toBeVisible();
  expect(screen.getByLabelText("New password")).toHaveValue("draft-secret");
  expect(screen.getByLabelText("Confirm password")).toHaveValue("different");
  expect(screen.getByLabelText("New password")).toHaveFocus();
});
