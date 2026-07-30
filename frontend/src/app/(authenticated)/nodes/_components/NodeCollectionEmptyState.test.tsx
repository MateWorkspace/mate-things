import { cleanup, render, screen } from "@testing-library/react";
import { afterEach, describe, expect, it } from "vitest";

import NodeCollectionEmptyState from "./NodeCollectionEmptyState";

describe("NodeCollectionEmptyState", () => {
  afterEach(cleanup);

  it("explains MQTT self-registration for a genuinely empty fleet", () => {
    render(
      <NodeCollectionEmptyState connection="all" limit={24} totalItems={0} />,
    );

    expect(screen.getByText("No nodes registered")).toBeVisible();
    expect(screen.getByText(/self-register through mqtt/i)).toBeVisible();
    expect(
      screen.queryByRole("link", { name: "Clear filters" }),
    ).not.toBeInTheDocument();
    expect(
      screen.queryByText(/try changing.*filters/i),
    ).not.toBeInTheDocument();
  });

  it("offers to clear URL filters when no nodes match", () => {
    render(
      <NodeCollectionEmptyState
        connection="all"
        limit={24}
        search="freezer"
        totalItems={0}
      />,
    );

    expect(screen.getByText("No matching nodes")).toBeVisible();
    expect(screen.getByRole("link", { name: "Clear filters" })).toHaveAttribute(
      "href",
      "/nodes?limit=24",
    );
  });
});
