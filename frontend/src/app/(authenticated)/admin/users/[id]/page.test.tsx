import { cleanup, render, screen } from "@testing-library/react";
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";

import { listAllRoles } from "@/lib/api/roles";
import { getUserById, getUserPermissions } from "@/lib/api/users";
import { requirePermission } from "@/lib/session";
import { USER } from "@/test/fixtures";

import UserDetails from "./page";

vi.mock("server-only", () => ({}));
vi.mock("next/navigation", () => ({
  notFound: vi.fn(),
  useRouter: () => ({ refresh: vi.fn() }),
}));
vi.mock("@/lib/api/roles", () => ({ listAllRoles: vi.fn() }));
vi.mock("@/lib/api/users", () => ({
  getUserById: vi.fn(),
  getUserPermissions: vi.fn(),
}));
vi.mock("@/lib/session", () => ({ requirePermission: vi.fn() }));

const ROLE = {
  id: "role-49",
  name: "Late role",
  description: "Beyond page one",
  is_default: false,
  preferences: {},
  created_at: "2026-08-10T00:00:00Z",
};

describe("UserDetails", () => {
  beforeEach(() => {
    vi.resetAllMocks();
    vi.mocked(getUserById).mockResolvedValue({ ...USER, role_id: ROLE.id });
    vi.mocked(getUserPermissions).mockResolvedValue([]);
    vi.mocked(listAllRoles).mockResolvedValue([ROLE]);
  });

  afterEach(cleanup);

  it("resolves the account role from the complete role collection", async () => {
    vi.mocked(requirePermission).mockResolvedValue({
      user: USER,
      permissions: new Set(["user:get", "role:get"]),
    });

    render(await UserDetails({ params: Promise.resolve({ id: USER.id }) }));

    expect(screen.getByText("Late role")).toBeVisible();
    expect(listAllRoles).toHaveBeenCalledOnce();
  });

  it("does not collect roles without role:get", async () => {
    vi.mocked(requirePermission).mockResolvedValue({
      user: USER,
      permissions: new Set(["user:get"]),
    });

    render(await UserDetails({ params: Promise.resolve({ id: USER.id }) }));

    expect(listAllRoles).not.toHaveBeenCalled();
  });
});
