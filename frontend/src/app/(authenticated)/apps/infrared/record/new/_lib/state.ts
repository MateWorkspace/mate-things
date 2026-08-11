import type { InfraredStateType } from "@/lib/api/infrared";
import type { ActionState } from "@/lib/forms/action-state";

export type NewRecordSessionActionState = ActionState<string>;

export const EMPTY_NEW_RECORD_SESSION_STATE: NewRecordSessionActionState = {
  status: "idle",
};

export interface WizardCreatedItem {
  id: string;
  name: string;
  type?: InfraredStateType;
}

export type WizardCreateActionState = ActionState<string> & {
  created?: WizardCreatedItem;
};

export const EMPTY_WIZARD_CREATE_STATE: WizardCreateActionState = {
  status: "idle",
};
