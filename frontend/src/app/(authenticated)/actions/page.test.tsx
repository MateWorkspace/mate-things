import { cleanup, render, screen } from "@testing-library/react";
import { redirect } from "next/navigation";
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";

import { listActions } from "@/lib/api/actions";
import { getNodeClassById } from "@/lib/api/node-classes";
import { listAllPayloadSchemas } from "@/lib/api/payload-schemas";
import { requirePermission } from "@/lib/session";
import { USER } from "@/test/fixtures";

import ActionsPage from "./page";

vi.mock("server-only", () => ({}));
vi.mock("./_lib/actions", () => ({
  createAction: vi.fn(),
  deleteAction: vi.fn(),
  updateAction: vi.fn(),
}));
vi.mock("@/lib/api/actions", () => ({ listActions: vi.fn() }));
vi.mock("@/lib/api/node-classes", () => ({ getNodeClassById: vi.fn() }));
vi.mock("@/lib/api/payload-schemas", () => ({
  listAllPayloadSchemas: vi.fn(),
}));
vi.mock("@/lib/session", () => ({ requirePermission: vi.fn() }));
vi.mock("next/navigation", () => ({
  redirect: vi.fn(),
  usePathname: () => "/actions",
  useRouter: () => ({ replace: vi.fn() }),
  useSearchParams: () => new URLSearchParams(),
}));

describe("ActionsPage", () => {
  beforeEach(() => {
    vi.resetAllMocks();
    vi.mocked(requirePermission).mockResolvedValue({
      user: USER,
      permissions: new Set(["action:get", "node_class:get"]),
    });
    vi.mocked(listActions).mockResolvedValue({
      data: [],
      page: { page: 8, limit: 12, total_items: 25 },
    });
    vi.mocked(getNodeClassById).mockResolvedValue({
      id: "class-1",
      name: "Cold Storage",
      description: "Sensors",
      preferences: {},
      created_at: "2026-08-10T00:00:00Z",
    });
    vi.mocked(listAllPayloadSchemas).mockResolvedValue([]);
  });

  afterEach(cleanup);

  it("redirects an out-of-range page while retaining active filters", async () => {
    render(
      await ActionsPage({
        searchParams: Promise.resolve({
          page: "8",
          limit: "12",
          search: "reboot",
          node_class_id: "class-1",
        }),
      }),
    );

    expect(redirect).toHaveBeenCalledWith(
      expect.stringMatching(
        /^\/actions\?(?=.*page=3)(?=.*limit=12)(?=.*search=reboot)(?=.*node_class_id=class-1)/,
      ),
    );
  });

  it("does not fetch or offer a node-class selector without node_class:get", async () => {
    vi.mocked(requirePermission).mockResolvedValue({
      user: USER,
      permissions: new Set(["action:get"]),
    });

    render(
      await ActionsPage({
        searchParams: Promise.resolve({ node_class_id: "class-1" }),
      }),
    );

    expect(getNodeClassById).not.toHaveBeenCalled();
    expect(screen.getByDisplayValue("class-1")).toHaveAttribute("readonly");
    expect(
      screen.queryByRole("button", { name: /node class/i }),
    ).not.toBeInTheDocument();
  });
});
