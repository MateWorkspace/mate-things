import { describe, expect, it, vi } from "vitest";

import { collectAllPages } from "./collect-all-pages";

describe("collectAllPages", () => {
  it("deduplicates and stops when a page makes no progress", async () => {
    const fetchPage = vi
      .fn()
      .mockResolvedValueOnce({
        data: [{ id: "a" }],
        page: 1,
        limit: 1,
        total: 3,
      })
      .mockResolvedValueOnce({
        data: [{ id: "a" }],
        page: 2,
        limit: 1,
        total: 3,
      });

    await expect(
      collectAllPages({
        fetchPage,
        keyOf: (item: { id: string }) => item.id,
      }),
    ).resolves.toEqual([{ id: "a" }]);
    expect(fetchPage).toHaveBeenCalledTimes(2);
  });

  it("retains first-seen order while deduplicating across pages", async () => {
    const fetchPage = vi
      .fn()
      .mockResolvedValueOnce({
        data: [{ id: "b" }, { id: "a" }],
        page: 1,
        limit: 2,
        total: 4,
      })
      .mockResolvedValueOnce({
        data: [{ id: "a" }, { id: "c" }],
        page: 2,
        limit: 2,
        total: 4,
      });

    await expect(
      collectAllPages({
        fetchPage,
        keyOf: (item: { id: string }) => item.id,
      }),
    ).resolves.toEqual([{ id: "b" }, { id: "a" }, { id: "c" }]);
  });

  it("uses each response limit when totals imply more pages", async () => {
    const fetchPage = vi
      .fn()
      .mockResolvedValueOnce({
        data: [{ id: "a" }, { id: "b" }],
        page: 1,
        limit: 2,
        total: 4,
      })
      .mockResolvedValueOnce({
        data: [{ id: "c" }],
        page: 2,
        limit: 1,
        total: 4,
      })
      .mockResolvedValueOnce({
        data: [{ id: "d" }],
        page: 3,
        limit: 1,
        total: 4,
      });

    await expect(
      collectAllPages({
        fetchPage,
        keyOf: (item: { id: string }) => item.id,
      }),
    ).resolves.toEqual([{ id: "a" }, { id: "b" }, { id: "c" }, { id: "d" }]);
    expect(fetchPage).toHaveBeenCalledTimes(3);
  });

  it("stops when the backend page does not advance", async () => {
    const fetchPage = vi
      .fn()
      .mockResolvedValueOnce({
        data: [{ id: "a" }],
        page: 1,
        limit: 1,
        total: 3,
      })
      .mockResolvedValueOnce({
        data: [{ id: "b" }],
        page: 1,
        limit: 1,
        total: 3,
      });

    await expect(
      collectAllPages({
        fetchPage,
        keyOf: (item: { id: string }) => item.id,
      }),
    ).resolves.toEqual([{ id: "a" }, { id: "b" }]);
    expect(fetchPage).toHaveBeenCalledTimes(2);
  });

  it("stops safely on an empty intermediate page", async () => {
    const fetchPage = vi
      .fn()
      .mockResolvedValueOnce({
        data: [{ id: "a" }],
        page: 1,
        limit: 1,
        total: 3,
      })
      .mockResolvedValueOnce({ data: [], page: 2, limit: 1, total: 3 });

    await expect(
      collectAllPages({
        fetchPage,
        keyOf: (item: { id: string }) => item.id,
      }),
    ).resolves.toEqual([{ id: "a" }]);
    expect(fetchPage).toHaveBeenCalledTimes(2);
  });

  it("throws when the safety ceiling is exhausted with pages remaining", async () => {
    const fetchPage = vi.fn(async (page: number) => ({
      data: [{ id: String(page) }],
      page,
      limit: 1,
      total: 3,
    }));

    await expect(
      collectAllPages({
        fetchPage,
        keyOf: (item: { id: string }) => item.id,
        maxPages: 2,
      }),
    ).rejects.toThrow(
      "Collection incomplete after reaching the 2-page safety ceiling.",
    );
    expect(fetchPage).toHaveBeenCalledTimes(2);
  });

  it("returns a collection completed exactly at the safety ceiling", async () => {
    const fetchPage = vi.fn(async (page: number) => ({
      data: [{ id: String(page) }],
      page,
      limit: 1,
      total: 2,
    }));

    await expect(
      collectAllPages({
        fetchPage,
        keyOf: (item: { id: string }) => item.id,
        maxPages: 2,
      }),
    ).resolves.toEqual([{ id: "1" }, { id: "2" }]);
    expect(fetchPage).toHaveBeenCalledTimes(2);
  });

  it("terminates without dividing by a zero response limit", async () => {
    const fetchPage = vi.fn().mockResolvedValue({
      data: [{ id: "a" }],
      page: 1,
      limit: 0,
      total: 2,
    });

    await expect(
      collectAllPages({
        fetchPage,
        keyOf: (item: { id: string }) => item.id,
      }),
    ).resolves.toEqual([{ id: "a" }]);
    expect(fetchPage).toHaveBeenCalledTimes(1);
  });
});
