"use client";

import { useActionState, useContext, useEffect, useRef } from "react";

import Button from "@/components/ui/button";
import Input from "@/components/ui/input";
import Label from "@/components/ui/label";
import { ToastContext } from "@/components/ui/toast-provider";
import type { UserResponse } from "@/lib/api/users";

import { saveProfileAction, type FormActionState } from "./profile-actions";

const INITIAL_STATE: FormActionState = { status: "idle" };

interface ProfileFormProps {
  user: UserResponse;
  onCancel: () => void;
}

export default function ProfileForm({ user, onCancel }: ProfileFormProps) {
  const [state, formAction, isPending] = useActionState(
    saveProfileAction,
    INITIAL_STATE,
  );
  const toast = useContext(ToastContext);
  const lastShown = useRef<FormActionState | null>(null);

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
      toast?.success(state.title, message);
    } else {
      toast?.error(state.title, message);
    }
  }, [state, toast]);

  return (
    <form action={formAction} className="space-y-4">
      <div>
        <Label htmlFor="profile-name">Name</Label>
        <Input
          id="profile-name"
          name="name"
          defaultValue={user.name}
          autoComplete="name"
          required
          aria-invalid={Boolean(state.fieldErrors?.name)}
          aria-describedby={
            state.fieldErrors?.name ? "profile-name-error" : undefined
          }
        />
        {state.fieldErrors?.name ? (
          <p id="profile-name-error" className="text-critical mt-1.5 text-sm">
            {state.fieldErrors.name}
          </p>
        ) : null}
      </div>

      <div>
        <Label htmlFor="profile-username">Username</Label>
        <Input
          id="profile-username"
          name="username"
          defaultValue={user.username}
          autoComplete="username"
          required
          aria-invalid={Boolean(state.fieldErrors?.username)}
          aria-describedby={
            state.fieldErrors?.username ? "profile-username-error" : undefined
          }
        />
        {state.fieldErrors?.username ? (
          <p
            id="profile-username-error"
            className="text-critical mt-1.5 text-sm"
          >
            {state.fieldErrors.username}
          </p>
        ) : null}
      </div>

      <div>
        <Label htmlFor="profile-bio">Bio</Label>
        <textarea
          id="profile-bio"
          name="bio"
          defaultValue={user.bio}
          rows={4}
          className="border-border bg-background text-foreground placeholder:text-foreground/40 focus-visible:border-focus focus-visible:ring-focus w-full resize-y rounded-xl border px-3.5 py-2.5 text-sm transition-colors focus-visible:ring-2 focus-visible:outline-none"
        />
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

      <div className="flex flex-wrap justify-end gap-2">
        <Button
          type="button"
          variant="secondary"
          onClick={onCancel}
          disabled={isPending}
        >
          Cancel
        </Button>
        <Button type="submit" disabled={isPending}>
          {isPending ? "Saving…" : "Save profile"}
        </Button>
      </div>
    </form>
  );
}
