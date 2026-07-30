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
        <label>
          Node class
          <select defaultValue="all">
            <option value="all">All classes</option>
          </select>
        </label>
      </CollectionToolbar>,
    );

    await user.click(screen.getByRole("button", { name: "Open filters" }));

    const drawer = screen.getByRole("dialog", { name: "Filter devices" });
    expect(drawer).toBeInTheDocument();
    expect(within(drawer).getByLabelText("Node class")).toBeInTheDocument();

    await user.click(screen.getByRole("button", { name: "Close filters" }));

    expect(
      screen.queryByRole("dialog", { name: "Filter devices" }),
    ).not.toBeInTheDocument();
  });
});
