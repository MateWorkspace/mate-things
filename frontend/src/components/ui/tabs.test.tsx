import { cleanup, render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { useState } from "react";
import { afterEach, describe, expect, it, vi } from "vitest";

import Tabs from "./tabs";

const tabs = [
  { id: "overview", label: "Overview", panelId: "overview-panel" },
  {
    id: "settings",
    label: "Settings",
    panelId: "settings-panel",
    disabled: true,
  },
  { id: "logs", label: "Logs", panelId: "logs-panel" },
  { id: "history", label: "History", panelId: "history-panel" },
];

function TabsHarness({ onChange }: { onChange: (id: string) => void }) {
  const [activeTab, setActiveTab] = useState("overview");

  return (
    <Tabs
      tabs={tabs}
      activeTab={activeTab}
      ariaLabel="Node workspace"
      onChange={(id) => {
        setActiveTab(id);
        onChange(id);
      }}
    />
  );
}

describe("Tabs", () => {
  afterEach(cleanup);

  it("uses roving focus to select the next and previous enabled tabs", async () => {
    const onChange = vi.fn();
    const user = userEvent.setup();

    render(<TabsHarness onChange={onChange} />);

    const overview = screen.getByRole("tab", { name: "Overview" });
    overview.focus();
    await user.keyboard("{ArrowRight}");

    const logs = screen.getByRole("tab", { name: "Logs" });
    expect(logs).toHaveFocus();
    expect(logs).toHaveAttribute("aria-selected", "true");
    expect(onChange).toHaveBeenLastCalledWith("logs");

    await user.keyboard("{ArrowLeft}");

    expect(overview).toHaveFocus();
    expect(overview).toHaveAttribute("aria-selected", "true");
    expect(onChange).toHaveBeenLastCalledWith("overview");
  });

  it("moves focus to the first and last enabled tabs with Home and End", async () => {
    const user = userEvent.setup();

    render(<TabsHarness onChange={vi.fn()} />);

    const logs = screen.getByRole("tab", { name: "Logs" });
    logs.focus();
    await user.keyboard("{End}");

    expect(screen.getByRole("tab", { name: "History" })).toHaveFocus();

    await user.keyboard("{Home}");

    expect(screen.getByRole("tab", { name: "Overview" })).toHaveFocus();
  });

  it("connects each tab to its required panel identifier", () => {
    render(<TabsHarness onChange={vi.fn()} />);

    const overview = screen.getByRole("tab", { name: "Overview" });
    expect(overview).toHaveAttribute("id", "tab-overview-panel");
    expect(overview).toHaveAttribute("aria-controls", "overview-panel");
    expect(overview).toHaveAttribute("tabindex", "0");
    expect(screen.getByRole("tab", { name: "Logs" })).toHaveAttribute(
      "tabindex",
      "-1",
    );
  });
});
