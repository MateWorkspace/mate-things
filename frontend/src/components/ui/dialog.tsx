"use client";

import { useEffect, useId, useRef, type ReactNode } from "react";

interface DialogProps {
  open: boolean;
  onClose: () => void;
  title: string;
  children: ReactNode;
}

export default function Dialog({
  open,
  onClose,
  title,
  children,
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
      className="border-border bg-background text-foreground backdrop:bg-ink/50 m-auto w-[min(100%-2rem,36rem)] rounded-2xl border p-0 shadow-xl"
    >
      <div className="p-6">
        <h2 id={titleId} className="font-display text-xl tracking-wide">
          {title}
        </h2>
        <div className="mt-4">{children}</div>
      </div>
    </dialog>
  );
}
