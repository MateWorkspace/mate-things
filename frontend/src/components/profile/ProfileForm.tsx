"use client";

import { useActionState, useEffect, useRef } from "react";

import ActionMessage from "@/components/forms/ActionMessage";
import FieldError from "@/components/forms/FieldError";
import Button from "@/components/ui/button";
import Input from "@/components/ui/input";
import Label from "@/components/ui/label";
import { useActionFeedback } from "@/hooks/use-action-feedback";
import { useFirstInvalidField } from "@/hooks/use-first-invalid-field";
import type { UserResponse } from "@/lib/api/users";

import { saveProfileAction, type FormActionState } from "./profile-actions";

const INITIAL_STATE: FormActionState = { status: "idle" };

interface ProfileFormProps {
  user: UserResponse;
  onCancel: () => void;
  onPendingChange?: (pending: boolean) => void;
}

export default function ProfileForm({
  user,
  onCancel,
  onPendingChange,
}: ProfileFormProps) {
  const [state, formAction, isPending] = useActionState(
    saveProfileAction,
    INITIAL_STATE,
  );
  const formRef = useRef<HTMLFormElement>(null);
  useActionFeedback(state);
  useFirstInvalidField(state, formRef);
  useEffect(() => {
    onPendingChange?.(isPending);
    return () => onPendingChange?.(false);
  }, [isPending, onPendingChange]);

  return (
    <form
      ref={formRef}
      action={formAction}
      onReset={(event) => event.preventDefault()}
      className="space-y-4"
    >
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
        <FieldError id="profile-name-error">
          {state.fieldErrors?.name}
        </FieldError>
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
        <FieldError id="profile-username-error">
          {state.fieldErrors?.username}
        </FieldError>
      </div>

      <div>
        <Label htmlFor="profile-bio">Bio</Label>
        <textarea
          id="profile-bio"
          name="bio"
          defaultValue={user.bio}
          rows={4}
          className="border-control-border bg-background text-foreground placeholder:text-foreground/40 focus-visible:border-focus focus-visible:ring-focus w-full resize-y rounded-xl border px-3.5 py-2.5 text-sm transition-colors focus-visible:ring-2 focus-visible:outline-none"
        />
      </div>

      <ActionMessage state={state} />

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
