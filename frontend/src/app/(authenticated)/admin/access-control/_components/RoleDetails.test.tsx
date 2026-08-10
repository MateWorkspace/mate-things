import { act, cleanup, render, screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { afterEach, beforeEach, expect, it, vi } from "vitest";

import type { PermissionResponse } from "@/lib/api/permissions";
import type { RoleResponse } from "@/lib/api/roles";

import RoleDetails from "./RoleDetails";

const mocks = vi.hoisted(() => ({
  refresh: vi.fn(),
  updateRoleAssignmentsAction: vi.fn(),
}));

vi.mock("next/navigation", () => ({
  useRouter: () => ({ refresh: mocks.refresh }),
}));

vi.mock("../_lib/actions", () => ({
  EMPTY_ROLE_ASSIGNMENT_STATE: {
    status: "idle",
    appliedIds: [],
    failed: [],
  },
  removeRoleAction: vi.fn(),
  saveRoleAction: vi.fn(),
  setDefaultRoleAction: vi.fn(),
  updateRoleAssignmentsAction: mocks.updateRoleAssignmentsAction,
}));

vi.mock("@/components/preferences/PreferencesDialog", () => ({
  default: () => null,
}));

const ROLE: RoleResponse = {
  id: "role-1",
  name: "Operators",
  description: "Fleet operators",
  is_default: false,
  preferences: {},
  created_at: "2026-08-10T00:00:00Z",
};

const permission = (id: string): PermissionResponse => ({
  id,
  name: `node:${id}`,
  description: id,
  preferences: {},
  created_at: "2026-08-10T00:00:00Z",
});

const PERMISSIONS = [
  permission("failed-revoke"),
  permission("applied-add"),
  permission("authoritative-off"),
  permission("authoritative-on"),
];

beforeEach(() => {
  vi.resetAllMocks();
  mocks.updateRoleAssignmentsAction.mockResolvedValue({
    status: "partial",
    title: "Assignments partially updated",
    message: "Failed changes remain selected for retry.",
    appliedIds: ["applied-add"],
    failed: [{ id: "failed-revoke", message: "Conflict" }],
  });
});

afterEach(cleanup);

it("keeps applied and failed desired role assignments before refresh and retries the desired payload", async () => {
  const user = userEvent.setup();
  const { rerender } = render(
    <RoleDetails
      role={ROLE}
      permissions={PERMISSIONS}
      selected={["failed-revoke", "authoritative-off"]}
      grants={[
        "role_permission:get",
        "role_permission:add",
        "role_permission:remove",
      ]}
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

  mocks.updateRoleAssignmentsAction.mockResolvedValueOnce({
    status: "success",
    title: "Assignments updated",
    message: "Role assignments were updated.",
    appliedIds: ["failed-revoke"],
    failed: [],
  });
  await user.click(screen.getByRole("button", { name: "Save assignments" }));
  await waitFor(() =>
    expect(mocks.updateRoleAssignmentsAction).toHaveBeenCalledTimes(2),
  );
  const retryData = mocks.updateRoleAssignmentsAction.mock.calls[1]?.[1];
  expect(retryData).toBeInstanceOf(FormData);
  expect((retryData as FormData).getAll("permission_ids")).toEqual([
    "applied-add",
    "authoritative-off",
  ]);

  expect(
    await screen.findByText("Role assignments were updated."),
  ).toBeVisible();
  rerender(
    <RoleDetails
      role={ROLE}
      permissions={PERMISSIONS}
      selected={["applied-add", "authoritative-off"]}
      grants={[
        "role_permission:get",
        "role_permission:add",
        "role_permission:remove",
      ]}
    />,
  );
  rerender(
    <RoleDetails
      role={ROLE}
      permissions={PERMISSIONS}
      selected={["failed-revoke", "applied-add", "authoritative-off"]}
      grants={[
        "role_permission:get",
        "role_permission:add",
        "role_permission:remove",
      ]}
    />,
  );
  expect(
    screen.getByRole("checkbox", { name: /failed-revoke/i }),
  ).toBeChecked();
});

it("preserves a newer role-assignment edit when an older request resolves", async () => {
  const user = userEvent.setup();
  let resolveAction!: (value: {
    status: "success";
    title: string;
    message: string;
    appliedIds: string[];
    failed: never[];
  }) => void;
  mocks.updateRoleAssignmentsAction.mockImplementationOnce(
    () =>
      new Promise((resolve) => {
        resolveAction = resolve;
      }),
  );
  const { rerender } = render(
    <RoleDetails
      role={ROLE}
      permissions={PERMISSIONS}
      selected={["failed-revoke", "authoritative-off"]}
      grants={[
        "role_permission:get",
        "role_permission:add",
        "role_permission:remove",
      ]}
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
      message: "Role assignments were updated.",
      appliedIds: ["applied-add"],
      failed: [],
    });
  });

  expect(
    await screen.findByText("Role assignments were updated."),
  ).toBeVisible();
  expect(applied).not.toBeChecked();

  rerender(
    <RoleDetails
      role={ROLE}
      permissions={PERMISSIONS}
      selected={["failed-revoke", "applied-add", "authoritative-off"]}
      grants={[
        "role_permission:get",
        "role_permission:add",
        "role_permission:remove",
      ]}
    />,
  );
  expect(
    screen.getByRole("checkbox", { name: /applied-add/i }),
  ).not.toBeChecked();
});

