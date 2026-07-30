import type { PageQuery } from "./api/types";

const ALLOWED_LIMITS = [12, 24, 48] as const;

type RawPageQuery = Record<string, string | string[] | undefined>;

export function parsePageQuery(
  raw: RawPageQuery,
): Required<Pick<PageQuery, "page" | "limit">> & Pick<PageQuery, "search"> {
  const page = Math.max(1, Number(raw.page) || 1);
  const candidate = Number(raw.limit) || 12;
  const limit = ALLOWED_LIMITS.includes(
    candidate as (typeof ALLOWED_LIMITS)[number],
  )
    ? candidate
    : 24;
  const search = String(raw.search ?? "").trim() || undefined;

  return { page, limit, search };
}
