"use client";

import { useRouter } from "next/navigation";

import { useInfraredRecordSessionBroadcast } from "@/hooks/use-infrared-record-session-broadcast";
import { useSmartRefresh } from "@/hooks/use-smart-refresh";
import type {
  InfraredRecordSessionResponse,
  InfraredStateDeviceRecordCaseResponse,
  InfraredStateResponse,
  InfraredTestCaseResponse,
} from "@/lib/api/infrared";

import WizardStepper from "../../_components/WizardStepper";
import { getBroadcastToken } from "../_lib/broadcast-token";
import RecordCaseList from "./RecordCaseList";
import RecordTestCaseList from "./RecordTestCaseList";
import WizardCaseRecorder from "./WizardCaseRecorder";

const WIZARD_STEPS = [
  "Device details",
  "State values",
  "Building cases",
  "Review & set default",
  "Recording",
  "Analyzing",
  "Testing",
  "Finishing",
] as const;

function currentStepIndex(
  recordingState: InfraredRecordSessionResponse["recording_state"],
  anyRawExists: boolean,
): number {
  switch (recordingState) {
    case "DRAFT":
    case "CASES_GENERATING":
      return 2;
    case "RECORDING":
      return anyRawExists ? 4 : 3;
    case "ANALYZING":
    case "FUNCTION_GENERATING":
      return 5;
    case "TEST_CASES_GENERATING":
    case "TESTING":
      return 6;
    case "COMPLETED":
    case "FAILED":
      return 7;
  }
}

export default function RecordSessionWizard({
  session,
  cases,
  testCases,
  canMutate,
  stateDefinitions,
}: {
  session: InfraredRecordSessionResponse;
  cases: readonly InfraredStateDeviceRecordCaseResponse[];
  testCases: readonly InfraredTestCaseResponse[];
  canMutate: boolean;
  stateDefinitions: readonly InfraredStateResponse[];
}) {
  const router = useRouter();
  const { exhausted } = useInfraredRecordSessionBroadcast(
    session.id,
    () => router.refresh(),
    getBroadcastToken,
  );
  // Live updates arrive over the WebSocket; polling only takes over once
  // reconnect attempts are exhausted (e.g. a second tab already holds the
  // session's single allowed listener).
  useSmartRefresh({
    intervalMs: 5000,
    suspended: !exhausted,
    onRefresh: () => {
      router.refresh();
    },
  });

  const anyRawExists = cases.some((c) => c.raw.length > 0);
  const stepIndex = currentStepIndex(session.recording_state, anyRawExists);

  return (
    <div className="space-y-6">
      <WizardStepper steps={WIZARD_STEPS} currentIndex={stepIndex} />

      {session.recording_state === "DRAFT" ||
      session.recording_state === "CASES_GENERATING" ? (
        <div className="border-border bg-muted rounded-2xl border p-6 text-center">
          <p className="font-semibold">Building the recording plan…</p>
          <p className="text-foreground/70 mt-1 text-sm">
            The AI is drafting cases for every state combination. This page
            updates automatically.
          </p>
        </div>
      ) : null}

      {session.recording_state === "RECORDING" && !anyRawExists ? (
        <div className="border-border space-y-4 rounded-2xl border p-6">
          <div>
            <p className="font-semibold">Review the recording plan</p>
            <p className="text-foreground/70 mt-1 text-sm">
              Each case below asks you to press the remote once its target state
              is active. Before you start, set the device to its default/off
              state so the first capture has a clean baseline.
            </p>
          </div>
          <RecordCaseList
            cases={cases}
            canMutate={false}
            stateDefinitions={stateDefinitions}
          />
        </div>
      ) : null}

      {session.recording_state === "RECORDING" && anyRawExists ? (
        <WizardCaseRecorder
          cases={cases}
          canMutate={canMutate}
          stateDefinitions={stateDefinitions}
        />
      ) : null}

      {session.recording_state === "ANALYZING" ||
      session.recording_state === "FUNCTION_GENERATING" ? (
        <div className="border-border bg-muted rounded-2xl border p-6 text-center">
          <p className="font-semibold">
            {session.recording_state === "FUNCTION_GENERATING"
              ? "Refining the encoder…"
              : "Analyzing captures…"}
          </p>
          <p className="text-foreground/70 mt-1 text-sm">
            This page updates automatically once the coder is ready to test.
          </p>
        </div>
      ) : null}

      {session.recording_state === "TEST_CASES_GENERATING" ? (
        <div className="border-border bg-muted rounded-2xl border p-6 text-center">
          <p className="font-semibold">Building test cases…</p>
        </div>
      ) : null}

      {session.recording_state === "TESTING" ? (
        <RecordTestCaseList testCases={testCases} canMutate={canMutate} />
      ) : null}
    </div>
  );
}
