import { render, screen } from "@testing-library/react";
import { describe, expect, it, vi } from "vitest";

import Forbidden from "./forbidden";

vi.mock("@/lib/api/session-actions", () => ({
  logoutAction: vi.fn(),
}));

describe("Forbidden", () => {
  it("renders an authenticated-scope access-denied state with logout", () => {
    render(<Forbidden />);

    expect(
      screen.getByRole("heading", { name: "Access denied" }),
    ).toBeVisible();
    expect(
      screen.getByText(
        "Your account is signed in, but it does not have permission to view this page.",
      ),
    ).toBeVisible();
    expect(screen.getByRole("button", { name: "Log out" })).toBeVisible();
  });
});
