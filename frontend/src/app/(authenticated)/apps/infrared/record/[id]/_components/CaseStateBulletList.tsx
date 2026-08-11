import type {
  InfraredStateDeviceRecordCaseResponse,
  InfraredStateDeviceRecordStateResponse,
} from "@/lib/api/infrared";

import { changedStateIds, sortedStates } from "../_lib/case-state-diff";

export default function CaseStateBulletList({
  states,
  previousCase,
}: {
  states: readonly InfraredStateDeviceRecordStateResponse[];
  previousCase: InfraredStateDeviceRecordCaseResponse | undefined;
}) {
  const changed = changedStateIds(states, previousCase);

  return (
    <ul className="mt-1 list-inside list-disc space-y-0.5 text-xs">
      {sortedStates(states).map((s) => (
        <li
          key={s.infrared_state_id}
          className={
            changed.has(s.infrared_state_id)
              ? "text-primary font-semibold"
              : "text-muted-foreground"
          }
        >
          {s.state_value}
        </li>
      ))}
    </ul>
  );
}
