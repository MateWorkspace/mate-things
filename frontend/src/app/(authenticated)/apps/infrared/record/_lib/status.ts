import type { StatusVariant } from "@/components/ui/status-badge";
import type { InfraredRecordingState } from "@/lib/api/infrared";

export const RECORDING_STATE_LABELS: Record<
  InfraredRecordingState,
  { label: string; variant: StatusVariant }
> = {
  DRAFT: { label: "Draft", variant: "neutral" },
  CASES_GENERATING: { label: "Building cases", variant: "info" },
  RECORDING: { label: "Recording", variant: "info" },
  ANALYZING: { label: "Analyzing", variant: "info" },
  FUNCTION_GENERATING: { label: "Generating encoder", variant: "info" },
  TEST_CASES_GENERATING: { label: "Building tests", variant: "info" },
  TESTING: { label: "Testing", variant: "info" },
  COMPLETED: { label: "Completed", variant: "success" },
  FAILED: { label: "Failed", variant: "critical" },
};
