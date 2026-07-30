import type { PageQuery, PageResponse } from "./api/types";

const ALLOWED_LIMITS = [12, 24, 48] as const;

type RawPageQuery = Record<string, string | string[] | undefined>;

function positiveSafeInteger(value: unknown): number | undefined {
  if (typeof value !== "string" || !/^\d+$/.test(value)) {
    return undefined;
  }

  const parsed = Number(value);
  return Number.isSafeInteger(parsed) && parsed > 0 ? parsed : undefined;
}

export function parsePageQuery(
  raw: RawPageQuery,
): Required<Pick<PageQuery, "page" | "limit">> & Pick<PageQuery, "search"> {
  const page = positiveSafeInteger(raw.page) ?? 1;
  const candidate = positiveSafeInteger(raw.limit);
  const limit =
    candidate === undefined
      ? 12
      : ALLOWED_LIMITS.includes(candidate as (typeof ALLOWED_LIMITS)[number])
        ? candidate
        : 24;
  const search = String(raw.search ?? "").trim() || undefined;

  return { page, limit, search };
}

export function getOutOfRangePageRedirect(
  pathname: string,
  raw: RawPageQuery,
  page: PageResponse,
): string | undefined {
  const requestedPage = positiveSafeInteger(raw.page) ?? 1;
  const responseLimit = positiveSafeInteger(String(page.limit));
  const totalItems =
    Number.isSafeInteger(page.total_items) && page.total_items >= 0
      ? page.total_items
      : undefined;

  if (!responseLimit || totalItems === undefined || totalItems === 0) {
    return undefined;
  }

  const totalPages = Math.max(1, Math.ceil(totalItems / responseLimit));
  if (requestedPage <= totalPages) {
    return undefined;
  }

  const params = new URLSearchParams();
  for (const [key, value] of Object.entries(raw)) {
    if (value === undefined || key === "page") {
      continue;
    }

    for (const item of Array.isArray(value) ? value : [value]) {
      params.append(key, item);
    }
  }
  params.set("page", String(totalPages));

  return `${pathname}?${params.toString()}`;
}
