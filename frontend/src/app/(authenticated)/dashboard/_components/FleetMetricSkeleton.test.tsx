import { render, screen, within } from "@testing-library/react";
import { describe, expect, it } from "vitest";

import FleetMetricSkeleton from "./FleetMetricSkeleton";

describe("FleetMetricSkeleton", () => {
  it("places its loading status inside a semantic card", () => {
    render(<FleetMetricSkeleton title="Fleet metrics" />);

    const card = screen.getByRole("article", {
      name: "Fleet metrics metric card",
    });

    expect(card).toHaveClass("bg-surface");
    expect(
      within(card).getByRole("status", { name: "Fleet metrics unavailable" }),
    ).toBeVisible();
  });
});
