"use client";

import { useCallback, useEffect, useRef, useState } from "react";

import type { ActionState } from "@/lib/forms/action-state";

interface UseActionDialogOptions {
  state: ActionState<string>;
  onSuccess?: () => void;
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
}: UseActionDialogOptions): ActionDialogController {
  const [open, setOpenState] = useState(false);
  const [formKey, setFormKey] = useState(0);
  const openRef = useRef(false);
  const handledSuccess = useRef<ActionState<string> | null>(null);
  const onSuccessRef = useRef(onSuccess);

  useEffect(() => {
    onSuccessRef.current = onSuccess;
  }, [onSuccess]);

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
    reset();
  }, [reset, state]);

  return { open, setOpen, reset, formKey };
}
