import { cleanup, render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { afterEach, describe, expect, it, vi } from "vitest";

import { nodeFixture } from "@/test/fixtures";

import NodeEditForm from "./NodeEditForm";

vi.mock("../_lib/actions", () => ({
  saveNodeAction: vi.fn(),
}));

vi.mock("next/navigation", () => ({
  useRouter: vi.fn(() => ({ refresh: vi.fn() })),
}));

describe("NodeEditForm", () => {
  afterEach(cleanup);

  it("opens a focused form with immutable device identity", async () => {
    const user = userEvent.setup();
    render(<NodeEditForm node={nodeFixture()} />);

    await user.click(screen.getByRole("button", { name: "Edit node" }));

    expect(screen.getByRole("dialog")).toBeVisible();
    expect(screen.getByLabelText("Name")).toHaveValue("Cold Storage Sensor 07");
    expect(screen.getByText("AC276E5E030C")).toBeVisible();
    expect(
      screen.queryByRole("textbox", { name: "Device ID" }),
    ).not.toBeInTheDocument();
    expect(screen.getByLabelText("Description")).toHaveValue(
      "Freezer room sensor",
    );
  });
});
