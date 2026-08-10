import { act, cleanup, render, screen, waitFor } from "@testing-library/react";
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

it("keeps applied and failed desired node-class actions before refresh and retries the desired payload", async () => {
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
  expect(
    screen.getByRole("checkbox", { name: /failed-revoke/i }),
  ).not.toBeChecked();
  expect(screen.getByRole("checkbox", { name: /applied-add/i })).toBeChecked();

  mocks.updateNodeClassActionsAction.mockResolvedValueOnce({
    status: "success",
    title: "Assignments updated",
    message: "Node-class actions were updated.",
    appliedIds: ["failed-revoke"],
    failed: [],
  });
  await user.click(screen.getByRole("button", { name: "Save assignments" }));
  await waitFor(() =>
    expect(mocks.updateNodeClassActionsAction).toHaveBeenCalledTimes(2),
  );
  const retryData = mocks.updateNodeClassActionsAction.mock.calls[1]?.[1];
  expect(retryData).toBeInstanceOf(FormData);
  expect((retryData as FormData).getAll("action_ids")).toEqual([
    "applied-add",
    "authoritative-off",
  ]);

  expect(
    await screen.findByText("Node-class actions were updated."),
  ).toBeVisible();
  rerender(
    <NodeClassActionChecklist
      nodeClassId="class-1"
      actions={ACTIONS}
      selected={new Set(["applied-add", "authoritative-off"])}
      editable
    />,
  );
  rerender(
    <NodeClassActionChecklist
      nodeClassId="class-1"
      actions={ACTIONS}
      selected={new Set(["failed-revoke", "applied-add", "authoritative-off"])}
      editable
    />,
  );
  expect(
    screen.getByRole("checkbox", { name: /failed-revoke/i }),
  ).toBeChecked();
});

it("preserves a newer node-class action edit when an older request resolves", async () => {
  const user = userEvent.setup();
  let resolveAction!: (value: {
    status: "success";
    title: string;
    message: string;
    appliedIds: string[];
    failed: never[];
  }) => void;
  mocks.updateNodeClassActionsAction.mockImplementationOnce(
    () =>
      new Promise((resolve) => {
        resolveAction = resolve;
      }),
  );
  const { rerender } = render(
    <NodeClassActionChecklist
      nodeClassId="class-1"
      actions={ACTIONS}
      selected={new Set(["failed-revoke", "authoritative-off"])}
      editable
    />,
  );

  const applied = screen.getByRole("checkbox", { name: /applied-add/i });
  await user.click(applied);
  await user.click(screen.getByRole("button", { name: "Save assignments" }));
  expect(await screen.findByRole("button", { name: "Saving…" })).toBeDisabled();
  await user.click(applied);

  await act(async () => {
    resolveAction({
      status: "success",
      title: "Assignments updated",
      message: "Node-class actions were updated.",
      appliedIds: ["applied-add"],
      failed: [],
    });
  });

  expect(
    await screen.findByText("Node-class actions were updated."),
  ).toBeVisible();
  expect(applied).not.toBeChecked();

  rerender(
    <NodeClassActionChecklist
      nodeClassId="class-1"
      actions={ACTIONS}
      selected={new Set(["failed-revoke", "applied-add", "authoritative-off"])}
      editable
    />,
  );
  expect(
    screen.getByRole("checkbox", { name: /applied-add/i }),
  ).not.toBeChecked();
});

it("yields confirmed node-class actions to refreshed props while retaining failures", async () => {
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

  expect(
    screen.getByRole("checkbox", { name: /failed-revoke/i }),
  ).not.toBeChecked();
  expect(screen.getByRole("checkbox", { name: /applied-add/i })).toBeChecked();

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

  rerender(
    <NodeClassActionChecklist
      nodeClassId="class-1"
      actions={ACTIONS}
      selected={new Set(["failed-revoke", "authoritative-on"])}
      editable
    />,
  );
  expect(
    screen.getByRole("checkbox", { name: /failed-revoke/i }),
  ).not.toBeChecked();
  expect(
    screen.getByRole("checkbox", { name: /applied-add/i }),
  ).not.toBeChecked();
});
