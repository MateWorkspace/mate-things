import { cleanup, render, screen, within } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { afterEach, describe, expect, it } from "vitest";

import CollectionToolbar from "./CollectionToolbar";

describe("CollectionToolbar", () => {
  afterEach(cleanup);

  it("opens the mobile filter drawer with the supplied filters", async () => {
    const user = userEvent.setup();

    render(
      <CollectionToolbar filterTitle="Filter devices">
        <label htmlFor="node-class-filter">Node class</label>
        <select id="node-class-filter" defaultValue="all">
          <option value="all">All classes</option>
        </select>
      </CollectionToolbar>,
    );

    expect(document.querySelectorAll("#node-class-filter")).toHaveLength(1);

    await user.click(screen.getByRole("button", { name: "Open filters" }));

    const drawer = screen.getByRole("dialog", { name: "Filter devices" });
    expect(drawer).toBeInTheDocument();
    expect(document.querySelectorAll("#node-class-filter")).toHaveLength(1);
    expect(within(drawer).getByLabelText("Node class")).toHaveAttribute(
      "id",
      "node-class-filter",
    );

    await user.click(screen.getByRole("button", { name: "Close filters" }));

    expect(
      screen.queryByRole("dialog", { name: "Filter devices" }),
    ).not.toBeInTheDocument();
  });
});
