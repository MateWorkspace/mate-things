import {
  INITIAL_ACTION_STATE,
  type ActionState,
} from "@/lib/forms/action-state";

export type UserActionState = ActionState<string>;

export const EMPTY_USER_STATE: UserActionState = INITIAL_ACTION_STATE;
