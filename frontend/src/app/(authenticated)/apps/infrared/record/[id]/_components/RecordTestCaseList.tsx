import StatusBadge from "@/components/ui/status-badge";
import type { InfraredTestCaseResponse } from "@/lib/api/infrared";

import RecordTestCaseControls from "./RecordTestCaseControls";

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
                <RecordTestCaseControls testCaseId={tc.id} />
              ) : null}
            </div>
          </div>
        </li>
      ))}
    </ul>
  );
}
