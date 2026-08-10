export type RawSearchParams = Record<string, string | string[] | undefined>;

export function firstQueryValue(
  value: string | string[] | undefined,
): string | undefined {
  return Array.isArray(value) ? value[0] : value;
}

export function parsePositiveSafeInteger(
  value: unknown,
  fallback: number,
): number {
  if (
    (typeof value !== "number" && typeof value !== "string") ||
    (typeof value === "string" && !/^\d+$/.test(value))
  ) {
    return fallback;
  }

  const parsed = Number(value);
  return Number.isSafeInteger(parsed) && parsed > 0 ? parsed : fallback;
}

export function parseBooleanQuery(value: unknown): boolean | undefined {
  return value === "true" ? true : value === "false" ? false : undefined;
}

export function parseEnumQuery<const T extends readonly string[]>(
  value: unknown,
  values: T,
): T[number] | undefined {
  return typeof value === "string" && values.includes(value)
    ? (value as T[number])
    : undefined;
}

export function parseLocalDateTime(value: unknown): Date | undefined {
  if (typeof value !== "string") {
    return undefined;
  }

  const match =
    /^(\d{4})-(\d{2})-(\d{2})T(\d{2}):(\d{2})(?::(\d{2})(?:\.(\d{1,3}))?)?$/.exec(
      value,
    );
  if (!match) {
    return undefined;
  }

  const [, yearValue, monthValue, dayValue, hourValue, minuteValue] = match;
  const secondValue = match[6] ?? "0";
  const millisecondValue = (match[7] ?? "0").padEnd(3, "0");
  const year = Number(yearValue);
  const month = Number(monthValue);
  const day = Number(dayValue);
  const hour = Number(hourValue);
  const minute = Number(minuteValue);
  const second = Number(secondValue);
  const millisecond = Number(millisecondValue);
  const parsed = new Date(0);
  parsed.setFullYear(year, month - 1, day);
  parsed.setHours(hour, minute, second, millisecond);

  return parsed.getFullYear() === year &&
    parsed.getMonth() === month - 1 &&
    parsed.getDate() === day &&
    parsed.getHours() === hour &&
    parsed.getMinutes() === minute &&
    parsed.getSeconds() === second &&
    parsed.getMilliseconds() === millisecond
    ? parsed
    : undefined;
}

export function toUtcQueryValue(value: Date): string {
  return value.toISOString();
}

const absoluteDateTimePattern =
  /^\d{4}-\d{2}-\d{2}T\d{2}:\d{2}(?::\d{2}(?:\.\d{1,3})?)?(?:Z|[+-]\d{2}:\d{2})$/;

/**
 * Parses an absolute ISO-8601 timestamp (one carrying an explicit `Z` or
 * `+HH:MM`/`-HH:MM` offset) - the shape `toUtcQueryValue` produces and that
 * generated links (e.g. the dashboard's "last window" cards) pass as
 * `start`/`end` query values. Deliberately disjoint from
 * `parseLocalDateTime`, which rejects any offset because it treats its
 * input as browser-local wall-clock time from a `datetime-local` field.
 */
export function parseAbsoluteDateTime(value: unknown): Date | undefined {
  if (typeof value !== "string" || !absoluteDateTimePattern.test(value)) {
    return undefined;
  }

  const parsed = new Date(value);
  return Number.isNaN(parsed.getTime()) ? undefined : parsed;
}

export function buildCollectionUrl(
  pathname: string,
  values: Readonly<Record<string, string | undefined>>,
): string {
  const params = new URLSearchParams();
  for (const [key, value] of Object.entries(values)) {
    if (value !== undefined) {
      params.append(key, value);
    }
  }

  const query = params.toString();
  return query ? `${pathname}?${query}` : pathname;
}

export function buildOutOfRangeRedirect(
  pathname: string,
  values: RawSearchParams,
  lastPage: number,
): string {
  const params = new URLSearchParams();
  for (const [key, value] of Object.entries(values)) {
    if (value === undefined || key === "page") {
      continue;
    }

    for (const item of Array.isArray(value) ? value : [value]) {
      params.append(key, item);
    }
  }
  params.set("page", String(lastPage));

  return `${pathname}?${params.toString()}`;
}
