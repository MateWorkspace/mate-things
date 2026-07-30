import { render, screen } from "@testing-library/react";
import { Search } from "lucide-react";
import { describe, expect, it } from "vitest";

import IconButton from "./icon-button";

const unnamedIconButton = (
  // @ts-expect-error IconButton requires an accessible name.
  <IconButton>
    <Search aria-hidden="true" />
  </IconButton>
);

describe("IconButton", () => {
  it("exposes its required accessible name and visible focus class", () => {
    render(
      <IconButton aria-label="Search devices">
        <Search aria-hidden="true" />
      </IconButton>,
    );

    expect(screen.getByRole("button", { name: "Search devices" })).toHaveClass(
      "focus-visible:ring-focus",
    );
  });

  it("keeps the type-level unnamed-use regression fixture reachable", () => {
    expect(unnamedIconButton).toBeTruthy();
  });
});
