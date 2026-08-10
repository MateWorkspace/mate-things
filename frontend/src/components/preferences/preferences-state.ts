import {
  INITIAL_ACTION_STATE,
  type ActionState,
} from "@/lib/forms/action-state";

export type PreferencesActionState = ActionState<string>;

export const EMPTY_PREFERENCES_STATE: PreferencesActionState =
  INITIAL_ACTION_STATE;
