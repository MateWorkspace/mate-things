import {
  act,
  cleanup,
  render,
  screen,
  waitFor,
  within,
} from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";

import type { NodeClassResponse } from "@/lib/api/node-classes";

import {
  createNodeClassAction,
  deleteNodeClassAction,
  type FormActionState,
} from "../_lib/actions";
import NodeClassForm from "./NodeClassForm";

vi.mock("../_lib/actions", () => ({
  createNodeClassAction: vi.fn(),
  deleteNodeClassAction: vi.fn(),
  updateNodeClassAction: vi.fn(),
}));

const NODE_CLASS: NodeClassResponse = {
  id: "class-1",
  name: "Cold Storage",
  description: "Temperature-controlled sensors",
  preferences: {},
  created_at: "2026-07-30T00:00:00Z",
};

function deferredAction() {
  let resolve!: (state: FormActionState) => void;
  const promise = new Promise<FormActionState>((promiseResolve) => {
    resolve = promiseResolve;
  });

  return { promise, resolve };
}

describe("NodeClassForm", () => {
  beforeEach(() => {
    vi.resetAllMocks();
  });

  afterEach(() => {
    cleanup();
    vi.restoreAllMocks();
  });

  it("opens a modal with the create fields", async () => {
    const user = userEvent.setup();
    render(<NodeClassForm />);

    await user.click(screen.getByRole("button", { name: "Create node class" }));

    const dialog = screen.getByRole("dialog", {
      name: "Create node class",
    });
    expect(within(dialog).getByLabelText("Name")).toBeRequired();
    expect(within(dialog).getByLabelText("Description")).toBeVisible();
    expect(
      within(dialog).getByRole("button", { name: "Create class" }),
    ).toBeVisible();
  });

  it("opens a prefilled edit modal for an existing class", async () => {
    const user = userEvent.setup();
    render(<NodeClassForm nodeClass={NODE_CLASS} />);

    await user.click(screen.getByRole("button", { name: "Edit Cold Storage" }));

    const dialog = screen.getByRole("dialog", {
      name: "Edit Cold Storage",
    });
    expect(within(dialog).getByLabelText("Name")).toHaveValue("Cold Storage");
    expect(within(dialog).getByLabelText("Description")).toHaveValue(
      "Temperature-controlled sensors",
    );
  });

  it("uses a separate named confirmation and honest dependency warning", async () => {
    const user = userEvent.setup();
    render(<NodeClassForm canDelete nodeClass={NODE_CLASS} />);

    await user.click(
      screen.getByRole("button", { name: "Delete Cold Storage" }),
    );

    const dialog = screen.getByRole("dialog", {
      name: "Delete Cold Storage",
    });
    expect(
      within(dialog).getByText(/dependent nodes, firmware, or actions/i),
    ).toBeVisible();
    expect(
      within(dialog).getByText(/backend may reject this deletion/i),
    ).toBeVisible();
    expect(within(dialog).getByLabelText("Confirm class name")).toHaveAttribute(
      "placeholder",
      "Cold Storage",
    );
  });

  it("mounts editor and delete dialogs without duplicate React keys", () => {
    const consoleError = vi
      .spyOn(console, "error")
      .mockImplementation(() => undefined);

    render(<NodeClassForm canDelete nodeClass={NODE_CLASS} />);

    expect(consoleError).not.toHaveBeenCalled();
  });

  it("restores trigger focus through a controlled editor close", async () => {
    const user = userEvent.setup();
    render(<NodeClassForm />);
    const trigger = screen.getByRole("button", {
      name: "Create node class",
    });

    await user.click(trigger);
    await user.click(screen.getByRole("button", { name: "Cancel" }));

    expect(
      screen.queryByRole("dialog", { name: "Create node class" }),
    ).not.toBeInTheDocument();
    expect(trigger).toHaveFocus();
  });

  it("submits the rendered class name as the typed-confirmation interlock", async () => {
    const user = userEvent.setup();
    vi.mocked(deleteNodeClassAction).mockResolvedValue({
      status: "error",
      title: "Deletion rejected",
    });
    render(<NodeClassForm canDelete nodeClass={NODE_CLASS} />);

    await user.click(
      screen.getByRole("button", { name: "Delete Cold Storage" }),
    );
    await user.type(
      screen.getByLabelText("Confirm class name"),
      "Cold Storage",
    );
    await user.click(
      screen.getByRole("button", { name: "Permanently delete class" }),
    );

    await waitFor(() => {
      expect(deleteNodeClassAction).toHaveBeenCalledOnce();
    });
    const submitted = vi.mocked(deleteNodeClassAction).mock.calls[0][1];
    expect(submitted.get("node_class_name")).toBe("Cold Storage");
  });

  it("keeps the editor open on Escape while creation is pending", async () => {
    const user = userEvent.setup();
    const deferred = deferredAction();
    vi.mocked(createNodeClassAction).mockReturnValue(deferred.promise);
    render(<NodeClassForm />);
    const trigger = screen.getByRole("button", {
      name: "Create node class",
    });

    await user.click(trigger);
    await user.type(screen.getByLabelText("Name"), "Cold Storage");
    await user.click(screen.getByRole("button", { name: "Create class" }));
    expect(
      await screen.findByRole("button", { name: "Saving…" }),
    ).toBeDisabled();

    await user.keyboard("{Escape}");

    expect(
      screen.getByRole("dialog", { name: "Create node class" }),
    ).toBeVisible();

    await act(async () => {
      deferred.resolve({
        status: "success",
        title: "Node class created",
      });
      await deferred.promise;
    });
    await waitFor(() => {
      expect(
        screen.queryByRole("dialog", { name: "Create node class" }),
      ).not.toBeInTheDocument();
    });
    expect(trigger).toHaveFocus();
  });

  it("keeps delete confirmation open on Escape while deletion is pending", async () => {
    const user = userEvent.setup();
    const deferred = deferredAction();
    vi.mocked(deleteNodeClassAction).mockReturnValue(deferred.promise);
    render(<NodeClassForm canDelete nodeClass={NODE_CLASS} />);
    const trigger = screen.getByRole("button", {
      name: "Delete Cold Storage",
    });

    await user.click(trigger);
    await user.type(
      screen.getByLabelText("Confirm class name"),
      "Cold Storage",
    );
    await user.click(
      screen.getByRole("button", { name: "Permanently delete class" }),
    );
    expect(
      await screen.findByRole("button", { name: "Deleting…" }),
    ).toBeDisabled();

    await user.keyboard("{Escape}");

    expect(
      screen.getByRole("dialog", { name: "Delete Cold Storage" }),
    ).toBeVisible();

    await act(async () => {
      deferred.resolve({
        status: "error",
        title: "Deletion rejected",
        message: "Dependent resources still exist.",
      });
      await deferred.promise;
    });
    expect(
      await screen.findByText("Dependent resources still exist."),
    ).toBeVisible();

    await user.keyboard("{Escape}");

    await waitFor(() => {
      expect(
        screen.queryByRole("dialog", { name: "Delete Cold Storage" }),
      ).not.toBeInTheDocument();
    });
    expect(trigger).toHaveFocus();
  });
});
