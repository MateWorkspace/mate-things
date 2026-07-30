import { cleanup, render, screen } from "@testing-library/react";
import { afterEach, describe, expect, it } from "vitest";

import { nodeFixture } from "@/test/fixtures";

import NodeCard from "./NodeCard";

const DISCONNECTED_NODE = nodeFixture();

describe("NodeCard", () => {
  afterEach(cleanup);

  it("shows identity and status without a create action", () => {
    render(<NodeCard node={DISCONNECTED_NODE} permissions={["node:get"]} />);

    expect(screen.getByText("Cold Storage Sensor 07")).toBeVisible();
    expect(screen.getByText("Disconnected")).toBeVisible();
    expect(screen.getByText("AC276E5E030C")).toBeVisible();
    expect(
      screen.queryByRole("button", { name: /create node/i }),
    ).not.toBeInTheDocument();
    expect(
      screen.queryByRole("link", { name: /create node/i }),
    ).not.toBeInTheDocument();
  });

  it("links to details and shows resolved fleet metadata", () => {
    render(
      <NodeCard
        firmwareName="freezer-v2.4.1"
        node={DISCONNECTED_NODE}
        nodeClassName="Cold storage"
        permissions={["node:get"]}
      />,
    );

    expect(screen.getByText("Cold storage")).toBeVisible();
    expect(screen.getByText("freezer-v2.4.1")).toBeVisible();
    expect(screen.getByText("Freezer room sensor")).toBeVisible();
    expect(screen.getByRole("link", { name: "View details" })).toHaveAttribute(
      "href",
      "/nodes/node-1",
    );
  });

  it("shows edit and OTA links only with their exact permissions", () => {
    const { rerender } = render(
      <NodeCard node={DISCONNECTED_NODE} permissions={["node:get"]} />,
    );

    expect(
      screen.queryByRole("link", { name: /edit cold storage sensor 07/i }),
    ).not.toBeInTheDocument();
    expect(
      screen.queryByRole("link", {
        name: /dispatch ota to cold storage sensor 07/i,
      }),
    ).not.toBeInTheDocument();

    rerender(
      <NodeCard
        node={DISCONNECTED_NODE}
        permissions={["node:get", "node:set", "ota:dispatch"]}
      />,
    );

    expect(
      screen.getByRole("link", { name: /edit cold storage sensor 07/i }),
    ).toHaveAttribute("href", "/nodes/node-1?tab=overview");
    expect(
      screen.getByRole("link", {
        name: /dispatch ota to cold storage sensor 07/i,
      }),
    ).toHaveAttribute("href", "/nodes/node-1?tab=firmware");
  });
});
