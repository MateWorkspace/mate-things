"use client";

import { useActionState } from "react";

import Button from "@/components/ui/button";
import StatusBadge from "@/components/ui/status-badge";

import { testLlmConnectionAction } from "../_lib/actions";
import { EMPTY_TEST_CONNECTION_STATE } from "../_lib/state";

export default function TestConnectionButton() {
  const [state, action, pending] = useActionState(
    testLlmConnectionAction,
    EMPTY_TEST_CONNECTION_STATE,
  );

  return (
    <form action={action} className="flex flex-wrap items-center gap-3">
      <Button type="submit" variant="secondary" disabled={pending}>
        {pending ? "Testing…" : "Test connection"}
      </Button>
      {state.status === "result" ? (
        <StatusBadge
          variant={
            state.connectionStatus === "CONNECTED" ? "success" : "critical"
          }
        >
          {state.connectionStatus === "CONNECTED"
            ? "Connected"
            : "Disconnected"}
        </StatusBadge>
      ) : state.status === "error" ? (
        <span className="text-critical text-sm">{state.message}</span>
      ) : null}
    </form>
  );
}
