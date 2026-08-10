import type { ReactNode } from "react";

interface FieldErrorProps {
  id?: string;
  children?: ReactNode;
  className?: string;
}

export default function FieldError({
  id,
  children,
  className = "",
}: FieldErrorProps) {
  if (!children) {
    return null;
  }

  return (
    <p id={id} className={`text-critical mt-1.5 text-sm ${className}`}>
      {children}
    </p>
  );
}
