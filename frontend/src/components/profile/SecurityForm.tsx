"use client";

import {
  useActionState,
  useEffect,
  useLayoutEffect,
  useRef,
  useState,
} from "react";

import ActionMessage from "@/components/forms/ActionMessage";
import FieldError from "@/components/forms/FieldError";
import Button from "@/components/ui/button";
import Input from "@/components/ui/input";
import Label from "@/components/ui/label";
import { useActionFeedback } from "@/hooks/use-action-feedback";
import { useFirstInvalidField } from "@/hooks/use-first-invalid-field";

import { changePasswordAction, type FormActionState } from "./profile-actions";

const INITIAL_STATE: FormActionState = { status: "idle" };

export default function SecurityForm({
  onPendingChange,
}: {
  onPendingChange?: (pending: boolean) => void;
}) {
  const [state, formAction, isPending] = useActionState(
    changePasswordAction,
    INITIAL_STATE,
  );
  const [currentPassword, setCurrentPassword] = useState("");
  const [newPassword, setNewPassword] = useState("");
  const [confirmPassword, setConfirmPassword] = useState("");
  const formRef = useRef<HTMLFormElement>(null);
  useActionFeedback(state);
  useFirstInvalidField(state, formRef);
  useEffect(() => {
    onPendingChange?.(isPending);
    return () => onPendingChange?.(false);
  }, [isPending, onPendingChange]);

  useLayoutEffect(() => {
    if (!formRef.current || state.status === "idle") return;

    const fields = formRef.current.elements;
    (fields.namedItem("current_password") as HTMLInputElement).value =
      state.status === "error" ? currentPassword : "";
    (fields.namedItem("new_password") as HTMLInputElement).value =
      state.status === "error" ? newPassword : "";
    (fields.namedItem("confirm_password") as HTMLInputElement).value =
      state.status === "error" ? confirmPassword : "";
  }, [confirmPassword, currentPassword, newPassword, state]);

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
          defaultValue={currentPassword}
          onChange={(event) => setCurrentPassword(event.target.value)}
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
          defaultValue={newPassword}
          onChange={(event) => setNewPassword(event.target.value)}
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
          defaultValue={confirmPassword}
          onChange={(event) => setConfirmPassword(event.target.value)}
          aria-invalid={Boolean(state.fieldErrors?.confirm_password)}
          aria-describedby={
            state.fieldErrors?.confirm_password
              ? "confirm-password-error"
              : undefined
          }
        />
        <FieldError id="confirm-password-error">
          {state.fieldErrors?.confirm_password}
        </FieldError>
      </div>

      <ActionMessage state={state} />

      <div className="flex justify-end">
        <Button type="submit" disabled={isPending}>
          {isPending ? "Changing…" : "Change password"}
        </Button>
      </div>
    </form>
  );
}
