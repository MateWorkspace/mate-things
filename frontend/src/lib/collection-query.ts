import type { PageQuery } from "./api/types";

const ALLOWED_LIMITS = [12, 24, 48] as const;

type RawPageQuery = Record<string, string | string[] | undefined>;

function firstValue(value: string | string[] | undefined): string | undefined {
  return Array.isArray(value) ? value[0] : value;
}

export function parsePageQuery(
  raw: RawPageQuery,
): Required<Pick<PageQuery, "page" | "limit">> & Pick<PageQuery, "search"> {
  const page = Math.max(1, Number(firstValue(raw.page)) || 1);
  const candidate = Number(firstValue(raw.limit)) || 12;
  const limit = ALLOWED_LIMITS.includes(
    candidate as (typeof ALLOWED_LIMITS)[number],
  )
    ? candidate
    : 24;
  const search = firstValue(raw.search)?.trim() || undefined;

  return { page, limit, search };
}
