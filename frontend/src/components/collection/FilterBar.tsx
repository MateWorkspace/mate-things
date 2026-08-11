import type { ReactNode } from "react";

interface FilterBarProps {
  clearHref: string;
  applyLabel?: string;
  children: ReactNode;
}

export default function FilterBar({
  clearHref,
  applyLabel = "Apply",
  children,
}: FilterBarProps) {
  return (
    <form className="border-border bg-muted space-y-3 rounded-2xl border p-4">
      <div className="grid gap-3 sm:grid-cols-2 lg:grid-cols-3">{children}</div>
      <div className="border-border flex justify-end gap-2 border-t pt-3">
        <a
          href={clearHref}
          className="border-border text-primary inline-flex min-h-11 items-center rounded-xl border px-4 text-sm font-semibold"
        >
          Clear
        </a>
        <button className="bg-primary text-surface min-h-11 rounded-xl px-6 text-sm font-semibold">
          {applyLabel}
        </button>
      </div>
    </form>
  );
}
