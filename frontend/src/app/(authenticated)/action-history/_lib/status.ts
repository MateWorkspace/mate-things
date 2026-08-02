import type { StatusVariant } from "@/components/ui/status-badge";
import type { ActionStatus } from "@/lib/api/action-logs";

export const ACTION_STATUS_LABELS: Record<
  ActionStatus,
  { label: string; variant: StatusVariant }
> = {
  UNEXECUTED: { label: "Unexecuted", variant: "neutral" },
  UNRESPONDED: { label: "Unresponded", variant: "warning" },
  FAILED: { label: "Failed", variant: "critical" },
  SUCCESS: { label: "Success", variant: "success" },
};
