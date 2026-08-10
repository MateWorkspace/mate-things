import { createRef } from "react";
import { cleanup, render, screen } from "@testing-library/react";
import { afterEach, expect, it, vi } from "vitest";

import ActionMessage from "@/components/forms/ActionMessage";
import FieldError from "@/components/forms/FieldError";
import { ToastContext } from "@/components/ui/toast-provider";
import type { ActionState } from "@/lib/forms/action-state";

import { useActionFeedback } from "./use-action-feedback";
import { useFirstInvalidField } from "./use-first-invalid-field";
import { useRefreshAfterAction } from "./use-refresh-after-action";

const refresh = vi.fn();

vi.mock("next/navigation", () => ({
  useRouter: () => ({ refresh }),
}));

function FeedbackHarness({ state }: { state: ActionState }) {
  useActionFeedback(state);
  return null;
}

function RefreshHarness({ state }: { state: { status: string } }) {
  useRefreshAfterAction(state);
  return null;
}

function FocusHarness({ state }: { state: ActionState<"name" | "email"> }) {
  const formRef = createRef<HTMLFormElement>();
  useFirstInvalidField(state, formRef);
  return (
    <form ref={formRef}>
      <input aria-label="Name" name="name" />
      <input aria-label="Email" name="email" />
    </form>
  );
}

function CompositeFocusHarness({ state }: { state: ActionState<"user_id"> }) {
  const formRef = createRef<HTMLFormElement>();
  useFirstInvalidField(state, formRef);
  return (
    <form ref={formRef}>
      <input type="hidden" name="user_id" />
      <button type="button" data-field-name="user_id">
        Select user
      </button>
    </form>
  );
}

afterEach(() => {
  cleanup();
  vi.clearAllMocks();
});

it("shows feedback once per result object, including consecutive successes", () => {
  const success = vi.fn();
  const error = vi.fn();
  const first = {
    status: "success",
    title: "Saved",
    message: "First",
  } satisfies ActionState;
  const second = {
    status: "success",
    title: "Saved",
    message: "Second",
  } satisfies ActionState;
  const { rerender } = render(
    <ToastContext.Provider value={{ success, error }}>
      <FeedbackHarness state={first} />
    </ToastContext.Provider>,
  );

  rerender(
    <ToastContext.Provider value={{ success, error }}>
      <FeedbackHarness state={first} />
    </ToastContext.Provider>,
  );
  rerender(
    <ToastContext.Provider value={{ success, error }}>
      <FeedbackHarness state={second} />
    </ToastContext.Provider>,
  );

  expect(success).toHaveBeenNthCalledWith(1, "Saved", "First");
  expect(success).toHaveBeenNthCalledWith(2, "Saved", "Second");
  expect(error).not.toHaveBeenCalled();
});

it("refreshes authoritative data for success and partial failure results", () => {
  const first = { status: "success" };
  const partial = { status: "partial" };
  const { rerender } = render(<RefreshHarness state={first} />);

  rerender(<RefreshHarness state={first} />);
  rerender(<RefreshHarness state={partial} />);

  expect(refresh).toHaveBeenCalledTimes(2);
});

it("focuses the first invalid field in form order", () => {
  render(
    <FocusHarness
      state={{
        status: "error",
        fieldErrors: { email: "Required", name: "Required" },
      }}
    />,
  );

  expect(screen.getByLabelText("Name")).toHaveFocus();
});

it("skips hidden inputs and focuses a composite field control", () => {
  render(
    <CompositeFocusHarness
      state={{
        status: "error",
        fieldErrors: { user_id: "Required" },
      }}
    />,
  );

  expect(screen.getByRole("button", { name: "Select user" })).toHaveFocus();
});

it("renders consistent action and field messages", () => {
  render(
    <>
      <ActionMessage state={{ status: "error", message: "Try again." }} />
      <FieldError id="name-error">Enter a name.</FieldError>
    </>,
  );

  expect(screen.getByText("Try again.")).toHaveAttribute(
    "aria-live",
    "assertive",
  );
  expect(screen.getByText("Enter a name.")).toHaveAttribute("id", "name-error");
});
