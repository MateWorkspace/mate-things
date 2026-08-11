"use client";

import { useActionState } from "react";

import ActionMessage from "@/components/forms/ActionMessage";
import Button from "@/components/ui/button";
import { useRefreshAfterAction } from "@/hooks/use-refresh-after-action";

import {
  recordTestCaseResultAction,
  transmitTestCaseAction,
} from "../_lib/actions";
import { EMPTY_RECORD_SESSION_ACTION_STATE } from "../_lib/state";

export default function RecordTestCaseControls({
  testCaseId,
}: {
  testCaseId: string;
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
    <div className="flex items-center gap-2">
      <form action={transmitAction}>
        <input type="hidden" name="test_case_id" value={testCaseId} />
        <Button type="submit" disabled={transmitPending}>
          {transmitPending ? "Transmitting…" : "Transmit"}
        </Button>
      </form>
      <form action={resultAction}>
        <input type="hidden" name="test_case_id" value={testCaseId} />
        <input type="hidden" name="passed" value="true" />
        <Button type="submit" disabled={resultPending}>
          Passed
        </Button>
      </form>
      <form action={resultAction}>
        <input type="hidden" name="test_case_id" value={testCaseId} />
        <input type="hidden" name="passed" value="false" />
        <Button type="submit" variant="secondary" disabled={resultPending}>
          Failed
        </Button>
      </form>
      <ActionMessage state={transmitState} />
      <ActionMessage state={resultState} />
    </div>
  );
}
