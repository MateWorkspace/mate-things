import {
  INITIAL_ACTION_STATE,
  type ActionState,
} from "@/lib/forms/action-state";

export type PayloadSchemaActionState = ActionState<string>;

export const EMPTY_SCHEMA_STATE: PayloadSchemaActionState =
  INITIAL_ACTION_STATE;
