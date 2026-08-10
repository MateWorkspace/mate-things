import type { PageQuery, PageResponse } from "./api/types";
import {
  buildOutOfRangeRedirect,
  firstQueryValue,
  parsePositiveSafeInteger,
  type RawSearchParams,
} from "./query";

const ALLOWED_LIMITS = [12, 24, 48] as const;

type RawPageQuery = RawSearchParams;

export function parsePageQuery(
  raw: RawPageQuery,
): Required<Pick<PageQuery, "page" | "limit">> & Pick<PageQuery, "search"> {
  const page = parsePositiveSafeInteger(firstQueryValue(raw.page), 1);
  const candidate = parsePositiveSafeInteger(firstQueryValue(raw.limit), 0);
  const limit =
    candidate === 0
      ? 12
      : ALLOWED_LIMITS.includes(candidate as (typeof ALLOWED_LIMITS)[number])
        ? candidate
        : 24;
  const search = firstQueryValue(raw.search)?.trim() || undefined;

  return { page, limit, search };
}

export function getOutOfRangePageRedirect(
  pathname: string,
  raw: RawPageQuery,
  page: PageResponse,
): string | undefined {
  const requestedPage = parsePositiveSafeInteger(firstQueryValue(raw.page), 1);
  const responseLimit = parsePositiveSafeInteger(page.limit, 0);
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

  return buildOutOfRangeRedirect(pathname, raw, totalPages);
}
