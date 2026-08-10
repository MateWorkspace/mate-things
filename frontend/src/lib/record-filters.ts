import {
  firstQueryValue,
  parseAbsoluteDateTime,
  parseLocalDateTime,
  parsePositiveSafeInteger,
  toUtcQueryValue,
  type RawSearchParams,
} from "./query";

export type RawRecordFilters = RawSearchParams;

export interface RecordFilters {
  start?: string;
  end?: string;
  nodeId?: string;
  nodeDeviceId?: string;
  actionId?: string;
  executionId?: string;
  metricName?: string;
  payloadSchemaName?: string;
  payloadSchemaVersion?: number;
  tag?: string;
}

export interface ParsedRecordFilters {
  filters: RecordFilters;
  error?: string;
  fieldErrors?: Record<string, string>;
}

function first(value: string | string[] | undefined): string {
  return (firstQueryValue(value) ?? "").trim();
}

function iso(value: string, field: string): { value?: string; error?: string } {
  if (!value) return {};
  const date = parseLocalDateTime(value) ?? parseAbsoluteDateTime(value);
  if (!date) {
    return { error: `${field} must be a valid date and time.` };
  }
  return { value: toUtcQueryValue(date) };
}

export function parseRecordFilters(raw: RawRecordFilters): ParsedRecordFilters {
  const start = iso(first(raw.start), "Start");
  const end = iso(first(raw.end), "End");
  const fieldErrors: Record<string, string> = {};
  if (start.error) fieldErrors.start = start.error;
  if (end.error) fieldErrors.end = end.error;

  const versionRaw = first(raw.payload_schema_version);
  let payloadSchemaVersion: number | undefined;
  if (versionRaw) {
    const version = parsePositiveSafeInteger(versionRaw, 0);
    if (version === 0) {
      fieldErrors.payload_schema_version =
        "Schema version must be a positive whole number.";
    } else {
      payloadSchemaVersion = version;
    }
  }

  if (
    start.value &&
    end.value &&
    new Date(start.value).getTime() > new Date(end.value).getTime()
  ) {
    fieldErrors.end = "End must be after start.";
  }

  const error = Object.values(fieldErrors)[0];
  return {
    filters: {
      start: start.value,
      end: end.value,
      nodeId: first(raw.node_id) || undefined,
      nodeDeviceId: first(raw.node_device_id) || undefined,
      actionId: first(raw.action_id) || undefined,
      executionId: first(raw.execution_id) || undefined,
      metricName: first(raw.metric_name) || undefined,
      payloadSchemaName: first(raw.payload_schema_name) || undefined,
      payloadSchemaVersion,
      tag: first(raw.tag) || undefined,
    },
    error,
    fieldErrors: error ? fieldErrors : undefined,
  };
}

export function toDatetimeLocal(value?: string): string {
  if (!value) return "";
  const date = new Date(value);
  if (Number.isNaN(date.getTime())) return "";
  const local = new Date(date.getTime() - date.getTimezoneOffset() * 60_000);
  return local.toISOString().slice(0, 16);
}

export function activeFilterEntries(
  filters: Record<string, string | number | undefined>,
): [string, string][] {
  return Object.entries(filters).flatMap(([key, value]) =>
    value === undefined || value === "" ? [] : [[key, String(value)]],
  );
}
