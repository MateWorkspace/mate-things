"use client";

import { createContext, useCallback, useEffect, useRef, useState } from "react";

import Toast, { type ToastVariant } from "@/components/ui/toast";

interface ToastItem {
  id: number;
  variant: ToastVariant;
  title: string;
  message: string;
  leaving: boolean;
}

export interface ToastContextValue {
  success: (title: string, message: string) => void;
  error: (title: string, message: string) => void;
}

export const ToastContext = createContext<ToastContextValue | null>(null);

const AUTO_DISMISS_MS = 3000;
const EXIT_ANIMATION_MS = 300;

export default function ToastProvider({
  children,
}: {
  children: React.ReactNode;
}) {
  const [toasts, setToasts] = useState<ToastItem[]>([]);
  const nextId = useRef(0);
  const containerRef = useRef<HTMLDivElement>(null);

  // Dialogs in this app use the native <dialog>.showModal(), which the
  // browser renders in the top layer - above every z-indexed element
  // regardless of stacking context. A plain fixed div can never paint over
  // that, so the toast container has to be promoted into the top layer too,
  // via the Popover API. A popover only moves to the top of the top-layer
  // stack when it's (re-)shown, so it has to be re-shown on every toast -
  // showing it once on mount would leave it permanently below any dialog
  // opened afterward.
  const bringToFront = useCallback(() => {
    const container = containerRef.current;
    if (!container) return;
    if (container.matches(":popover-open")) {
      container.hidePopover();
    }
    container.showPopover();
  }, []);

  useEffect(() => {
    bringToFront();
  }, [bringToFront]);

  const dismiss = useCallback((id: number) => {
    setToasts((current) =>
      current.map((toast) =>
        toast.id === id ? { ...toast, leaving: true } : toast,
      ),
    );
    setTimeout(() => {
      setToasts((current) => current.filter((toast) => toast.id !== id));
    }, EXIT_ANIMATION_MS);
  }, []);

  const push = useCallback(
    (variant: ToastVariant, title: string, message: string) => {
      const id = nextId.current++;
      setToasts((current) => [
        ...current,
        { id, variant, title, message, leaving: false },
      ]);
      bringToFront();
      setTimeout(() => dismiss(id), AUTO_DISMISS_MS);
    },
    [bringToFront, dismiss],
  );

  const value: ToastContextValue = {
    success: (title, message) => push("success", title, message),
    error: (title, message) => push("error", title, message),
  };

  return (
    <ToastContext.Provider value={value}>
      {children}
      <div
        ref={containerRef}
        popover="manual"
        className="pointer-events-none fixed top-4 right-4 bottom-auto left-auto m-0 flex w-full max-w-sm flex-col gap-2 border-0 bg-transparent p-0 overflow-visible"
      >
        {toasts.map((toast) => (
          <Toast
            key={toast.id}
            variant={toast.variant}
            title={toast.title}
            message={toast.message}
            leaving={toast.leaving}
            onClose={() => dismiss(toast.id)}
          />
        ))}
      </div>
    </ToastContext.Provider>
  );
}