it("yields confirmed role assignments to refreshed props while retaining failures", async () => {
  const user = userEvent.setup();
  const { rerender } = render(
    <RoleDetails
      role={ROLE}
      permissions={PERMISSIONS}
      selected={["failed-revoke", "authoritative-off"]}
      grants={[
        "role_permission:get",
        "role_permission:add",
        "role_permission:remove",
      ]}
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
    <RoleDetails
      role={ROLE}
      permissions={PERMISSIONS}
      selected={["failed-revoke", "applied-add", "authoritative-on"]}
      grants={[
        "role_permission:get",
        "role_permission:add",
        "role_permission:remove",
      ]}
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
    <RoleDetails
      role={ROLE}
      permissions={PERMISSIONS}
      selected={["failed-revoke", "authoritative-on"]}
      grants={[
        "role_permission:get",
        "role_permission:add",
        "role_permission:remove",
      ]}
    />,
  );
  expect(
    screen.getByRole("checkbox", { name: /failed-revoke/i }),
  ).not.toBeChecked();
  expect(
    screen.getByRole("checkbox", { name: /applied-add/i }),
  ).not.toBeChecked();
});

it("releases an already-confirmed no-op role-assignment revision", async () => {
  const user = userEvent.setup();
  mocks.updateRoleAssignmentsAction.mockResolvedValueOnce({
    status: "success",
    title: "Assignments unchanged",
    message: "No role assignments needed changes.",
    appliedIds: [],
    failed: [],
  });
  const { rerender } = render(
    <RoleDetails
      role={ROLE}
      permissions={PERMISSIONS}
      selected={["authoritative-off"]}
      grants={[
        "role_permission:get",
        "role_permission:add",
        "role_permission:remove",
      ]}
    />,
  );

  await user.click(screen.getByRole("checkbox", { name: /applied-add/i }));
  rerender(
    <RoleDetails
      role={ROLE}
      permissions={PERMISSIONS}
      selected={["applied-add", "authoritative-off"]}
      grants={[
        "role_permission:get",
        "role_permission:add",
        "role_permission:remove",
      ]}
    />,
  );
  await user.click(screen.getByRole("button", { name: "Save assignments" }));
  expect(
    await screen.findByText("No role assignments needed changes."),
  ).toBeVisible();

  rerender(
    <RoleDetails
      role={ROLE}
      permissions={PERMISSIONS}
      selected={["authoritative-off"]}
      grants={[
        "role_permission:get",
        "role_permission:add",
        "role_permission:remove",
      ]}
    />,
  );
  expect(
    screen.getByRole("checkbox", { name: /applied-add/i }),
  ).not.toBeChecked();
});

