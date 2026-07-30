import { cleanup, render, screen } from "@testing-library/react";
import { redirect } from "next/navigation";
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";

import {
  listNodeClasses,
  type NodeClassResponse,
} from "@/lib/api/node-classes";
import { requirePermission } from "@/lib/session";
import { USER } from "@/test/fixtures";

import NodeClassesPage from "./page";

vi.mock("../node-classes/_lib/actions", () => ({
  createNodeClassAction: vi.fn(),
  deleteNodeClassAction: vi.fn(),
  updateNodeClassAction: vi.fn(),
}));

vi.mock("@/lib/api/node-classes", () => ({
  listNodeClasses: vi.fn(),
}));

vi.mock("@/lib/session", () => ({
  requirePermission: vi.fn(),
}));

vi.mock("next/navigation", () => ({
  redirect: vi.fn(),
}));

const NODE_CLASS: NodeClassResponse = {
  id: "class-1",
  name: "Cold Storage",
  description: "Temperature-controlled sensors",
  preferences: {},
  created_at: "2026-07-30T00:00:00Z",
};

function permit(...permissions: string[]) {
  vi.mocked(requirePermission).mockResolvedValue({
    user: USER,
    permissions: new Set(["node_class:get", ...permissions]),
  });
}

describe("NodeClassesPage", () => {
  beforeEach(() => {
    vi.resetAllMocks();
    vi.mocked(listNodeClasses).mockResolvedValue({
      data: [NODE_CLASS],
      page: { page: 2, limit: 12, total_items: 28 },
    });
  });

  afterEach(cleanup);

  it("passes normalized URL pagination and search to the class wrapper", async () => {
    permit("node_class:add");

    render(
      await NodeClassesPage({
        searchParams: Promise.resolve({
          page: "2",
          limit: "12",
          search: " cold ",
        }),
      }),
    );

    expect(listNodeClasses).toHaveBeenCalledWith({
      page: 2,
      limit: 12,
      search: "cold",
    });
    expect(
      screen.getByRole("button", { name: "Create node class" }),
    ).toBeVisible();
    expect(screen.getByRole("link", { name: "Next page" })).toHaveAttribute(
      "href",
      expect.stringContaining("search=cold"),
    );
  });

  it("omits the create affordance without node_class:add", async () => {
    permit();

    render(
      await NodeClassesPage({
        searchParams: Promise.resolve({}),
      }),
    );

    expect(
      screen.queryByRole("button", { name: "Create node class" }),
    ).not.toBeInTheDocument();
  });

  it("redirects an out-of-range filtered page to the last page", async () => {
    permit();
    vi.mocked(listNodeClasses).mockResolvedValue({
      data: [],
      page: { page: 9, limit: 12, total_items: 25 },
    });

    render(
      await NodeClassesPage({
        searchParams: Promise.resolve({
          page: "9",
          limit: "12",
          search: "cold",
        }),
      }),
    );

    expect(redirect).toHaveBeenCalledWith(
      expect.stringMatching(/^\/node-classes\?(?=.*page=3)(?=.*search=cold)/),
    );
  });
});
