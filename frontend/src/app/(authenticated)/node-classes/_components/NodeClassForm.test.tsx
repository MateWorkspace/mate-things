import { cleanup, render, screen, within } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { afterEach, describe, expect, it, vi } from "vitest";

import type { NodeClassResponse } from "@/lib/api/node-classes";

import NodeClassForm from "./NodeClassForm";

vi.mock("../_lib/actions", () => ({
  createNodeClassAction: vi.fn(),
  deleteNodeClassAction: vi.fn(),
  updateNodeClassAction: vi.fn(),
}));

const NODE_CLASS: NodeClassResponse = {
  id: "class-1",
  name: "Cold Storage",
  description: "Temperature-controlled sensors",
  preferences: {},
  created_at: "2026-07-30T00:00:00Z",
};

describe("NodeClassForm", () => {
  afterEach(cleanup);

  it("opens a modal with the create fields", async () => {
    const user = userEvent.setup();
    render(<NodeClassForm />);

    await user.click(screen.getByRole("button", { name: "Create node class" }));

    const dialog = screen.getByRole("dialog", {
      name: "Create node class",
    });
    expect(within(dialog).getByLabelText("Name")).toBeRequired();
    expect(within(dialog).getByLabelText("Description")).toBeVisible();
    expect(
      within(dialog).getByRole("button", { name: "Create class" }),
    ).toBeVisible();
  });

  it("opens a prefilled edit modal for an existing class", async () => {
    const user = userEvent.setup();
    render(<NodeClassForm nodeClass={NODE_CLASS} />);

    await user.click(screen.getByRole("button", { name: "Edit Cold Storage" }));

    const dialog = screen.getByRole("dialog", {
      name: "Edit Cold Storage",
    });
    expect(within(dialog).getByLabelText("Name")).toHaveValue("Cold Storage");
    expect(within(dialog).getByLabelText("Description")).toHaveValue(
      "Temperature-controlled sensors",
    );
  });

  it("uses a separate named confirmation and honest dependency warning", async () => {
    const user = userEvent.setup();
    render(<NodeClassForm canDelete nodeClass={NODE_CLASS} />);

    await user.click(
      screen.getByRole("button", { name: "Delete Cold Storage" }),
    );

    const dialog = screen.getByRole("dialog", {
      name: "Delete Cold Storage",
    });
    expect(
      within(dialog).getByText(/dependent nodes, firmware, or actions/i),
    ).toBeVisible();
    expect(
      within(dialog).getByText(/backend may reject this deletion/i),
    ).toBeVisible();
    expect(within(dialog).getByLabelText("Confirm class name")).toHaveAttribute(
      "placeholder",
      "Cold Storage",
    );
  });
});
