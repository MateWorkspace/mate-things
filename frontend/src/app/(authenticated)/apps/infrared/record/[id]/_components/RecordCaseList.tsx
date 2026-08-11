"use client";

import { useActionState } from "react";

import ActionMessage from "@/components/forms/ActionMessage";
import Button from "@/components/ui/button";
import StatusBadge from "@/components/ui/status-badge";
import { useRefreshAfterAction } from "@/hooks/use-refresh-after-action";
import type {
  InfraredStateDeviceRecordCaseResponse,
  InfraredStateResponse,
} from "@/lib/api/infrared";

import { retryCaseAction } from "../_lib/actions";
import { findPreviousCase } from "../_lib/case-state-diff";
import { EMPTY_RECORD_SESSION_ACTION_STATE } from "../_lib/state";
import CaseStateTable from "./CaseStateTable";
import RecordCaseRawControls from "./RecordCaseRawControls";

const CASE_STATUS_VARIANT = {
  PENDING: "neutral",
  ACTIVE: "info",
  ACCEPTED: "success",
} as const;

function summarizeRaw(rawCount: number): string {
  return rawCount === 1 ? "1 capture" : `${rawCount} captures`;
}

export default function RecordCaseList({
  cases,
  canMutate,
  stateDefinitions,
  allCases = cases,
}: {
  cases: readonly InfraredStateDeviceRecordCaseResponse[];
  canMutate: boolean;
  stateDefinitions: readonly InfraredStateResponse[];
  /** Full recording-order case list, for looking up "the previous case" -
   * only needed when `cases` is a filtered subset (e.g. the active case
   * excluded); defaults to `cases` itself otherwise. */
  allCases?: readonly InfraredStateDeviceRecordCaseResponse[];
}) {
  const [retryState, retryAction, retryPending] = useActionState(
    retryCaseAction,
    EMPTY_RECORD_SESSION_ACTION_STATE,
  );
  useRefreshAfterAction(retryState);

  return (
    <ul className="space-y-3">
      {cases.map((c) => (
        <li
          key={c.id}
          className="border-border space-y-3 rounded-2xl border p-4"
        >
          <div className="flex flex-wrap items-center justify-between gap-2">
            <div>
              <p className="text-xl font-semibold">Step {c.step}</p>
              <p className="text-muted-foreground text-xs">
                {summarizeRaw(c.raw.length)}
              </p>
              <CaseStateTable
                states={c.states}
                previousCase={findPreviousCase(allCases, c.step)}
                stateDefinitions={stateDefinitions}
              />
            </div>
            <div className="flex items-center gap-2">
              <StatusBadge variant={CASE_STATUS_VARIANT[c.status]}>
                {c.status}
              </StatusBadge>
              {canMutate && c.status === "ACCEPTED" ? (
                <form action={retryAction}>
                  <input type="hidden" name="case_id" value={c.id} />
                  <Button
                    type="submit"
                    variant="secondary"
                    disabled={retryPending}
                  >
                    {retryPending ? "Resetting…" : "Retry"}
                  </Button>
                </form>
              ) : null}
            </div>
          </div>
          {c.raw.map((raw) => (
            <RecordCaseRawControls
              key={raw.id}
              caseId={c.id}
              raw={raw}
              canMutate={canMutate}
            />
          ))}
        </li>
      ))}
      <ActionMessage state={retryState} />
    </ul>
  );
}
