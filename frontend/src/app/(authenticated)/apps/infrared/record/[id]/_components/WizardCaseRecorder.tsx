import LocalDateTime from "@/components/ui/local-date-time";
import StatusBadge from "@/components/ui/status-badge";
import type {
  InfraredStateDeviceRecordCaseResponse,
  InfraredStateResponse,
} from "@/lib/api/infrared";

import { findPreviousCase } from "../_lib/case-state-diff";
import CaseStateTable from "./CaseStateTable";
import RecordCaseList from "./RecordCaseList";
import RecordCaseRawControls from "./RecordCaseRawControls";

function formatDuration(durationUs: number): string {
  return `${Math.round(durationUs / 1000)}ms`;
}

export default function WizardCaseRecorder({
  cases,
  canMutate,
  stateDefinitions,
}: {
  cases: readonly InfraredStateDeviceRecordCaseResponse[];
  canMutate: boolean;
  stateDefinitions: readonly InfraredStateResponse[];
}) {
  const activeCase = cases.find((c) => c.status === "ACTIVE");
  // The active case is rendered prominently above with its own controls -
  // excluding it from the roster below avoids showing its Accept/Discard
  // buttons twice for the same raw.
  const rosterCases = cases.filter((c) => c.status !== "ACTIVE");

  return (
    <div className="space-y-4">
      {activeCase ? (
        <div className="border-border space-y-3 rounded-2xl border p-6">
          <div className="flex flex-wrap items-center justify-between gap-2">
            <div>
              <p className="text-foreground/70 text-xs font-semibold tracking-wide uppercase">
                Press the remote now
              </p>
              <p className="mt-1 text-xl font-semibold">
                Step {activeCase.step}
              </p>
              <CaseStateTable
                states={activeCase.states}
                previousCase={findPreviousCase(cases, activeCase.step)}
                stateDefinitions={stateDefinitions}
              />
            </div>
            <StatusBadge variant="info">Waiting for capture</StatusBadge>
          </div>
          {activeCase.raw.length ? (
            <ul className="space-y-2">
              {activeCase.raw.map((raw) => (
                <li
                  key={raw.id}
                  className="border-border flex flex-wrap items-center justify-between gap-2 rounded-xl border p-3 text-sm"
                >
                  <span>
                    {raw.pulse_count} pulses · {formatDuration(raw.duration_us)}{" "}
                    · <LocalDateTime value={raw.created_at} />
                  </span>
                  <RecordCaseRawControls
                    caseId={activeCase.id}
                    raw={raw}
                    canMutate={canMutate}
                  />
                </li>
              ))}
            </ul>
          ) : (
            <p className="text-foreground/70 text-sm">
              No captures yet — press the button on the remote that matches this
              state.
            </p>
          )}
        </div>
      ) : null}
      <div className="space-y-3">
        <h2 className="font-display text-primary text-xl tracking-wide">
          All cases
        </h2>
        <RecordCaseList
          cases={rosterCases}
          allCases={cases}
          canMutate={canMutate}
          stateDefinitions={stateDefinitions}
        />
      </div>
    </div>
  );
}
