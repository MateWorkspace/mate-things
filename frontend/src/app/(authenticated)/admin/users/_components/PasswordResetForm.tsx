"use client";

import { useActionState, useRef, useState } from "react";

import ActionMessage from "@/components/forms/ActionMessage";
import FieldError from "@/components/forms/FieldError";
import Button from "@/components/ui/button";
import Dialog from "@/components/ui/dialog";
import Input from "@/components/ui/input";
import Label from "@/components/ui/label";
import type { UserResponse } from "@/lib/api/users";
import { useFirstInvalidField } from "@/hooks/use-first-invalid-field";

import { resetUserPasswordAction } from "../_lib/actions";
import { EMPTY_USER_STATE } from "../_lib/state";

export default function PasswordResetForm({ user }: { user: UserResponse }) {
  const [open, setOpen] = useState(false);
  const [generation, setGeneration] = useState(0);
  return (
    <>
      <Button
        type="button"
        variant="secondary"
        onClick={() => {
          setGeneration((current) => current + 1);
          setOpen(true);
        }}
      >
        Reset password
      </Button>
      <PasswordResetDialog
        key={generation}
        user={user}
        open={open}
        onClose={() => setOpen(false)}
      />
    </>
  );
}

function PasswordResetDialog({
  user,
  open,
  onClose,
}: {
  user: UserResponse;
  open: boolean;
  onClose: () => void;
}) {
  const [state, action, pending] = useActionState(
    resetUserPasswordAction,
    EMPTY_USER_STATE,
  );
  const [password, setPassword] = useState("");
  const [passwordConfirmation, setPasswordConfirmation] = useState("");
  const formRef = useRef<HTMLFormElement>(null);
  useFirstInvalidField(state, formRef);
  return (
    <Dialog
      open={open && state.status !== "success"}
      onClose={onClose}
      title={`Reset ${user.username}'s password`}
      variant="sheet"
      dismissible={!pending}
    >
      <form
        ref={formRef}
        action={action}
        onReset={(event) => event.preventDefault()}
        className="space-y-4"
      >
        <input type="hidden" name="user_id" value={user.id} />
        <p className="text-muted-foreground text-sm">
          You are changing the credentials for{" "}
          <strong className="text-foreground">
            {user.name} (@{user.username})
          </strong>
          .
        </p>
        <div>
          <Label htmlFor="admin-new-password">New password</Label>
          <Input
            id="admin-new-password"
            name="password"
            type="password"
            autoComplete="new-password"
            required
            value={password}
            onChange={(event) => setPassword(event.target.value)}
            aria-invalid={Boolean(state.fieldErrors?.password)}
            aria-describedby={
              state.fieldErrors?.password
                ? "admin-new-password-error"
                : undefined
            }
          />
          <FieldError id="admin-new-password-error">
            {state.fieldErrors?.password}
          </FieldError>
        </div>
        <div>
          <Label htmlFor="admin-confirm-password">Confirm password</Label>
          <Input
            id="admin-confirm-password"
            name="password_confirmation"
            type="password"
            autoComplete="new-password"
            required
            value={passwordConfirmation}
            onChange={(event) => setPasswordConfirmation(event.target.value)}
            aria-invalid={Boolean(state.fieldErrors?.password_confirmation)}
            aria-describedby={
              state.fieldErrors?.password_confirmation
                ? "admin-confirm-password-error"
                : undefined
            }
          />
          <FieldError id="admin-confirm-password-error">
            {state.fieldErrors?.password_confirmation}
          </FieldError>
        </div>
        <ActionMessage state={state} />
        <div className="flex justify-end gap-2">
          <Button
            type="button"
            variant="secondary"
            disabled={pending}
            onClick={onClose}
          >
            Cancel
          </Button>
          <Button type="submit" disabled={pending}>
            {pending ? "Resetting…" : "Reset password"}
          </Button>
        </div>
      </form>
    </Dialog>
  );
}
