import { cleanup, render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { afterEach, beforeEach, expect, it, vi } from "vitest";

import SecurityForm from "./SecurityForm";

const mocks = vi.hoisted(() => ({
  changePasswordAction: vi.fn(),
}));

vi.mock("./profile-actions", () => ({
  changePasswordAction: mocks.changePasswordAction,
}));

beforeEach(() => {
  mocks.changePasswordAction.mockReset();
});

afterEach(cleanup);

it("preserves all password values and focuses confirmation after an error", async () => {
  mocks.changePasswordAction.mockResolvedValue({
    status: "error",
    title: "Passwords do not match",
    message: "Confirm your new password and try again.",
    fieldErrors: { confirm_password: "Passwords do not match." },
  });
  const user = userEvent.setup();
  render(<SecurityForm />);

  await user.type(screen.getByLabelText("Current password"), "current-secret");
  await user.type(screen.getByLabelText("New password"), "new-secret");
  await user.type(screen.getByLabelText("Confirm new password"), "different");
  await user.click(screen.getByRole("button", { name: "Change password" }));

  expect(await screen.findByText("Passwords do not match.")).toBeVisible();
  expect(screen.getByLabelText("Current password")).toHaveValue(
    "current-secret",
  );
  expect(screen.getByLabelText("New password")).toHaveValue("new-secret");
  expect(screen.getByLabelText("Confirm new password")).toHaveValue(
    "different",
  );
  expect(screen.getByLabelText("Confirm new password")).toHaveFocus();
});

it("clears sensitive drafts after each success and accepts a second submission", async () => {
  mocks.changePasswordAction.mockResolvedValue({
    status: "success",
    title: "Password changed",
    message: "Your password has been updated.",
  });
  const user = userEvent.setup();
  render(<SecurityForm />);

  await user.type(screen.getByLabelText("Current password"), "current-secret");
  await user.type(screen.getByLabelText("New password"), "new-secret");
  await user.type(screen.getByLabelText("Confirm new password"), "new-secret");
  await user.click(screen.getByRole("button", { name: "Change password" }));

  expect(
    await screen.findByText("Your password has been updated."),
  ).toBeVisible();
  expect(screen.getByLabelText("Current password")).toHaveValue("");
  expect(screen.getByLabelText("New password")).toHaveValue("");
  expect(screen.getByLabelText("Confirm new password")).toHaveValue("");

  await user.type(screen.getByLabelText("Current password"), "next-current");
  await user.type(screen.getByLabelText("New password"), "next-secret");
  await user.type(screen.getByLabelText("Confirm new password"), "next-secret");

  expect(screen.getByLabelText("Current password")).toHaveValue("next-current");
  expect(screen.getByLabelText("New password")).toHaveValue("next-secret");
  expect(screen.getByLabelText("Confirm new password")).toHaveValue(
    "next-secret",
  );

  await user.click(screen.getByRole("button", { name: "Change password" }));

  expect(mocks.changePasswordAction).toHaveBeenCalledTimes(2);
  const secondFormData = mocks.changePasswordAction.mock.calls[1]?.[1] as
    FormData | undefined;
  expect(secondFormData?.get("current_password")).toBe("next-current");
  expect(secondFormData?.get("new_password")).toBe("next-secret");
  expect(secondFormData?.get("confirm_password")).toBe("next-secret");
  expect(screen.getByLabelText("Current password")).toHaveValue("");
  expect(screen.getByLabelText("New password")).toHaveValue("");
  expect(screen.getByLabelText("Confirm new password")).toHaveValue("");
});
