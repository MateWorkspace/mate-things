import { cleanup, render, screen } from "@testing-library/react";
import { afterEach, describe, expect, it } from "vitest";

import { nodeFixture } from "@/test/fixtures";

import NodeOverview from "./NodeOverview";

describe("NodeOverview", () => {
  afterEach(cleanup);

  it("shows authoritative identity and connection details without inferred health", () => {
    render(<NodeOverview node={nodeFixture()} />);

    expect(
      screen.getByRole("heading", { name: "Node identity" }),
    ).toBeVisible();
    expect(screen.getByText("Disconnected")).toBeVisible();
    expect(screen.getByText("AC276E5E030C")).toBeVisible();
    expect(screen.getByText("ESP32")).toBeVisible();
    expect(
      screen.queryByText(/configuration complete/i),
    ).not.toBeInTheDocument();
    expect(screen.queryByText(/healthy/i)).not.toBeInTheDocument();
  });
});
