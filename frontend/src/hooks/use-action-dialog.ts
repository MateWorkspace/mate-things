"use client";

import { useCallback, useEffect, useRef, useState } from "react";

import type { ActionState } from "@/lib/forms/action-state";

interface UseActionDialogOptions {
  state: ActionState<string>;
  onSuccess?: () => void;
  /**
   * Whether to close (and reset/remount) the dialog as soon as `state`
   * turns "success". Defaults to true. Set false for dialogs that show a
   * result in place - a "Done" confirmation, or a one-time secret like an
   * API key - and only close when the caller invokes `reset()` themselves
   * (e.g. from that "Done" button).
   */
  closeOnSuccess?: boolean;
}

interface ActionDialogController {
  open: boolean;
  setOpen(open: boolean): void;
  reset(): void;
  formKey: number;
}

export function useActionDialog({
  state,
  onSuccess,
  closeOnSuccess = true,
}: UseActionDialogOptions): ActionDialogController {
  const [open, setOpenState] = useState(false);
  const [formKey, setFormKey] = useState(0);
  const openRef = useRef(false);
  const handledSuccess = useRef<ActionState<string> | null>(null);
  const onSuccessRef = useRef(onSuccess);
  const closeOnSuccessRef = useRef(closeOnSuccess);

  useEffect(() => {
    onSuccessRef.current = onSuccess;
    closeOnSuccessRef.current = closeOnSuccess;
  }, [closeOnSuccess, onSuccess]);

  const setOpen = useCallback((nextOpen: boolean) => {
    openRef.current = nextOpen;
    setOpenState(nextOpen);
  }, []);

  const reset = useCallback(() => {
    const wasOpen = openRef.current;
    openRef.current = false;
    setOpenState(false);
    if (wasOpen) {
      setFormKey((key) => key + 1);
    }
  }, []);

  useEffect(() => {
    if (state.status !== "success" || handledSuccess.current === state) {
      return;
    }

    handledSuccess.current = state;
    onSuccessRef.current?.();

    if (!closeOnSuccessRef.current) {
      return;
    }

    reset();
  }, [reset, state]);

  return { open, setOpen, reset, formKey };
}
