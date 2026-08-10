import {
  cleanup,
  render,
  screen,
  waitFor,
  within,
} from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";

import type { ApiKeyResponse } from "@/lib/api/api-keys";

import GenerateApiKeyDialog from "./GenerateApiKeyDialog";
import RegenerateApiKeyDialog from "./RegenerateApiKeyDialog";

const mocks = vi.hoisted(() => ({
  generateApiKeyAction: vi.fn(),
  regenerateApiKeyAction: vi.fn(),
  refresh: vi.fn(),
}));

vi.mock("next/navigation", () => ({
  useRouter: () => ({ refresh: mocks.refresh }),
}));

vi.mock("../_lib/actions", () => ({
  generateApiKeyAction: mocks.generateApiKeyAction,
  regenerateApiKeyAction: mocks.regenerateApiKeyAction,
}));

vi.mock("@/components/users/UserSearchCombobox", () => ({
  default: ({ name }: { name: string }) => (
    <input aria-label="User" name={name} defaultValue="user-1" />
  ),
}));

const API_KEY: ApiKeyResponse = {
  id: "api-key-1",
  user_id: "user-1",
  user_name: "Alice Example",
  user_username: "alice",
  key_last_four: "1234",
  created_at: "2026-08-10T00:00:00Z",
};

beforeEach(() => {
  vi.resetAllMocks();
});

afterEach(cleanup);

describe("API key secret lifecycle", () => {
  it("shows a generated secret once and discards it when the result closes", async () => {
    const secret = "mate_generated_secret";
    mocks.generateApiKeyAction.mockResolvedValue({
      status: "success",
      title: "API key generated",
      message: "Copy this key now - it won't be shown again.",
      key: secret,
    });
    const user = userEvent.setup();
    render(<GenerateApiKeyDialog />);

    await user.click(screen.getByRole("button", { name: "Generate API key" }));
    await user.click(screen.getByRole("button", { name: "Generate" }));

    expect(await screen.findByText(secret)).toBeVisible();
    expect(screen.getAllByText(secret)).toHaveLength(1);

    await user.click(screen.getByRole("button", { name: "Done" }));
    await waitFor(() =>
      expect(screen.queryByText(secret)).not.toBeInTheDocument(),
    );

    await user.click(screen.getByRole("button", { name: "Generate API key" }));
    expect(screen.queryByText(secret)).not.toBeInTheDocument();
    expect(screen.getByRole("button", { name: "Generate" })).toBeVisible();
  });

  it("shows a regenerated secret once and discards it on a dialog dismissal", async () => {
    const secret = "mate_regenerated_secret";
    mocks.regenerateApiKeyAction.mockResolvedValue({
      status: "success",
      title: "API key regenerated",
      message: "Copy this key now - it won't be shown again.",
      key: secret,
    });
    const user = userEvent.setup();
    render(<RegenerateApiKeyDialog apiKey={API_KEY} />);

    await user.click(screen.getByRole("button", { name: "Regenerate" }));
    await user.click(
      within(screen.getByRole("dialog")).getByRole("button", {
        name: "Regenerate",
      }),
    );

    expect(await screen.findByText(secret)).toBeVisible();
    expect(screen.getAllByText(secret)).toHaveLength(1);

    await user.keyboard("{Escape}");
    await waitFor(() =>
      expect(screen.queryByText(secret)).not.toBeInTheDocument(),
    );

    await user.click(screen.getByRole("button", { name: "Regenerate" }));
    expect(screen.queryByText(secret)).not.toBeInTheDocument();
    expect(
      within(screen.getByRole("dialog")).getByRole("button", {
        name: "Regenerate",
      }),
    ).toBeVisible();
  });
});
