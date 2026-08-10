import { cleanup, render, screen, waitFor } from "@testing-library/react";
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

it("reconciles refreshed role assignments while retaining failed desired changes", async () => {
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
});
