const DEFAULT_MAX_PAGES = 1_000;

export async function collectAllPages<T>(options: {
  fetchPage: (
    page: number,
  ) => Promise<{ data: T[]; page: number; limit: number; total: number }>;
  keyOf: (item: T) => string;
  maxPages?: number;
}): Promise<T[]> {
  const collected: T[] = [];
  const seen = new Set<string>();
  const maxPages =
    options.maxPages !== undefined &&
    Number.isSafeInteger(options.maxPages) &&
    options.maxPages > 0
      ? options.maxPages
      : DEFAULT_MAX_PAGES;
  let requestedPage = 1;
  let previousResponsePage = 0;

  for (let fetched = 0; fetched < maxPages; fetched += 1) {
    const response = await options.fetchPage(requestedPage);
    let added = 0;

    for (const item of response.data) {
      const key = options.keyOf(item);
      if (!seen.has(key)) {
        seen.add(key);
        collected.push(item);
        added += 1;
      }
    }

    if (added === 0) {
      break;
    }

    if (
      !Number.isSafeInteger(response.page) ||
      response.page <= previousResponsePage
    ) {
      break;
    }
    previousResponsePage = response.page;

    if (
      Number.isSafeInteger(response.total) &&
      response.total >= 0 &&
      collected.length >= response.total
    ) {
      break;
    }

    if (!Number.isSafeInteger(response.limit) || response.limit <= 0) {
      break;
    }

    const totalPages = Math.ceil(response.total / response.limit);
    if (!Number.isSafeInteger(totalPages) || response.page >= totalPages) {
      break;
    }

    requestedPage = response.page + 1;
  }

  return collected;
}
