import {
  INITIAL_ACTION_STATE,
  type ActionState,
} from "@/lib/forms/action-state";

export type ApiKeyActionState = ActionState<string> & {
  key?: string;
};

export const EMPTY_API_KEY_STATE: ApiKeyActionState = INITIAL_ACTION_STATE;
