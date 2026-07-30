"use client";

import { useActionState, useContext, useEffect, useRef } from "react";

import Button from "@/components/ui/button";
import Input from "@/components/ui/input";
import Label from "@/components/ui/label";
import { ToastContext } from "@/components/ui/toast-provider";

import { changePasswordAction, type FormActionState } from "./profile-actions";

const INITIAL_STATE: FormActionState = { status: "idle" };

export default function SecurityForm() {
  const [state, formAction, isPending] = useActionState(
    changePasswordAction,
    INITIAL_STATE,
  );
  const formRef = useRef<HTMLFormElement>(null);
  const lastShown = useRef<FormActionState | null>(null);
  const toast = useContext(ToastContext);

  useEffect(() => {
    if (
      state.status === "idle" ||
      !state.title ||
      state === lastShown.current
    ) {
      return;
    }

    lastShown.current = state;
    const message = state.message ?? "";
    if (state.status === "success") {
      formRef.current?.reset();
      toast?.success(state.title, message);
    } else {
      toast?.error(state.title, message);
    }
  }, [state, toast]);

  return (
    <form ref={formRef} action={formAction} className="space-y-4">
      <div>
        <Label htmlFor="current-password">Current password</Label>
        <Input
          id="current-password"
          name="current_password"
          type="password"
          autoComplete="current-password"
          required
        />
      </div>

      <div>
        <Label htmlFor="new-password">New password</Label>
        <Input
          id="new-password"
          name="new_password"
          type="password"
          autoComplete="new-password"
          required
        />
      </div>

      <div>
        <Label htmlFor="confirm-password">Confirm new password</Label>
        <Input
          id="confirm-password"
          name="confirm_password"
          type="password"
          autoComplete="new-password"
          required
          aria-invalid={Boolean(state.fieldErrors?.confirm_password)}
          aria-describedby={
            state.fieldErrors?.confirm_password
              ? "confirm-password-error"
              : undefined
          }
        />
        {state.fieldErrors?.confirm_password ? (
          <p
            id="confirm-password-error"
            className="text-critical mt-1.5 text-sm"
          >
            {state.fieldErrors.confirm_password}
          </p>
        ) : null}
      </div>

      <p
        aria-live="polite"
        className={
          state.status === "error"
            ? "text-critical text-sm"
            : "text-success text-sm"
        }
      >
        {state.message}
      </p>

      <div className="flex justify-end">
        <Button type="submit" disabled={isPending}>
          {isPending ? "Changing…" : "Change password"}
        </Button>
      </div>
    </form>
  );
}
