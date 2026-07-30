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
