import { beforeEach, describe, expect, it, vi } from "vitest";

import { apiFetch } from "@/lib/api/client";

import { listAllNodeClasses, type NodeClassResponse } from "./node-classes";

vi.mock("@/lib/api/client", () => ({
  apiFetch: vi.fn(),
  buildQuery: vi.fn(
    ({ page, limit }: { page?: number; limit?: number }) =>
      `?page=${page}&limit=${limit}`,
  ),
}));

function nodeClass(index: number): NodeClassResponse {
  return {
    id: `class-${index}`,
    name: `Node class ${index}`,
    description: "",
    preferences: {},
    created_at: "2026-07-30T00:00:00Z",
  };
}

describe("listAllNodeClasses", () => {
  beforeEach(() => {
    vi.resetAllMocks();
  });

  it("loads every page so options after the first 48 remain reachable", async () => {
    vi.mocked(apiFetch)
      .mockResolvedValueOnce({
        data: Array.from({ length: 48 }, (_, index) => nodeClass(index + 1)),
        page: { page: 1, limit: 48, total_items: 49 },
      })
      .mockResolvedValueOnce({
        data: [nodeClass(49)],
        page: { page: 2, limit: 48, total_items: 49 },
      });

    const result = await listAllNodeClasses();

    expect(result).toHaveLength(49);
    expect(result.at(-1)).toMatchObject({
      id: "class-49",
      name: "Node class 49",
    });
    expect(apiFetch).toHaveBeenNthCalledWith(
      1,
      "/node-classes?page=1&limit=48",
    );
    expect(apiFetch).toHaveBeenNthCalledWith(
      2,
      "/node-classes?page=2&limit=48",
    );
  });

  it("deduplicates and stops if a server repeats a page", async () => {
    vi.mocked(apiFetch)
      .mockResolvedValueOnce({
        data: [nodeClass(1)],
        page: { page: 1, limit: 1, total_items: 1000 },
      })
      .mockResolvedValueOnce({
        data: [nodeClass(1)],
        page: { page: 2, limit: 1, total_items: 1000 },
      });

    await expect(listAllNodeClasses()).resolves.toEqual([nodeClass(1)]);
    expect(apiFetch).toHaveBeenCalledTimes(2);
  });
});
