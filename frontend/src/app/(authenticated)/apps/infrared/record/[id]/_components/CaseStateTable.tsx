import type {
  InfraredStateDeviceRecordCaseResponse,
  InfraredStateDeviceRecordStateResponse,
  InfraredStateResponse,
} from "@/lib/api/infrared";

import { changedStateIds, sortedStates } from "../_lib/case-state-diff";

export default function CaseStateTable({
  states,
  previousCase,
  stateDefinitions,
}: {
  states: readonly InfraredStateDeviceRecordStateResponse[];
  previousCase: InfraredStateDeviceRecordCaseResponse | undefined;
  /** Device type's state definitions, for name labels - an empty array
   * (e.g. the caller lacks infrared_reference:get) degrades to a generic
   * "State" label rather than breaking. */
  stateDefinitions: readonly InfraredStateResponse[];
}) {
  const changed = changedStateIds(states, previousCase);
  const nameById = new Map(stateDefinitions.map((s) => [s.id, s.name]));

  return (
    <table className="border-border/40 mt-2 w-fit border-collapse text-lg">
      <tbody>
        {sortedStates(states).map((s) => (
          <tr key={s.infrared_state_id}>
            <td className="border-border/40 text-muted-foreground border px-2 py-1 text-right whitespace-nowrap">
              {nameById.get(s.infrared_state_id) ?? "State"}
            </td>
            <td
              className={`border-border/40 border px-2 py-1 text-left whitespace-nowrap ${
                changed.has(s.infrared_state_id) ? "text-primary font-bold" : ""
              }`}
            >
              {s.state_value}
            </td>
          </tr>
        ))}
      </tbody>
    </table>
  );
}
