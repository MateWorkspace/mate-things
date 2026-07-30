import { render, screen } from "@testing-library/react";
import { describe, expect, it } from "vitest";

import StatusBadge from "./status-badge";

describe("StatusBadge", () => {
  it("pairs critical color with visible text", () => {
    render(<StatusBadge variant="critical">Disconnected</StatusBadge>);

    expect(screen.getByText("Disconnected")).toHaveAttribute(
      "data-variant",
      "critical",
    );
  });
});
