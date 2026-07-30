"use client";

import { useEffect, useId, useRef, type ReactNode } from "react";
import { X } from "lucide-react";

import IconButton from "./icon-button";

type DialogVariant = "default" | "drawer";

interface DialogProps {
  open: boolean;
  onClose: () => void;
  title: string;
  children: ReactNode;
  variant?: DialogVariant;
}

export default function Dialog({
  open,
  onClose,
  title,
  children,
  variant = "default",
}: DialogProps) {
  const dialogRef = useRef<HTMLDialogElement>(null);
  const previousFocusRef = useRef<HTMLElement | null>(null);
  const closingFromPropRef = useRef(false);
  const titleId = useId();

  const restoreFocus = () => {
    previousFocusRef.current?.focus();
    previousFocusRef.current = null;
  };

  useEffect(() => {
    const dialog = dialogRef.current;

    if (!dialog) {
      return;
    }

    if (open) {
      previousFocusRef.current =
        document.activeElement instanceof HTMLElement
          ? document.activeElement
          : null;

      if (!dialog.open) {
        dialog.showModal();
      }

      dialog.focus();
      return;
    }

    if (dialog.open) {
      closingFromPropRef.current = true;
      dialog.close();
      closingFromPropRef.current = false;
    }

    restoreFocus();
  }, [open]);

  useEffect(() => {
    if (!open) {
      return;
    }

    const handleKeyDown = (event: KeyboardEvent) => {
      if (event.key === "Escape") {
        event.preventDefault();
        onClose();
      }
    };

    window.addEventListener("keydown", handleKeyDown, true);
    return () => window.removeEventListener("keydown", handleKeyDown, true);
  }, [onClose, open]);

  return (
    <dialog
      ref={dialogRef}
      aria-labelledby={titleId}
      onCancel={(event) => {
        event.preventDefault();
        onClose();
      }}
      onClose={() => {
        if (!closingFromPropRef.current) {
          restoreFocus();
          onClose();
        }
      }}
      className={
        variant === "drawer"
          ? "border-border bg-background text-foreground backdrop:bg-ink/50 fixed inset-y-0 left-0 m-0 h-dvh max-h-dvh w-[min(88vw,20rem)] max-w-none overflow-y-auto rounded-none border-0 border-r p-0 shadow-xl"
          : "border-border bg-background text-foreground backdrop:bg-ink/50 m-auto w-[min(100%-2rem,36rem)] rounded-2xl border p-0 shadow-xl"
      }
    >
      <div className={variant === "drawer" ? "p-3" : "p-6"}>
        <div className="flex items-center justify-between gap-3">
          <h2 id={titleId} className="font-display text-xl tracking-wide">
            {title}
          </h2>
          {variant === "drawer" ? (
            <IconButton aria-label="Close navigation" onClick={onClose}>
              <X aria-hidden="true" className="size-5" />
            </IconButton>
          ) : null}
        </div>
        <div className="mt-4">{children}</div>
      </div>
    </dialog>
  );
}
