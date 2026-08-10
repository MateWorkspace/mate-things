import { cleanup, render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { afterEach, beforeEach, expect, it, vi } from "vitest";

import GenerateApiKeyDialog from "./GenerateApiKeyDialog";

const mocks = vi.hoisted(() => ({
  generateApiKeyAction: vi.fn(),
  refresh: vi.fn(),
}));

vi.mock("next/navigation", () => ({
  useRouter: () => ({ refresh: mocks.refresh }),
}));

vi.mock("../_lib/actions", () => ({
  generateApiKeyAction: mocks.generateApiKeyAction,
}));

vi.mock("@/components/users/UserSearchCombobox", () => ({
  default: ({ name }: { name: string }) => (
    <>
      <input type="hidden" name={name} />
      <button type="button">Choose user</button>
    </>
  ),
}));

beforeEach(() => {
  mocks.generateApiKeyAction.mockReset();
  mocks.refresh.mockReset();
});

afterEach(cleanup);

it("focuses the expiration field when the action rejects it", async () => {
  mocks.generateApiKeyAction.mockResolvedValue({
    status: "error",
    title: "Check the form",
    message: "Complete the required fields.",
    fieldErrors: { expires_at: "Enter a valid expiration date." },
  });
  const user = userEvent.setup();
  render(<GenerateApiKeyDialog />);

  await user.click(screen.getByRole("button", { name: "Generate API key" }));
  await user.click(screen.getByRole("button", { name: "Generate" }));

  expect(
    await screen.findByText("Enter a valid expiration date."),
  ).toBeVisible();
  expect(screen.getByLabelText("Expires at (optional)")).toHaveFocus();
});
