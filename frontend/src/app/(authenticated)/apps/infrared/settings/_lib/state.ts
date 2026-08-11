import type { ActionState } from "@/lib/forms/action-state";

export type InfraredSettingsActionState = ActionState<string>;

export const EMPTY_INFRARED_SETTINGS_STATE: InfraredSettingsActionState = {
  status: "idle",
};
