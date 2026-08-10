export type ParseResult<T> =
  | { ok: true; raw: string; value: T }
  | { ok: false; raw: string; error: string };

interface FormTextOptions {
  trim?: boolean;
}

export function formText(
  formData: FormData,
  name: string,
  { trim = true }: FormTextOptions = {},
): string {
  const entry = formData.get(name);
  if (typeof entry !== "string") {
    return "";
  }
  return trim ? entry.trim() : entry;
}

export function optionalString(
  formData: FormData,
  name: string,
): string | undefined {
  return formText(formData, name) || undefined;
}

export function requiredString(
  formData: FormData,
  name: string,
  error = "This field is required.",
): ParseResult<string> {
  const raw = formText(formData, name, { trim: false });
  const value = raw.trim();

  return value ? { ok: true, raw, value } : { ok: false, raw, error };
}

export function parseJsonObject(
  raw: string,
): ParseResult<Record<string, unknown>> {
  try {
    const value: unknown = JSON.parse(raw);
    if (!value || typeof value !== "object" || Array.isArray(value)) {
      return { ok: false, raw, error: "Enter a JSON object." };
    }
    return { ok: true, raw, value: value as Record<string, unknown> };
  } catch {
    return { ok: false, raw, error: "Enter valid JSON." };
  }
}

export function optionalDateTime(
  formData: FormData,
  name: string,
): ParseResult<string | undefined> {
  const raw = formText(formData, name, { trim: false });
  const value = raw.trim();
  if (!value) {
    return { ok: true, raw, value: undefined };
  }

  const parsed = new Date(value);
  if (Number.isNaN(parsed.getTime())) {
    return { ok: false, raw, error: "Enter a valid date and time." };
  }
  return { ok: true, raw, value: parsed.toISOString() };
}
