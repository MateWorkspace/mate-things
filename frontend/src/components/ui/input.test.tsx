import { render, screen } from "@testing-library/react";
import { describe, expect, it } from "vitest";

import Input from "./input";

describe("Input", () => {
  it("uses the semantic control-boundary token", () => {
    render(<Input aria-label="Device name" />);

    expect(screen.getByRole("textbox", { name: "Device name" })).toHaveClass(
      "border-control-border",
    );
  });
});
