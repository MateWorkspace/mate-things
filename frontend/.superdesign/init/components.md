# Shared UI components

Next.js 16 App Router, React 19, strict TypeScript, and custom Tailwind CSS v4. No external component library is used.

## `src/components/ui/button.tsx`

Rounded shared action primitive with native button props and a `primary | secondary` variant.

```tsx
import type { ButtonHTMLAttributes } from "react";

type ButtonVariant = "primary" | "secondary";

interface ButtonProps extends ButtonHTMLAttributes<HTMLButtonElement> {
  variant?: ButtonVariant;
}

const VARIANT_CLASSES: Record<ButtonVariant, string> = {
  primary:
    "bg-primary text-surface hover:opacity-90 active:opacity-80 disabled:opacity-50",
  secondary:
    "border border-primary/25 text-primary hover:bg-highlight/40 active:bg-highlight/60 disabled:opacity-50",
};

export default function Button({
  variant = "primary",
  className = "",
  ...props
}: ButtonProps) {
  return (
    <button
      className={`focus-visible:ring-primary focus-visible:ring-offset-background inline-flex items-center justify-center rounded-xl px-4 py-2.5 text-sm font-semibold transition-colors focus-visible:ring-2 focus-visible:ring-offset-2 focus-visible:outline-none disabled:cursor-not-allowed ${VARIANT_CLASSES[variant]} ${className}`}
      {...props}
    />
  );
}
```

## `src/components/ui/input.tsx`

Shared native input primitive with branded focus treatment.

```tsx
import type { InputHTMLAttributes } from "react";

export default function Input({
  className = "",
  ...props
}: InputHTMLAttributes<HTMLInputElement>) {
  return (
    <input
      className={`border-ink/15 bg-background text-foreground placeholder:text-foreground/40 focus-visible:border-primary focus-visible:ring-primary w-full rounded-xl border px-3.5 py-2.5 text-sm transition-colors focus-visible:ring-2 focus-visible:outline-none ${className}`}
      {...props}
    />
  );
}
```

## `src/components/ui/label.tsx`

Shared form-label primitive.

```tsx
import type { LabelHTMLAttributes } from "react";

export default function Label({
  className = "",
  ...props
}: LabelHTMLAttributes<HTMLLabelElement>) {
  return (
    <label
      className={`text-foreground/80 mb-1.5 block text-sm font-medium ${className}`}
      {...props}
    />
  );
}
```

## `src/components/ui/toast-provider.tsx`

Global success/error notification context and toast stack.

```tsx
"use client";

import { createContext, useCallback, useRef, useState } from "react";

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
      setTimeout(() => dismiss(id), AUTO_DISMISS_MS);
    },
    [dismiss],
  );

  const value: ToastContextValue = {
    success: (title, message) => push("success", title, message),
    error: (title, message) => push("error", title, message),
  };

  return (
    <ToastContext.Provider value={value}>
      {children}
      <div className="pointer-events-none fixed top-4 right-4 z-50 flex w-full max-w-sm flex-col gap-2">
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
```

## `src/components/ui/toast.tsx`

Animated, dismissible success/error notification card.

```tsx
"use client";

import { CheckCircle2, X, XCircle } from "lucide-react";
import { useEffect, useState } from "react";

export type ToastVariant = "success" | "error";

interface ToastProps {
  variant: ToastVariant;
  title: string;
  message: string;
  leaving: boolean;
  onClose: () => void;
}

const DURATION_MS = 3000;

const VARIANT_STYLES: Record<
  ToastVariant,
  { icon: typeof CheckCircle2; iconClass: string; barClass: string }
> = {
  success: {
    icon: CheckCircle2,
    iconClass: "text-emerald-600",
    barClass: "bg-emerald-600",
  },
  error: {
    icon: XCircle,
    iconClass: "text-red-600",
    barClass: "bg-red-600",
  },
};

export default function Toast({
  variant,
  title,
  message,
  leaving,
  onClose,
}: ToastProps) {
  const [entered, setEntered] = useState(false);
  const [barFilled, setBarFilled] = useState(false);

  useEffect(() => {
    const raf = requestAnimationFrame(() => {
      setEntered(true);
      setBarFilled(true);
    });
    return () => cancelAnimationFrame(raf);
  }, []);

  const { icon: Icon, iconClass, barClass } = VARIANT_STYLES[variant];
  const visible = entered && !leaving;

  return (
    <div
      role="alert"
      data-toast
      className={`border-ink/10 bg-background pointer-events-auto relative w-full max-w-sm overflow-hidden rounded-2xl border shadow-lg transition-all duration-300 ease-out ${
        visible ? "translate-y-0 opacity-100" : "-translate-y-4 opacity-0"
      }`}
    >
      <div className="flex items-start gap-3 p-4 pr-10">
        <Icon
          className={`mt-0.5 h-5 w-5 shrink-0 ${iconClass}`}
          aria-hidden="true"
        />
        <div className="min-w-0 flex-1">
          <p className="font-display text-foreground text-base tracking-wide">
            {title}
          </p>
          <p className="text-foreground/70 mt-0.5 text-sm">{message}</p>
        </div>
      </div>
      <button
        type="button"
        onClick={onClose}
        aria-label="Dismiss notification"
        className="text-foreground/50 hover:bg-ink/10 hover:text-foreground focus-visible:ring-primary absolute top-2 right-2 rounded-full p-1.5 transition-colors focus-visible:ring-2 focus-visible:outline-none"
      >
        <X className="h-4 w-4" aria-hidden="true" />
      </button>
      <div className="bg-ink/10 h-1 w-full">
        <div
          className={`h-full ${barClass} transition-[width] ease-linear`}
          style={{
            width: barFilled ? "100%" : "0%",
            transitionDuration: `${DURATION_MS}ms`,
          }}
        />
      </div>
    </div>
  );
}
```
