"use client";

import { useEffect, useId, useRef, type ReactNode } from "react";
import { X } from "lucide-react";

import IconButton from "./icon-button";

type DialogVariant = "default" | "sheet";

interface DialogProps {
  open: boolean;
  onClose: () => void;
  title: string;
  children: ReactNode;
  variant?: DialogVariant;
  dismissible?: boolean;
}

export default function Dialog({
  open,
  onClose,
  title,
  children,
  variant = "default",
  dismissible = true,
}: DialogProps) {
  const dialogRef = useRef<HTMLDialogElement>(null);
  const previousFocusRef = useRef<HTMLElement | null>(null);
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

      window.requestAnimationFrame(() => {
        const closeButton = dialog.querySelector<HTMLButtonElement>(
          "[data-dialog-close]",
        );
        const firstControl = dialog.querySelector<HTMLElement>(
          "button:not(:disabled):not([data-dialog-close]), input:not(:disabled), select:not(:disabled), textarea:not(:disabled), [tabindex]:not([tabindex='-1'])",
        );
        (closeButton && !closeButton.disabled
          ? closeButton
          : (firstControl ?? dialog)
        ).focus();
      });
      return;
    }

    if (dialog.open) {
      dialog.close();
    }
  }, [open]);

  useEffect(() => {
    if (!open) {
      return;
    }

    const handleKeyDown = (event: KeyboardEvent) => {
      if (event.key !== "Escape") {
        return;
      }
      const target = event.target;
      event.preventDefault();
      if (target instanceof Element && target.closest("[data-escape-local]")) {
        return;
      }
      if (!dismissible) {
        return;
      }
      onClose();
    };

    window.addEventListener("keydown", handleKeyDown, true);
    return () => window.removeEventListener("keydown", handleKeyDown, true);
  }, [dismissible, onClose, open]);

  return (
    <dialog
      ref={dialogRef}
      tabIndex={-1}
      aria-labelledby={titleId}
      aria-modal="true"
      onCancel={(event) => {
        event.preventDefault();
        if (dismissible) {
          onClose();
        }
      }}
      onClick={(event) => {
        if (dismissible && event.target === event.currentTarget) {
          onClose();
        }
      }}
      onClose={() => {
        restoreFocus();
        onClose();
      }}
      className={
        variant === "sheet"
          ? "border-border bg-background text-foreground backdrop:bg-ink/50 fixed inset-x-0 top-auto bottom-0 m-0 max-h-[calc(100dvh-1rem)] w-full max-w-none overflow-y-auto rounded-t-2xl border p-0 shadow-xl sm:inset-0 sm:m-auto sm:h-fit sm:max-h-[calc(100dvh-2rem)] sm:w-[min(100%-2rem,36rem)] sm:rounded-2xl"
          : "border-border bg-background text-foreground backdrop:bg-ink/50 m-auto w-[min(100%-2rem,36rem)] rounded-2xl border p-0 shadow-xl"
      }
    >
      <div className={variant === "sheet" ? "p-4 sm:p-6" : "p-6"}>
        <div className="flex items-center justify-between gap-3">
          <h2 id={titleId} className="font-display text-xl tracking-wide">
            {title}
          </h2>
          <IconButton
            data-dialog-close
            disabled={!dismissible}
            aria-label={`Close ${title}`}
            onClick={() => {
              if (dismissible) {
                onClose();
              }
            }}
          >
            <X aria-hidden="true" className="size-5" />
          </IconButton>
        </div>
        <div className="mt-4">{children}</div>
      </div>
    </dialog>
  );
}
