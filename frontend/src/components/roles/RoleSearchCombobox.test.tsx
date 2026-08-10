import { fireEvent, render, screen, within } from "@testing-library/react";
import { beforeEach, describe, expect, it, vi } from "vitest";

import { searchRolesAction } from "@/lib/actions/entity-search-actions";

import RoleSearchCombobox from "./RoleSearchCombobox";

vi.mock("@/lib/actions/entity-search-actions", () => ({
  searchRolesAction: vi.fn(),
}));

describe("RoleSearchCombobox", () => {
  beforeEach(() => {
    vi.resetAllMocks();
    vi.mocked(searchRolesAction).mockResolvedValue({
      items: [
        {
          value: "role-1",
          label: "Operator",
          description: "Fleet operator",
          annotation: "(default)",
        },
      ],
      page: 1,
      totalPages: 1,
    });
  });

  it("shows the default annotation in its option but selects the plain role name", async () => {
    const { container } = render(<RoleSearchCombobox name="role_id" />);

    fireEvent.click(screen.getByRole("button", { name: "Any role" }));

    const defaultRole = await screen.findByRole("option", {
      name: "Operator(default)",
    });
    expect(within(defaultRole).getByText("(default)")).toBeVisible();
    fireEvent.click(defaultRole);

    expect(screen.getByRole("button", { name: "Operator" })).toBeVisible();
    expect(screen.queryByText("(default)")).not.toBeInTheDocument();
    expect(
      container.querySelector<HTMLInputElement>('input[name="role_id"]'),
    ).toHaveValue("role-1");
  });
});
