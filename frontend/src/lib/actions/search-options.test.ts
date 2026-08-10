import { readFile } from "node:fs/promises";
import { resolve } from "node:path";

import { describe, expect, it } from "vitest";

import {
  parseSearchOptionsRequest,
  toSearchOptionsPage,
} from "@/lib/actions/search-options";

const API_RESOURCE_MODULES = [
  "action-logs",
  "actions",
  "api-keys",
  "auth",
  "client",
  "firmwares",
  "node-classes",
  "node-config",
  "node-logs",
  "nodes",
  "ota",
  "payload-schemas",
  "permissions",
  "preferences",
  "profile",
  "role-permissions",
  "roles",
  "telemetry",
  "users",
] as const;

describe("server-only API transport boundary", () => {
  it.each(API_RESOURCE_MODULES)(
    "marks %s as server-only instead of exposing its exports as Server Actions",
    async (moduleName) => {
      const source = await readFile(
        resolve(process.cwd(), `src/lib/api/${moduleName}.ts`),
        "utf8",
      );

      expect(source).toMatch(/^import ["']server-only["'];/);
      expect(source).not.toMatch(/^["']use server["'];/);
    },
  );

  it("keeps shared API response types safe for type-only client imports", async () => {
    const source = await readFile(
      resolve(process.cwd(), "src/lib/api/types.ts"),
      "utf8",
    );

    expect(source).not.toContain('import "server-only"');
    expect(source).not.toMatch(/^["']use server["'];/);
  });
});

describe("parseSearchOptionsRequest", () => {
  it("trims query and preserves a valid bounded request", () => {
    expect(
      parseSearchOptionsRequest({ query: "  cold room  ", page: 2, limit: 6 }),
    ).toEqual({ query: "cold room", page: 2, limit: 6 });
  });

  it.each([
    undefined,
    null,
    [],
    { query: 1, page: 1 },
    { query: "node", page: 0 },
    { query: "node", page: -1 },
    { query: "node", page: 1.5 },
    { query: "node", page: Number.MAX_SAFE_INTEGER + 1 },
    { query: "node", page: 1, limit: 0 },
    { query: "node", page: 1, limit: 1.5 },
    { query: "node", page: 1, limit: 101 },
  ])("rejects malformed or unbounded input %#", (input) => {
    expect(() => parseSearchOptionsRequest(input)).toThrow("Invalid search");
  });
});

describe("toSearchOptionsPage", () => {
  it("returns only mapped options and pagination metadata", () => {
    const response = {
      data: [{ id: "node-1", secret: "not serialized" }],
      page: { page: 2, limit: 6, total_items: 13 },
    };

    expect(
      toSearchOptionsPage(response, (item) => ({
        value: item.id,
        label: "Node one",
      })),
    ).toEqual({
      items: [{ value: "node-1", label: "Node one" }],
      page: 2,
      totalPages: 3,
    });
  });
});
