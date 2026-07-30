import Link from "next/link";

import type { PageResponse } from "@/lib/api/types";

type CollectionSearchParams = Record<string, string | string[] | undefined>;

interface PaginationProps {
  page: PageResponse;
  searchParams: CollectionSearchParams;
  pathname?: string;
}

function pageHref(
  pathname: string,
  searchParams: CollectionSearchParams,
  page: number,
): string {
  const params = new URLSearchParams();

  for (const [key, value] of Object.entries(searchParams)) {
    if (value === undefined || key === "page") {
      continue;
    }

    for (const item of Array.isArray(value) ? value : [value]) {
      params.append(key, item);
    }
  }

  params.set("page", String(page));
  return `${pathname}?${params.toString()}`;
}

function PaginationControl({
  children,
  disabled,
  href,
}: {
  children: string;
  disabled: boolean;
  href: string;
}) {
  const className =
    "focus-visible:ring-focus rounded-xl px-4 py-2.5 text-sm font-semibold focus-visible:ring-2 focus-visible:ring-offset-2 focus-visible:ring-offset-background focus-visible:outline-none";

  if (disabled) {
    return (
      <span
        aria-disabled="true"
        className={`${className} cursor-not-allowed opacity-50`}
      >
        {children}
      </span>
    );
  }

  return (
    <Link
      className={`text-primary hover:bg-highlight/40 ${className}`}
      href={href}
    >
      {children}
    </Link>
  );
}

export default function Pagination({
  page,
  pathname = "",
  searchParams,
}: PaginationProps) {
  const totalPages = Math.max(1, Math.ceil(page.total_items / page.limit));
  const currentPage = Math.min(Math.max(1, page.page), totalPages);
  const previousDisabled = currentPage === 1;
  const nextDisabled = currentPage === totalPages;

  return (
    <nav
      aria-label="Pagination"
      className="flex items-center justify-between gap-3"
    >
      <PaginationControl
        disabled={previousDisabled}
        href={pageHref(pathname, searchParams, currentPage - 1)}
      >
        Previous page
      </PaginationControl>
      <p className="text-muted-foreground text-sm" aria-live="polite">
        Page {currentPage} of {totalPages}
      </p>
      <PaginationControl
        disabled={nextDisabled}
        href={pageHref(pathname, searchParams, currentPage + 1)}
      >
        Next page
      </PaginationControl>
    </nav>
  );
}
