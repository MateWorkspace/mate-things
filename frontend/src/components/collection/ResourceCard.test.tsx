import { render, screen } from "@testing-library/react";
import { describe, expect, it } from "vitest";

import ResourceCard from "./ResourceCard";

describe("ResourceCard", () => {
  it("groups a resource heading, summary, details, and contextual actions", () => {
    render(
      <ResourceCard
        actions={<button type="button">View node</button>}
        summary="Connected"
        title="Greenhouse sensor"
      >
        <p>Device ID: 01-23-45</p>
      </ResourceCard>,
    );

    expect(
      screen.getByRole("heading", { level: 2, name: "Greenhouse sensor" }),
    ).toBeInTheDocument();
    expect(screen.getByText("Connected")).toBeInTheDocument();
    expect(screen.getByText("Device ID: 01-23-45")).toBeInTheDocument();
    expect(
      screen.getByRole("button", { name: "View node" }),
    ).toBeInTheDocument();
  });
});
