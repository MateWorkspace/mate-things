import { cleanup, render, screen } from "@testing-library/react";
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";

import { listAllPermissions } from "@/lib/api/permissions";
import { getRolePermissions, listAllRoles } from "@/lib/api/roles";
import { requireAnyPermission } from "@/lib/session";
import { USER } from "@/test/fixtures";

import AccessControlPage from "./page";

vi.mock("server-only", () => ({}));
vi.mock("next/navigation", () => ({
  useRouter: () => ({ refresh: vi.fn() }),
}));
vi.mock("@/lib/api/permissions", () => ({ listAllPermissions: vi.fn() }));
vi.mock("@/lib/api/roles", () => ({
  getRolePermissions: vi.fn(),
  listAllRoles: vi.fn(),
}));
vi.mock("@/lib/session", () => ({ requireAnyPermission: vi.fn() }));

const ROLE = {
  id: "role-49",
  name: "Beyond first page",
  description: "Late role",
  is_default: false,
  preferences: {},
  created_at: "2026-08-10T00:00:00Z",
};

const PERMISSION = {
  id: "permission-49",
  name: "late:get",
  description: "Late permission",
  preferences: {},
  created_at: "2026-08-10T00:00:00Z",
};

function permit(...permissions: string[]) {
  vi.mocked(requireAnyPermission).mockResolvedValue({
    user: USER,
    permissions: new Set(permissions),
  });
}

describe("AccessControlPage", () => {
  beforeEach(() => {
    vi.resetAllMocks();
    vi.mocked(listAllRoles).mockResolvedValue([ROLE]);
    vi.mocked(listAllPermissions).mockResolvedValue([PERMISSION]);
    vi.mocked(getRolePermissions).mockResolvedValue([]);
  });

  afterEach(cleanup);

  it("renders complete role options and does not read forbidden permissions", async () => {
    permit("role:get");

    render(
      await AccessControlPage({
        searchParams: Promise.resolve({ tab: "roles" }),
      }),
    );

    expect(screen.getByText("Beyond first page")).toBeVisible();
    expect(listAllRoles).toHaveBeenCalledOnce();
    expect(listAllPermissions).not.toHaveBeenCalled();
  });

  it("renders complete permission options and does not read forbidden roles", async () => {
    permit("permission:get");

    render(
      await AccessControlPage({
        searchParams: Promise.resolve({ tab: "permissions" }),
      }),
    );

    expect(
      screen.getByRole("heading", { name: "late:get", level: 2 }),
    ).toBeVisible();
    expect(listAllPermissions).toHaveBeenCalledOnce();
    expect(listAllRoles).not.toHaveBeenCalled();
  });
});
