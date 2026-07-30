import { cleanup, render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { afterEach, describe, expect, it, vi } from "vitest";

import DeleteNodeDialog from "./DeleteNodeDialog";

vi.mock("../_lib/actions", () => ({
  deleteNodeAction: vi.fn(),
}));

describe("DeleteNodeDialog", () => {
  afterEach(cleanup);

  it("requires typed confirmation and names the exact device ID", async () => {
    const user = userEvent.setup();
    render(
      <DeleteNodeDialog
        deviceId="AC276E5E030C"
        nodeId="node-1"
        nodeName="Cold Storage Sensor 07"
      />,
    );

    await user.click(screen.getByRole("button", { name: "Delete node" }));

    expect(screen.getByText("AC276E5E030C")).toBeVisible();
    expect(screen.getByLabelText("Confirm device ID")).toBeRequired();
    expect(
      screen.getByRole("button", { name: "Permanently delete node" }),
    ).toBeVisible();
  });
});
