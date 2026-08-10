import type { ActionState } from "@/lib/forms/action-state";

interface ActionMessageProps {
  state: ActionState<string>;
  className?: string;
}

export default function ActionMessage({
  state,
  className = "",
}: ActionMessageProps) {
  if (state.status === "idle" || !state.message) {
    return null;
  }

  const tone = state.status === "error" ? "text-critical" : "text-success";

  return (
    <p
      aria-live={state.status === "error" ? "assertive" : "polite"}
      className={`${tone} text-sm ${className}`}
    >
      {state.message}
    </p>
  );
}
