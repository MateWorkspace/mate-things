import { beforeEach, describe, expect, it, vi } from "vitest";

import { ApiError } from "@/lib/api/client";
import {
  assignRolePermission,
  getRolePermissions,
  revokeRolePermission,
} from "@/lib/api/roles";
import { requireSessionContext } from "@/lib/session";
import { formData } from "@/test/form-data";
import { USER } from "@/test/fixtures";

import {
  EMPTY_ROLE_ASSIGNMENT_STATE,
  updateRoleAssignmentsAction,
} from "./actions";

vi.mock("server-only", () => ({}));

vi.mock("@/lib/api/permissions", () => ({
  createPermission: vi.fn(),
  deletePermission: vi.fn(),
  updatePermission: vi.fn(),
}));

vi.mock("@/lib/api/roles", () => ({
  assignRolePermission: vi.fn(),
  createRole: vi.fn(),
  deleteRole: vi.fn(),
  getRolePermissions: vi.fn(),
  revokeRolePermission: vi.fn(),
  setDefaultRole: vi.fn(),
  updateRole: vi.fn(),
}));

vi.mock("@/lib/session", () => ({
  requireSessionContext: vi.fn(),
}));

function permit(...permissions: string[]) {
  vi.mocked(requireSessionContext).mockResolvedValue({
    user: USER,
    permissions: new Set(permissions),
  });
}

describe("role permission assignments", () => {
  beforeEach(() => {
    vi.resetAllMocks();
  });

  it("returns the exact applied and failed permission ids", async () => {
    permit(
      "role_permission:get",
      "role_permission:add",
      "role_permission:remove",
    );
    vi.mocked(getRolePermissions).mockResolvedValue([
      { id: "permission-keep" },
      { id: "permission-remove" },
    ] as Awaited<ReturnType<typeof getRolePermissions>>);
    vi.mocked(assignRolePermission).mockImplementation(async (_roleId, id) => {
      if (id === "permission-add-failed") {
        throw new ApiError(409, "Conflict", "Permission is unavailable.");
      }
      return { id: `assignment-${id}` };
    });

    const data = formData({ role_id: "role-1" });
    data.append("permission_ids", "permission-keep");
    data.append("permission_ids", "permission-add-ok");
    data.append("permission_ids", "permission-add-failed");

    const result = await updateRoleAssignmentsAction(
      EMPTY_ROLE_ASSIGNMENT_STATE,
      data,
    );

    expect(result.status).toBe("partial");
    expect(result.appliedIds).toEqual([
      "permission-add-ok",
      "permission-remove",
    ]);
    expect(result.failed).toEqual([
      { id: "permission-add-failed", message: "Permission is unavailable." },
    ]);
    expect(revokeRolePermission).toHaveBeenCalledWith(
      "role-1",
      "permission-remove",
    );
  });

  it("reports an authoritative-read failure without attempting mutations", async () => {
    permit("role_permission:get", "role_permission:add");
    vi.mocked(getRolePermissions).mockRejectedValue(
      new ApiError(503, "Unavailable", "Assignments could not be loaded."),
    );
    const data = formData({ role_id: "role-1" });
    data.append("permission_ids", "permission-add");

    const result = await updateRoleAssignmentsAction(
      EMPTY_ROLE_ASSIGNMENT_STATE,
      data,
    );

    expect(result).toMatchObject({
      status: "error",
      title: "Unavailable",
      message: "Assignments could not be loaded.",
      appliedIds: [],
      failed: [],
    });
    expect(assignRolePermission).not.toHaveBeenCalled();
    expect(revokeRolePermission).not.toHaveBeenCalled();
  });
});
