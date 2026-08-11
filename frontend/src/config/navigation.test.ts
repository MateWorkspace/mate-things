import { describe, expect, it } from "vitest";

import { NAVIGATION_GROUPS, visibleNavigation } from "./navigation";

describe("navigation groups", () => {
  it("orders groups overview through administration, with Application before Administration", () => {
    expect(NAVIGATION_GROUPS.map((group) => group.label)).toEqual([
      "Overview",
      "Fleet",
      "Operations",
      "Observability",
      "Application",
      "Administration",
    ]);
  });

  it("nests infrared pages under a single Infrared app instead of as flat items", () => {
    const application = NAVIGATION_GROUPS.find(
      (group) => group.label === "Application",
    );

    expect(application?.items).toEqual([]);
    expect(application?.apps).toHaveLength(1);
    expect(application?.apps[0]).toMatchObject({
      key: "infrared",
      label: "Infrared",
    });
    expect(application?.apps[0].items.map((item) => item.label)).toEqual([
      "Settings",
      "Record",
      "Command",
    ]);
  });

  it("keeps flat groups unaffected by the apps concept", () => {
    const administration = NAVIGATION_GROUPS.find(
      (group) => group.label === "Administration",
    );

    expect(administration?.apps).toEqual([]);
    expect(administration?.items.map((item) => item.label)).toContain(
      "LLM Config",
    );
  });
});

describe("visibleNavigation", () => {
  it("drops an app entirely when the caller can see none of its pages", () => {
    const groups = visibleNavigation(new Set());
    const application = groups.find((group) => group.label === "Application");

    expect(application).toBeUndefined();
  });

  it("keeps an app but filters its items to only visible pages", () => {
    const groups = visibleNavigation(new Set(["infrared_reference:get"]));
    const application = groups.find((group) => group.label === "Application");

    expect(application?.apps).toHaveLength(1);
    expect(application?.apps[0].items.map((item) => item.label)).toEqual([
      "Settings",
    ]);
  });

  it("shows every infrared page once the caller has both required permissions", () => {
    const groups = visibleNavigation(
      new Set(["infrared_reference:get", "infrared_record_session:get"]),
    );
    const application = groups.find((group) => group.label === "Application");

    expect(application?.apps[0].items.map((item) => item.label)).toEqual([
      "Settings",
      "Record",
      "Command",
    ]);
  });
});
