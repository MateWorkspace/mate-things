"use client";

import { useActionState, useEffect, useRef } from "react";

import Button from "@/components/ui/button";
import Input from "@/components/ui/input";
import Label from "@/components/ui/label";
import { useToast } from "@/hooks/use-toast";

import { loginAction, type LoginFormState } from "../_lib/actions";

const initialState: LoginFormState = {};

export default function LoginForm() {
  const [state, formAction, isPending] = useActionState(
    loginAction,
    initialState,
  );
  const toast = useToast();
  const lastShown = useRef<LoginFormState | null>(null);

  useEffect(() => {
    if (state.title && state !== lastShown.current) {
      lastShown.current = state;
      toast.error(state.title, state.message ?? "");
    }
  }, [state, toast]);

  return (
    <form action={formAction} className="flex flex-col gap-5">
      <div>
        <Label htmlFor="username">Username</Label>
        <Input
          id="username"
          name="username"
          autoComplete="username"
          defaultValue={state.username}
          autoFocus
          required
        />
      </div>

      <div>
        <Label htmlFor="password">Password</Label>
        <Input
          id="password"
          name="password"
          type="password"
          autoComplete="current-password"
          required
        />
      </div>

      <Button type="submit" disabled={isPending} className="mt-1 w-full">
        {isPending ? "Signing in…" : "Sign in"}
      </Button>
    </form>
  );
}
