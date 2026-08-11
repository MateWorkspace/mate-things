import type { InfraredRecordingState } from "@/lib/api/infrared";

import { RECORDING_STATE_LABELS } from "./status";

export interface RecordSessionFilters {
  recordingState?: InfraredRecordingState;
  deviceTypeId?: string;
  start?: string;
  end?: string;
}

export interface ParsedRecordSessionFilters {
  filters: RecordSessionFilters;
  error?: string;
}

type RawRecordSessionParams = Record<string, string | string[] | undefined>;

const VALID_STATES = Object.keys(
  RECORDING_STATE_LABELS,
) as InfraredRecordingState[];

function first(value: string | string[] | undefined): string {
  return (Array.isArray(value) ? value[0] : (value ?? "")).trim();
}

function iso(value: string, field: string): { value?: string; error?: string } {
  if (!value) return {};
  const date = new Date(value);
  if (Number.isNaN(date.getTime())) {
    return { error: `${field} must be a valid date and time.` };
  }
  return { value: date.toISOString() };
}

export function parseRecordSessionFilters(
  raw: RawRecordSessionParams,
): ParsedRecordSessionFilters {
  const start = iso(first(raw.start), "Start");
  const end = iso(first(raw.end), "End");

  let error = start.error ?? end.error;
  if (
    !error &&
    start.value &&
    end.value &&
    new Date(start.value).getTime() > new Date(end.value).getTime()
  ) {
    error = "End must be after start.";
  }

  const stateRaw = first(raw.recording_state).toUpperCase();
  const recordingState = VALID_STATES.includes(
    stateRaw as InfraredRecordingState,
  )
    ? (stateRaw as InfraredRecordingState)
    : undefined;
  if (!error && stateRaw && !recordingState) {
    error = "Select a valid status.";
  }

  return {
    filters: {
      start: start.value,
      end: end.value,
      recordingState,
      deviceTypeId: first(raw.infrared_device_type_id) || undefined,
    },
    error,
  };
}
