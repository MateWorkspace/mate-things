import type {
  InfraredStateDeviceRecordCaseResponse,
  InfraredStateDeviceRecordStateResponse,
} from "@/lib/api/infrared";

/**
 * The case immediately before `step` in recording order, searched within
 * `allCases` rather than whatever subset is actually being rendered - a
 * filtered list (e.g. the roster with the active case excluded) would
 * otherwise skip a step and produce a wrong diff.
 */
export function findPreviousCase(
  allCases: readonly InfraredStateDeviceRecordCaseResponse[],
  step: number,
): InfraredStateDeviceRecordCaseResponse | undefined {
  let previous: InfraredStateDeviceRecordCaseResponse | undefined;
  for (const c of allCases) {
    if (c.step < step && (!previous || c.step > previous.step)) {
      previous = c;
    }
  }
  return previous;
}

/**
 * infrared_state_id set of every state whose value differs from the
 * previous case - matched by state id, not array position, since the
 * backend doesn't guarantee `states` is ordered the same way case to case.
 * No previous case (the first case in the plan) means nothing is
 * highlighted - there's nothing to have changed from yet.
 */
export function changedStateIds(
  states: readonly InfraredStateDeviceRecordStateResponse[],
  previousCase: InfraredStateDeviceRecordCaseResponse | undefined,
): ReadonlySet<string> {
  if (!previousCase) {
    return new Set();
  }
  const previousValueByStateId = new Map(
    previousCase.states.map((s) => [s.infrared_state_id, s.state_value]),
  );
  const changed = new Set<string>();
  for (const s of states) {
    if (previousValueByStateId.get(s.infrared_state_id) !== s.state_value) {
      changed.add(s.infrared_state_id);
    }
  }
  return changed;
}

/** Stable, case-to-case-consistent order so the same state always renders
 * on the same line - without this, highlighting "what changed" would be
 * much harder to scan even though the values would still be correct. */
export function sortedStates(
  states: readonly InfraredStateDeviceRecordStateResponse[],
): InfraredStateDeviceRecordStateResponse[] {
  return [...states].sort((a, b) =>
    a.infrared_state_id.localeCompare(b.infrared_state_id),
  );
}
