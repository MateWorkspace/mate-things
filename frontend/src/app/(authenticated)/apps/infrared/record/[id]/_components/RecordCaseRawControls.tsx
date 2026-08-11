"use client";

import { useActionState } from "react";

import ActionMessage from "@/components/forms/ActionMessage";
import Button from "@/components/ui/button";
import { useRefreshAfterAction } from "@/hooks/use-refresh-after-action";
import type { InfraredStateDeviceRecordRawResponse } from "@/lib/api/infrared";

import { acceptRawAction, discardRawAction } from "../_lib/actions";
import { EMPTY_RECORD_SESSION_ACTION_STATE } from "../_lib/state";

export default function RecordCaseRawControls({
  caseId,
  raw,
  canMutate,
}: {
  caseId: string;
  raw: InfraredStateDeviceRecordRawResponse;
  canMutate: boolean;
}) {
  const [acceptState, acceptAction, acceptPending] = useActionState(
    acceptRawAction,
    EMPTY_RECORD_SESSION_ACTION_STATE,
  );
  const [discardState, discardAction, discardPending] = useActionState(
    discardRawAction,
    EMPTY_RECORD_SESSION_ACTION_STATE,
  );
  useRefreshAfterAction(acceptState);
  useRefreshAfterAction(discardState);

  if (raw.status !== "CAPTURED" || !canMutate) {
    return null;
  }

  return (
    <div className="flex items-center gap-2">
      <form action={acceptAction}>
        <input type="hidden" name="case_id" value={caseId} />
        <input type="hidden" name="raw_id" value={raw.id} />
        <Button type="submit" disabled={acceptPending || discardPending}>
          {acceptPending ? "Accepting…" : "Accept"}
        </Button>
      </form>
      <form action={discardAction}>
        <input type="hidden" name="case_id" value={caseId} />
        <input type="hidden" name="raw_id" value={raw.id} />
        <Button
          type="submit"
          variant="secondary"
          disabled={acceptPending || discardPending}
        >
          {discardPending ? "Discarding…" : "Discard"}
        </Button>
      </form>
      <ActionMessage state={acceptState} />
      <ActionMessage state={discardState} />
    </div>
  );
}
