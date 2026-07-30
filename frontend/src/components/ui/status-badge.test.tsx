import { cleanup, render, screen } from "@testing-library/react";
import { afterEach, describe, expect, it } from "vitest";

import StatusBadge from "./status-badge";

describe("StatusBadge", () => {
  afterEach(cleanup);

  it("pairs critical color with visible text", () => {
    render(<StatusBadge variant="critical">Disconnected</StatusBadge>);

    expect(screen.getByText("Disconnected")).toHaveAttribute(
      "data-variant",
      "critical",
    );
  });

  it.each([
    ["success", "text-success"],
    ["warning", "text-warning"],
    ["critical", "text-critical"],
    ["info", "text-info"],
    ["neutral", "text-foreground"],
  ] as const)(
    "uses an opaque semantic surface with the %s foreground token",
    (variant, foregroundClass) => {
      render(<StatusBadge variant={variant}>Status</StatusBadge>);

      expect(screen.getByText("Status")).toHaveClass(
        "bg-background",
        foregroundClass,
      );
    },
  );
});
