import FilterBar from "@/components/collection/FilterBar";
import TimeRangeFilter from "@/components/records/TimeRangeFilter";
import Select from "@/components/ui/select";
import type {
  InfraredDeviceTypeResponse,
  InfraredRecordingState,
} from "@/lib/api/infrared";

import { RECORDING_STATE_LABELS } from "../_lib/status";

const STATE_KEYS = Object.keys(
  RECORDING_STATE_LABELS,
) as InfraredRecordingState[];

export default function RecordSessionFilters({
  deviceTypes,
  start,
  end,
  recordingState,
  deviceTypeId,
}: {
  deviceTypes: readonly InfraredDeviceTypeResponse[];
  start?: string;
  end?: string;
  recordingState?: InfraredRecordingState;
  deviceTypeId?: string;
}) {
  return (
    <FilterBar clearHref="/apps/infrared/record">
      <TimeRangeFilter start={start} end={end} />
      <div>
        <label
          className="text-foreground/70 mb-1.5 block text-xs font-semibold"
          htmlFor="record-session-status"
        >
          Status
        </label>
        <Select
          id="record-session-status"
          name="recording_state"
          defaultValue={recordingState ?? ""}
        >
          <option value="">Any status</option>
          {STATE_KEYS.map((state) => (
            <option key={state} value={state}>
              {RECORDING_STATE_LABELS[state].label}
            </option>
          ))}
        </Select>
      </div>
      <div>
        <label
          className="text-foreground/70 mb-1.5 block text-xs font-semibold"
          htmlFor="record-session-device-type"
        >
          Device type
        </label>
        <Select
          id="record-session-device-type"
          name="infrared_device_type_id"
          defaultValue={deviceTypeId ?? ""}
        >
          <option value="">Any device type</option>
          {deviceTypes.map((dt) => (
            <option key={dt.id} value={dt.id}>
              {dt.name}
            </option>
          ))}
        </Select>
      </div>
    </FilterBar>
  );
}
