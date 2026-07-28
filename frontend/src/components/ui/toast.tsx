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

const DURATION_MS = 5000;

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
  // Starts off-screen/invisible, then flips true on the next frame so the
  // transition to the "settled" position actually animates (slide down).
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
      className={`pointer-events-auto relative w-full max-w-sm overflow-hidden rounded-2xl border border-ink/10 bg-background shadow-lg transition-all duration-300 ease-out ${
        visible ? "translate-y-0 opacity-100" : "-translate-y-4 opacity-0"
      }`}
    >
      <div className="flex items-start gap-3 p-4 pr-10">
        <Icon className={`mt-0.5 h-5 w-5 shrink-0 ${iconClass}`} aria-hidden="true" />
        <div className="min-w-0 flex-1">
          <p className="font-display text-base tracking-wide text-foreground">
            {title}
          </p>
          <p className="mt-0.5 text-sm text-foreground/70">{message}</p>
        </div>
      </div>
      <button
        type="button"
        onClick={onClose}
        aria-label="Dismiss notification"
        className="absolute right-2 top-2 rounded-full p-1.5 text-foreground/50 transition-colors hover:bg-ink/10 hover:text-foreground focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-primary"
      >
        <X className="h-4 w-4" aria-hidden="true" />
      </button>
      <div className="h-1 w-full bg-ink/10">
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
