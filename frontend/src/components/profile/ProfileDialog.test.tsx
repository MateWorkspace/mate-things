import { cleanup, render, screen, within } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { afterEach, describe, expect, it, vi } from "vitest";

import AppBar from "@/components/layout/AppBar";
import ToastProvider from "@/components/ui/toast-provider";
import { USER } from "@/test/fixtures";

import ProfileDialog from "./ProfileDialog";
import { saveProfileAction } from "./profile-actions";

vi.mock("./profile-actions", () => ({
  saveProfileAction: vi.fn(),
  changePasswordAction: vi.fn(),
}));

vi.mock("@/lib/api/session-actions", () => ({
  logoutAction: vi.fn(),
}));

describe("ProfileDialog", () => {
  afterEach(cleanup);

  it("opens from the profile trigger and enters edit mode", async () => {
    const user = userEvent.setup();

    render(<AppBar user={USER} permissions={["profile:get", "profile:set"]} />);

    await user.click(
      screen.getByRole("button", { name: /Open profile for Alex Morgan/ }),
    );

    expect(screen.getByRole("dialog", { name: "Your profile" })).toBeVisible();

    await user.click(screen.getByRole("button", { name: "Edit profile" }));

    expect(screen.getByLabelText("Name")).toHaveValue("Alex Morgan");
  });

  it("omits edit controls without profile:set", () => {
    render(
      <ProfileDialog
        open
        user={USER}
        permissions={["profile:get"]}
        onClose={() => undefined}
      />,
    );

    expect(
      screen.queryByRole("button", { name: "Edit profile" }),
    ).not.toBeInTheDocument();
  });

  it("summarizes the anonymous identity, role identifier, and effective permissions", () => {
    render(
      <ProfileDialog
        open
        user={USER}
        permissions={["profile:get", "profile:set"]}
        onClose={() => undefined}
      />,
    );

    expect(
      screen.getByRole("img", { name: "Anonymous profile icon" }),
    ).toBeVisible();
    expect(screen.getByText(USER.role_id)).toBeVisible();
    expect(screen.getByText("profile:get")).toBeVisible();
    expect(screen.getByText("profile:set")).toBeVisible();
  });

  it("uses a viewport-bounded mobile sheet that recenters on desktop", () => {
    render(
      <ProfileDialog
        open
        user={USER}
        permissions={["profile:get"]}
        onClose={() => undefined}
      />,
    );

    expect(screen.getByRole("dialog", { name: "Your profile" })).toHaveClass(
      "bottom-0",
      "max-h-[calc(100dvh-1rem)]",
      "overflow-y-auto",
      "sm:inset-0",
      "sm:m-auto",
    );
  });

  it("returns to Profile when the Security permission is removed", async () => {
    const user = userEvent.setup();
    const { rerender } = render(
      <ProfileDialog
        open
        user={USER}
        permissions={["profile:get", "profile:set", "profile_security:set"]}
        onClose={() => undefined}
      />,
    );

    await user.click(screen.getByRole("tab", { name: "Security" }));
    expect(screen.getByLabelText("Current password")).toBeVisible();

    rerender(
      <ProfileDialog
        open
        user={USER}
        permissions={["profile:get", "profile:set"]}
        onClose={() => undefined}
      />,
    );

    const profileTab = screen.getByRole("tab", { name: "Profile" });
    expect(profileTab).toHaveAttribute("aria-selected", "true");
    expect(profileTab).toHaveAttribute("tabindex", "0");
    expect(
      screen.queryByRole("tab", { name: "Security" }),
    ).not.toBeInTheDocument();
    expect(screen.queryByLabelText("Current password")).not.toBeInTheDocument();
  });

  it("announces a saved profile inline and in a toast", async () => {
    const user = userEvent.setup();
    vi.mocked(saveProfileAction).mockResolvedValue({
      status: "success",
      title: "Profile updated",
      message: "Your profile details have been saved.",
    });

    render(
      <ToastProvider>
        <ProfileDialog
          open
          user={USER}
          permissions={["profile:get", "profile:set"]}
          onClose={() => undefined}
        />
      </ToastProvider>,
    );

    await user.click(screen.getByRole("button", { name: "Edit profile" }));
    await user.click(screen.getByRole("button", { name: "Save profile" }));

    const dialog = screen.getByRole("dialog", { name: "Your profile" });
    expect(
      await within(dialog).findByText("Your profile details have been saved."),
    ).toBeVisible();
    expect(await screen.findByRole("alert")).toHaveTextContent(
      "Profile updated",
    );
  });
});
