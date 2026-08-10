import {
  INITIAL_ACTION_STATE,
  type ActionState,
} from "@/lib/forms/action-state";

export type AccessActionState = ActionState<string>;

export const EMPTY_ACCESS_STATE: AccessActionState = INITIAL_ACTION_STATE;
