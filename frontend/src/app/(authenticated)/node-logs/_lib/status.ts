import type { StatusVariant } from "@/components/ui/status-badge";
import type { NodeLogLevel } from "@/lib/api/node-logs";

export const NODE_LOG_LEVEL_LABELS: Record<
  NodeLogLevel,
  { label: string; variant: StatusVariant }
> = {
  NONE: { label: "None", variant: "neutral" },
  ERROR: { label: "Error", variant: "critical" },
  WARN: { label: "Warning", variant: "warning" },
  INFO: { label: "Info", variant: "info" },
  DEBUG: { label: "Debug", variant: "neutral" },
};
