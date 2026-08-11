"use client";

import { useActionState } from "react";

import ActionMessage from "@/components/forms/ActionMessage";
import Button from "@/components/ui/button";
import StatusBadge from "@/components/ui/status-badge";
import { useRefreshAfterAction } from "@/hooks/use-refresh-after-action";
import type { InfraredTestCaseResponse } from "@/lib/api/infrared";

import {
  recordTestCaseResultAction,
  transmitTestCaseAction,
} from "../_lib/actions";
import { EMPTY_RECORD_SESSION_ACTION_STATE } from "../_lib/state";

const TEST_STATUS_VARIANT = {
  PENDING: "neutral",
  PASSED: "success",
  FAILED: "critical",
} as const;

export default function RecordTestCaseList({
  testCases,
  canMutate,
}: {
  testCases: readonly InfraredTestCaseResponse[];
  canMutate: boolean;
}) {
  const [transmitState, transmitAction, transmitPending] = useActionState(
    transmitTestCaseAction,
    EMPTY_RECORD_SESSION_ACTION_STATE,
  );
  const [resultState, resultAction, resultPending] = useActionState(
    recordTestCaseResultAction,
    EMPTY_RECORD_SESSION_ACTION_STATE,
  );
  useRefreshAfterAction(resultState);

  return (
    <ul className="space-y-3">
      {testCases.map((tc) => (
        <li key={tc.id} className="border-border rounded-2xl border p-4">
          <div className="flex flex-wrap items-center justify-between gap-2">
            <div>
              <p className="font-semibold">{tc.description}</p>
              <p className="text-muted-foreground text-xs">
                {tc.states.map((s) => s.state_value).join(", ")}
              </p>
            </div>
            <div className="flex items-center gap-2">
              <StatusBadge variant={TEST_STATUS_VARIANT[tc.status]}>
                {tc.status}
              </StatusBadge>
              {canMutate && tc.status === "PENDING" ? (
                <>
                  <form action={transmitAction}>
                    <input type="hidden" name="test_case_id" value={tc.id} />
                    <Button type="submit" disabled={transmitPending}>
                      {transmitPending ? "Transmitting…" : "Transmit"}
                    </Button>
                  </form>
                  <form action={resultAction}>
                    <input type="hidden" name="test_case_id" value={tc.id} />
                    <input type="hidden" name="passed" value="true" />
                    <Button type="submit" disabled={resultPending}>
                      Passed
                    </Button>
                  </form>
                  <form action={resultAction}>
                    <input type="hidden" name="test_case_id" value={tc.id} />
                    <input type="hidden" name="passed" value="false" />
                    <Button
                      type="submit"
                      variant="secondary"
                      disabled={resultPending}
                    >
                      Failed
                    </Button>
                  </form>
                </>
              ) : null}
            </div>
          </div>
        </li>
      ))}
      <ActionMessage state={transmitState} />
      <ActionMessage state={resultState} />
    </ul>
  );
}
