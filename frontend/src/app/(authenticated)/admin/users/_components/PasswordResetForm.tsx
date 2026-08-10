"use client";

import { useActionState, useState } from "react";

import ActionMessage from "@/components/forms/ActionMessage";
import Button from "@/components/ui/button";
import Dialog from "@/components/ui/dialog";
import Input from "@/components/ui/input";
import Label from "@/components/ui/label";
import type { UserResponse } from "@/lib/api/users";

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
  return (
    <Dialog
      open={open && state.status !== "success"}
      onClose={onClose}
      title={`Reset ${user.username}'s password`}
      variant="sheet"
      dismissible={!pending}
    >
      <form
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
          />
        </div>
        <div>
          <Label htmlFor="admin-confirm-password">Confirm password</Label>
          <Input
            id="admin-confirm-password"
            name="password_confirmation"
            type="password"
            autoComplete="new-password"
            required
          />
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
