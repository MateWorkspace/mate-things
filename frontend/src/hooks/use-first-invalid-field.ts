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
    const field = Array.from(
      formRef.current?.querySelectorAll<HTMLElement>(
        "[name], [data-field-name], [data-field-names]",
      ) ?? [],
    ).find((element) => {
      if (
        (element instanceof HTMLInputElement && element.type === "hidden") ||
        element.matches(":disabled")
      ) {
        return false;
      }

      const names = [
        element.getAttribute("name"),
        element.dataset.fieldName,
        ...(element.dataset.fieldNames?.split(/\s+/) ?? []),
      ].filter((name): name is string => Boolean(name));

      return names.some((name) => invalidNames.has(name));
    });
    field?.focus();
  }, [formRef, state]);
}
