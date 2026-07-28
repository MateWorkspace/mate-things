import type { LabelHTMLAttributes } from "react";

export default function Label({
  className = "",
  ...props
}: LabelHTMLAttributes<HTMLLabelElement>) {
  return (
    <label
      className={`mb-1.5 block text-sm font-medium text-foreground/80 ${className}`}
      {...props}
    />
  );
}
