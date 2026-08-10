import { cleanup, render, screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { afterEach, beforeEach, expect, it, vi } from "vitest";

import type { ActionResponse } from "@/lib/api/actions";

import NodeClassActionChecklist from "./NodeClassActionChecklist";

const mocks = vi.hoisted(() => ({
  refresh: vi.fn(),
  updateNodeClassActionsAction: vi.fn(),
}));

vi.mock("next/navigation", () => ({
  useRouter: () => ({ refresh: mocks.refresh }),
}));

vi.mock("../_lib/actions", () => ({
  EMPTY_NODE_CLASS_ASSIGNMENT_STATE: {
    status: "idle",
    appliedIds: [],
    failed: [],
  },
  updateNodeClassActionsAction: mocks.updateNodeClassActionsAction,
}));

const action = (id: string): ActionResponse => ({
  id,
  name: id,
  description: id,
  payload_schema_name: "command",
  payload_schema_version: 1,
  preferences: {},
  created_at: "2026-08-10T00:00:00Z",
});

const ACTIONS = [
  action("failed-revoke"),
  action("applied-add"),
  action("authoritative-off"),
  action("authoritative-on"),
];

beforeEach(() => {
  vi.resetAllMocks();
  mocks.updateNodeClassActionsAction.mockResolvedValue({
    status: "partial",
    title: "Assignments partially updated",
    message: "Failed changes remain selected for retry.",
    appliedIds: ["applied-add"],
    failed: [{ id: "failed-revoke", message: "Conflict" }],
  });
});

afterEach(cleanup);

it("reconciles refreshed node-class actions while retaining failed desired changes", async () => {
  const user = userEvent.setup();
  const { rerender } = render(
    <NodeClassActionChecklist
      nodeClassId="class-1"
      actions={ACTIONS}
      selected={new Set(["failed-revoke", "authoritative-off"])}
      editable
    />,
  );

  await user.click(screen.getByRole("checkbox", { name: /failed-revoke/i }));
  await user.click(screen.getByRole("checkbox", { name: /applied-add/i }));
  await user.click(screen.getByRole("button", { name: "Save assignments" }));

  expect(
    await screen.findByText("Failed changes remain selected for retry."),
  ).toBeVisible();
  await waitFor(() => expect(mocks.refresh).toHaveBeenCalledOnce());

  rerender(
    <NodeClassActionChecklist
      nodeClassId="class-1"
      actions={ACTIONS}
      selected={new Set(["failed-revoke", "applied-add", "authoritative-on"])}
      editable
    />,
  );

  expect(
    screen.getByRole("checkbox", { name: /failed-revoke/i }),
  ).not.toBeChecked();
  expect(screen.getByRole("checkbox", { name: /applied-add/i })).toBeChecked();
  expect(
    screen.getByRole("checkbox", { name: /authoritative-off/i }),
  ).not.toBeChecked();
  expect(
    screen.getByRole("checkbox", { name: /authoritative-on/i }),
  ).toBeChecked();
});
