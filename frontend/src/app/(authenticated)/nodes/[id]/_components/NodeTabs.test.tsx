import { cleanup, render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";

import NodeTabs from "./NodeTabs";
import { normalizeNodeTab } from "../_lib/node-tabs";

const push = vi.fn();

vi.mock("next/navigation", () => ({
  usePathname: () => "/nodes/node-1",
  useRouter: () => ({ push }),
  useSearchParams: () =>
    new URLSearchParams("tab=overview&source=fleet-health"),
}));

describe("NodeTabs", () => {
  beforeEach(() => {
    push.mockReset();
  });

  afterEach(cleanup);

  it("normalizes unknown and repeated tab values to a supported tab", () => {
    expect(normalizeNodeTab("configuration")).toBe("configuration");
    expect(normalizeNodeTab(["logs", "overview"])).toBe("logs");
    expect(normalizeNodeTab("unknown")).toBe("overview");
    expect(normalizeNodeTab(undefined)).toBe("overview");
  });

  it("renders the full node workspace with the selected tab exposed", () => {
    render(<NodeTabs activeTab="overview" nodeId="node-1" />);

    expect(screen.getByRole("tab", { name: "Overview" })).toHaveAttribute(
      "aria-selected",
      "true",
    );
    expect(screen.getByRole("tab", { name: "Configuration" })).toHaveAttribute(
      "aria-selected",
      "false",
    );
    expect(screen.getAllByRole("tab")).toHaveLength(6);
  });

  it("updates only the tab query while preserving workspace context", async () => {
    const user = userEvent.setup();
    render(<NodeTabs activeTab="overview" nodeId="node-1" />);

    await user.click(screen.getByRole("tab", { name: "Configuration" }));

    expect(push).toHaveBeenCalledWith(
      "/nodes/node-1?tab=configuration&source=fleet-health",
    );
  });
});
