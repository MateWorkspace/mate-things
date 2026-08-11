import type { ActionState } from "@/lib/forms/action-state";

export type InfraredStateActionState = ActionState<string>;

export const EMPTY_INFRARED_STATE_STATE: InfraredStateActionState = {
  status: "idle",
};
