import { cleanup, render, screen } from "@testing-library/react";
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";

import { getRoleById, listAllRoles } from "@/lib/api/roles";
import { listUsers } from "@/lib/api/users";
import { requirePermission } from "@/lib/session";
import { USER } from "@/test/fixtures";

import UsersPage from "./page";

vi.mock("server-only", () => ({}));
vi.mock("next/navigation", () => ({
  redirect: vi.fn(),
  usePathname: () => "/admin/users",
  useRouter: () => ({ replace: vi.fn() }),
  useSearchParams: () => new URLSearchParams(),
}));
vi.mock("@/lib/api/roles", () => ({
  getRoleById: vi.fn(),
  listAllRoles: vi.fn(),
}));
vi.mock("@/lib/api/users", () => ({ listUsers: vi.fn() }));
vi.mock("@/lib/session", () => ({ requirePermission: vi.fn() }));

const ROLE = {
  id: "role-49",
  name: "Late role",
  description: "Beyond page one",
  is_default: false,
  preferences: {},
  created_at: "2026-08-10T00:00:00Z",
};

describe("UsersPage", () => {
  beforeEach(() => {
    vi.resetAllMocks();
    vi.mocked(listAllRoles).mockResolvedValue([ROLE]);
    vi.mocked(getRoleById).mockResolvedValue(ROLE);
    vi.mocked(listUsers).mockResolvedValue({
      data: [{ ...USER, role_id: ROLE.id }],
      page: { page: 1, limit: 12, total_items: 1 },
    });
  });

  afterEach(cleanup);

  it("resolves displayed user roles from the complete role collection", async () => {
    vi.mocked(requirePermission).mockResolvedValue({
      user: USER,
      permissions: new Set(["user:get", "role:get"]),
    });

    render(await UsersPage({ searchParams: Promise.resolve({}) }));

    expect(screen.getByText("Late role")).toBeVisible();
    expect(listAllRoles).toHaveBeenCalledOnce();
  });

  it("does not collect roles without role:get", async () => {
    vi.mocked(requirePermission).mockResolvedValue({
      user: USER,
      permissions: new Set(["user:get"]),
    });

    render(await UsersPage({ searchParams: Promise.resolve({}) }));

    expect(listAllRoles).not.toHaveBeenCalled();
  });
});
