import { render, screen } from "@testing-library/react";
import { describe, expect, it } from "vitest";

import Button from "./button";

describe("Button", () => {
  it("preserves native accessibility and applies its variant", () => {
    render(<Button variant="secondary">Cancel</Button>);
    expect(screen.getByRole("button", { name: "Cancel" })).toHaveClass(
      "text-primary",
    );
  });
});
