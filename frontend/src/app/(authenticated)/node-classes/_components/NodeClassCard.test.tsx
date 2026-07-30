import { cleanup, render, screen } from "@testing-library/react";
import { afterEach, describe, expect, it, vi } from "vitest";

import type { NodeClassResponse } from "@/lib/api/node-classes";

import NodeClassCard from "./NodeClassCard";

vi.mock("../_lib/actions", () => ({
  createNodeClassAction: vi.fn(),
  deleteNodeClassAction: vi.fn(),
  updateNodeClassAction: vi.fn(),
}));

const NODE_CLASS: NodeClassResponse = {
  id: "class-1",
  name: "Cold Storage",
  description: "Temperature-controlled sensors",
  preferences: {},
  created_at: "2026-07-30T00:00:00Z",
};

describe("NodeClassCard", () => {
  afterEach(cleanup);

  it("shows class context and links to the detail route", () => {
    render(
      <NodeClassCard nodeClass={NODE_CLASS} permissions={["node_class:get"]} />,
    );

    expect(screen.getByText("Cold Storage")).toBeVisible();
    expect(screen.getByText("Temperature-controlled sensors")).toBeVisible();
    expect(screen.getByText("Created")).toBeVisible();
    expect(screen.getByRole("link", { name: "View details" })).toHaveAttribute(
      "href",
      "/node-classes/class-1",
    );
  });

  it("shows derived relationship counts only when supplied", () => {
    const { rerender } = render(
      <NodeClassCard nodeClass={NODE_CLASS} permissions={["node_class:get"]} />,
    );

    expect(screen.queryByText("Nodes")).not.toBeInTheDocument();
    expect(screen.queryByText("Firmware")).not.toBeInTheDocument();

    rerender(
      <NodeClassCard
        firmwareCount={2}
        nodeClass={NODE_CLASS}
        nodeCount={8}
        permissions={["node_class:get"]}
      />,
    );

    expect(screen.getByText("Nodes")).toBeVisible();
    expect(screen.getByText("8")).toBeVisible();
    expect(screen.getByText("Firmware")).toBeVisible();
    expect(screen.getByText("2")).toBeVisible();
  });

  it("shows the edit modal trigger only with node_class:set", () => {
    const { rerender } = render(
      <NodeClassCard nodeClass={NODE_CLASS} permissions={["node_class:get"]} />,
    );

    expect(
      screen.queryByRole("button", { name: "Edit Cold Storage" }),
    ).not.toBeInTheDocument();

    rerender(
      <NodeClassCard
        nodeClass={NODE_CLASS}
        permissions={["node_class:get", "node_class:set"]}
      />,
    );

    expect(
      screen.getByRole("button", { name: "Edit Cold Storage" }),
    ).toBeVisible();
  });
});
