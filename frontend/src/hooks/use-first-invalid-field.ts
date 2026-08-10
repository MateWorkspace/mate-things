"use client";

import { useEffect, type RefObject } from "react";

import type { ActionState } from "@/lib/forms/action-state";

export function useFirstInvalidField<TFields extends string>(
  state: ActionState<TFields>,
  formRef: RefObject<HTMLFormElement | null>,
): void {
  useEffect(() => {
    if (state.status !== "error" || !state.fieldErrors) {
      return;
    }

    const invalidNames = new Set(Object.keys(state.fieldErrors));
    const field = Array.from(formRef.current?.elements ?? []).find(
      (element): element is HTMLElement & { name: string } =>
        element instanceof HTMLElement &&
        "name" in element &&
        typeof element.name === "string" &&
        invalidNames.has(element.name),
    );
    field?.focus();
  }, [formRef, state]);
}
