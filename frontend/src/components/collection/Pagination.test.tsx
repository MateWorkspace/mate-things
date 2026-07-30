import { cleanup, render, screen } from "@testing-library/react";
import { afterEach, describe, expect, it } from "vitest";

import Pagination from "./Pagination";

describe("Pagination", () => {
  afterEach(cleanup);

  it("preserves filters while moving to the next page", () => {
    render(
      <Pagination
        page={{ page: 2, limit: 12, total_items: 40 }}
        searchParams={{ search: "sensor", node_class_id: "cold" }}
      />,
    );

    expect(screen.getByRole("link", { name: "Next page" })).toHaveAttribute(
      "href",
      expect.stringContaining("page=3"),
    );
    expect(screen.getByRole("link", { name: "Next page" })).toHaveAttribute(
      "href",
      expect.stringContaining("node_class_id=cold"),
    );
  });

  it("marks impossible navigation as disabled while reporting the page position", () => {
    render(
      <Pagination
        page={{ page: 1, limit: 12, total_items: 8 }}
        searchParams={{ search: "sensor" }}
      />,
    );

    expect(screen.getByText("Previous page")).toHaveAttribute(
      "aria-disabled",
      "true",
    );
    expect(screen.getByText("Next page")).toHaveAttribute(
      "aria-disabled",
      "true",
    );
    expect(screen.getByText("Page 1 of 1")).toBeInTheDocument();
  });
});
