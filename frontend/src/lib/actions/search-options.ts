import type { PageDataResponse } from "@/lib/api/types";

export type SearchOption = {
  value: string;
  label: string;
  description?: string;
};

export type SearchOptionsRequest = {
  query: string;
  page: number;
  limit?: number;
};

export type SearchOptionsPage = {
  items: SearchOption[];
  page: number;
  totalPages: number;
};

const MAX_SEARCH_OPTIONS_LIMIT = 100;

function isRecord(input: unknown): input is Record<string, unknown> {
  return typeof input === "object" && input !== null && !Array.isArray(input);
}

function isPositiveSafeInteger(input: unknown): input is number {
  return Number.isSafeInteger(input) && Number(input) > 0;
}

export function parseSearchOptionsRequest(
  input: unknown,
): SearchOptionsRequest {
  if (
    !isRecord(input) ||
    typeof input.query !== "string" ||
    !isPositiveSafeInteger(input.page) ||
    (input.limit !== undefined &&
      (!isPositiveSafeInteger(input.limit) ||
        input.limit > MAX_SEARCH_OPTIONS_LIMIT))
  ) {
    throw new Error("Invalid search options request");
  }

  return {
    query: input.query.trim(),
    page: input.page,
    ...(input.limit === undefined ? {} : { limit: input.limit }),
  };
}

export function toSearchOptionsPage<T>(
  response: PageDataResponse<T>,
  toOption: (item: T) => SearchOption,
): SearchOptionsPage {
  return {
    items: response.data.map(toOption),
    page: response.page.page,
    totalPages: Math.max(
      1,
      Math.ceil(response.page.total_items / response.page.limit),
    ),
  };
}
