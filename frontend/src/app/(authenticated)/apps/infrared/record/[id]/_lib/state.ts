import type { ActionState } from "@/lib/forms/action-state";

export type RecordSessionActionState = ActionState<string>;

export const EMPTY_RECORD_SESSION_ACTION_STATE: RecordSessionActionState = {
  status: "idle",
};