it("releases a role-assignment revision confirmed while its request is pending", async () => {
  const user = userEvent.setup();
  let resolveAction!: (value: {
    status: "success";
    title: string;
    message: string;
    appliedIds: string[];
    failed: never[];
  }) => void;
  mocks.updateRoleAssignmentsAction.mockImplementationOnce(
    () =>
      new Promise((resolve) => {
        resolveAction = resolve;
      }),
  );
  const { rerender } = render(
    <RoleDetails
      role={ROLE}
      permissions={PERMISSIONS}
      selected={["authoritative-off"]}
      grants={[
        "role_permission:get",
        "role_permission:add",
        "role_permission:remove",
      ]}
    />,
  );

  await user.click(screen.getByRole("checkbox", { name: /applied-add/i }));
  await user.click(screen.getByRole("button", { name: "Save assignments" }));
  expect(await screen.findByRole("button", { name: "Saving…" })).toBeDisabled();
  rerender(
    <RoleDetails
      role={ROLE}
      permissions={PERMISSIONS}
      selected={["applied-add", "authoritative-off"]}
      grants={[
        "role_permission:get",
        "role_permission:add",
        "role_permission:remove",
      ]}
    />,
  );

  await act(async () => {
    resolveAction({
      status: "success",
      title: "Assignments updated",
      message: "Role assignments were updated.",
      appliedIds: ["applied-add"],
      failed: [],
    });
  });
  expect(
    await screen.findByText("Role assignments were updated."),
  ).toBeVisible();

  rerender(
    <RoleDetails
      role={ROLE}
      permissions={PERMISSIONS}
      selected={["authoritative-off"]}
      grants={[
        "role_permission:get",
        "role_permission:add",
        "role_permission:remove",
      ]}
    />,
  );
  expect(
    screen.getByRole("checkbox", { name: /applied-add/i }),
  ).not.toBeChecked();
});

it("protects a newer role-assignment revision when an older no-op resolves", async () => {
  const user = userEvent.setup();
  let resolveAction!: (value: {
    status: "success";
    title: string;
    message: string;
    appliedIds: never[];
    failed: never[];
  }) => void;
  mocks.updateRoleAssignmentsAction.mockImplementationOnce(
    () =>
      new Promise((resolve) => {
        resolveAction = resolve;
      }),
  );
  const { rerender } = render(
    <RoleDetails
      role={ROLE}
      permissions={PERMISSIONS}
      selected={["authoritative-off"]}
      grants={[
        "role_permission:get",
        "role_permission:add",
        "role_permission:remove",
      ]}
    />,
  );

  await user.click(screen.getByRole("checkbox", { name: /applied-add/i }));
  await user.click(screen.getByRole("button", { name: "Save assignments" }));
  expect(await screen.findByRole("button", { name: "Saving…" })).toBeDisabled();
  rerender(
    <RoleDetails
      role={ROLE}
      permissions={PERMISSIONS}
      selected={["applied-add", "authoritative-off"]}
      grants={[
        "role_permission:get",
        "role_permission:add",
        "role_permission:remove",
      ]}
    />,
  );
  await user.click(screen.getByRole("checkbox", { name: /applied-add/i }));

  await act(async () => {
    resolveAction({
      status: "success",
      title: "Assignments unchanged",
      message: "No role assignments needed changes.",
      appliedIds: [],
      failed: [],
    });
  });
  expect(
    await screen.findByText("No role assignments needed changes."),
  ).toBeVisible();
  expect(
    screen.getByRole("checkbox", { name: /applied-add/i }),
  ).not.toBeChecked();

  rerender(
    <RoleDetails
      role={ROLE}
      permissions={PERMISSIONS}
      selected={["authoritative-off"]}
      grants={[
        "role_permission:get",
        "role_permission:add",
        "role_permission:remove",
      ]}
    />,
  );
  rerender(
    <RoleDetails
      role={ROLE}
      permissions={PERMISSIONS}
      selected={["applied-add", "authoritative-off"]}
      grants={[
        "role_permission:get",
        "role_permission:add",
        "role_permission:remove",
      ]}
    />,
  );
  expect(
    screen.getByRole("checkbox", { name: /applied-add/i }),
  ).not.toBeChecked();
});
