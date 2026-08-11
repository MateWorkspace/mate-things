import { describe, expect, it } from "vitest";

import { NAVIGATION } from "@/config/navigation";

import {
  canVisitRoute,
  isProtectedRoute,
  visibleNavigation,
} from "./navigation";

describe("NAVIGATION shape", () => {
  it("orders groups Overview through Administration", () => {
    expect(NAVIGATION.map((group) => group.label)).toEqual([
      "Overview",
      "Fleet",
      "Operations",
      "Observability",
      "Applications",
      "Administration",
    ]);
  });

  it("nests infrared pages under a single app instead of as flat pages", () => {
    const applications = NAVIGATION.find(
      (group) => group.label === "Applications",
    );

    expect(applications?.pages ?? []).toEqual([]);
    expect(applications?.apps).toHaveLength(1);
    expect(applications?.apps?.[0]).toMatchObject({
      key: "infrared",
      label: "Infrared",
    });
    expect(applications?.apps?.[0].pages.map((page) => page.label)).toEqual([
      "Settings",
      "Record",
      "Command",
    ]);
  });

  it("keeps flat groups unaffected by the apps concept", () => {
    const administration = NAVIGATION.find(
      (group) => group.label === "Administration",
    );

    expect(administration?.apps ?? []).toEqual([]);
    expect(administration?.pages?.map((page) => page.label)).toContain(
      "LLM Config",
    );
  });
});

describe("isProtectedRoute", () => {
  it.each([
    ["/ble-direct", true],
    ["/broadcast-sessions/live", true],
    ["/admin/api-keys", true],
    ["/apps/infrared/settings", true],
    ["/login", false],
    ["/", false],
    ["/auth/invalid-session", false],
  ])("classifies %s", (pathname, expected) => {
    expect(isProtectedRoute(pathname)).toBe(expected);
  });

  it("fails closed (treats as protected) for a pathname with no route", () => {
    expect(isProtectedRoute("/nodes-old")).toBe(true);
  });
});

describe("canVisitRoute", () => {
  it("allows a route when any required permission is present", () => {
    expect(
      canVisitRoute("/admin/access-control", new Set(["permission:get"])),
    ).toBe(true);
    expect(canVisitRoute("/admin/access-control", new Set(["user:get"]))).toBe(
      false,
    );
  });

  it("allows pages with no required permissions to any logged-in user", () => {
    expect(canVisitRoute("/dashboard", new Set())).toBe(true);
    expect(canVisitRoute("/ble-direct", new Set())).toBe(true);
  });

  it("fails closed (denies visiting) for a pathname with no route", () => {
    expect(canVisitRoute("/nodes-old", new Set(["node:get", "user:get"]))).toBe(
      false,
    );
  });

  it("does not treat similarly named paths as a match", () => {
    expect(canVisitRoute("/nodes-old", new Set(["node:get"]))).toBe(false);
  });
});

describe("visibleNavigation", () => {
  it("drops an app entirely when the caller can see none of its pages", () => {
    const groups = visibleNavigation(new Set());
    const applications = groups.find((group) => group.label === "Applications");

    expect(applications).toBeUndefined();
  });

  it("keeps an app but filters its pages to only visible ones", () => {
    const groups = visibleNavigation(new Set(["infrared_reference:get"]));
    const applications = groups.find((group) => group.label === "Applications");

    expect(applications?.apps).toHaveLength(1);
    expect(applications?.apps?.[0].pages.map((page) => page.label)).toEqual([
      "Settings",
    ]);
  });

  it("shows every infrared page once the caller has both required permissions", () => {
    const groups = visibleNavigation(
      new Set(["infrared_reference:get", "infrared_record_session:get"]),
    );
    const applications = groups.find((group) => group.label === "Applications");

    expect(applications?.apps?.[0].pages.map((page) => page.label)).toEqual([
      "Settings",
      "Record",
      "Command",
    ]);
  });
});
