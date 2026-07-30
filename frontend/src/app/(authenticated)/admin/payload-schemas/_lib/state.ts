export interface PayloadSchemaActionState {
  status: "idle" | "success" | "error";
  title?: string;
  message?: string;
  fieldErrors?: Record<string, string>;
}

export const EMPTY_SCHEMA_STATE: PayloadSchemaActionState = { status: "idle" };
