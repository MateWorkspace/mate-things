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
