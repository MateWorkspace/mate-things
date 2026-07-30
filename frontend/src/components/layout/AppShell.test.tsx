import { cleanup, render, screen, within } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { afterEach, describe, expect, it, vi } from "vitest";

import { USER } from "@/test/fixtures";
import { setSidebarCollapsedAction } from "@/lib/api/session-actions";

import AppShell from "./AppShell";

vi.mock("next/navigation", () => ({
  usePathname: () => "/nodes",
}));

vi.mock("@/lib/api/session-actions", () => ({
  setSidebarCollapsedAction: vi.fn().mockResolvedValue(undefined),
  logoutAction: vi.fn().mockResolvedValue(undefined),
}));

vi.mock("@/components/profile/profile-actions", () => ({
  saveProfileAction: vi.fn(),
  changePasswordAction: vi.fn(),
}));

describe("AppShell", () => {
  afterEach(cleanup);

  it("toggles labels on desktop and keeps permitted links", async () => {
    const user = userEvent.setup();

    render(
      <AppShell user={USER} permissions={["node:get"]}>
        <main>Page</main>
      </AppShell>,
    );

    expect(screen.getByRole("link", { name: "Nodes" })).toBeVisible();
    expect(
      screen.queryByRole("link", { name: "Users" }),
    ).not.toBeInTheDocument();

    await user.click(screen.getByRole("button", { name: "Collapse sidebar" }));

    expect(
      screen.getByRole("button", { name: "Expand sidebar" }),
    ).toBeVisible();
    expect(setSidebarCollapsedAction).toHaveBeenCalledWith(true);
  });

  it("opens mobile navigation in a modal drawer and provides a close control", async () => {
    const user = userEvent.setup();

    render(
      <AppShell user={USER} permissions={["node:get"]}>
        <main>Page</main>
      </AppShell>,
    );

    await user.click(screen.getByRole("button", { name: "Open navigation" }));

    const drawer = screen.getByRole("dialog", { name: "Navigation" });
    expect(within(drawer).getByRole("link", { name: "Nodes" })).toHaveAttribute(
      "aria-current",
      "page",
    );

    await user.click(
      within(drawer).getByRole("button", { name: "Close navigation" }),
    );

    expect(
      screen.queryByRole("dialog", { name: "Navigation" }),
    ).not.toBeInTheDocument();
  });

  it("restores the persisted desktop collapse preference", () => {
    render(
      <AppShell user={USER} permissions={["node:get"]} initialSidebarCollapsed>
        <main>Page</main>
      </AppShell>,
    );

    expect(
      screen.getByRole("button", { name: "Expand sidebar" }),
    ).toBeVisible();
  });

  it("keeps the profile name visible in the mobile trigger", () => {
    render(
      <AppShell user={USER} permissions={[]}>
        <main>Page</main>
      </AppShell>,
    );

    const profileTrigger = screen.getByRole("button", {
      name: `Open profile for ${USER.name}`,
    });
    const profileName = within(profileTrigger).getByText(USER.name);

    expect(profileName).not.toHaveClass("hidden");
  });
});
