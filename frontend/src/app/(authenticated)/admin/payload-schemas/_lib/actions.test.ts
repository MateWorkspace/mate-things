import { beforeEach, describe, expect, it, vi } from "vitest";

import {
  createPayloadSchema,
  deletePayloadSchema,
  updatePayloadSchema,
} from "@/lib/api/payload-schemas";
import { requireSessionContext } from "@/lib/session";
import { formData } from "@/test/form-data";
import { USER } from "@/test/fixtures";

import {
  createPayloadSchemaAction,
  updatePayloadSchemaAction,
} from "./actions";
import { EMPTY_SCHEMA_STATE } from "./state";

vi.mock("server-only", () => ({}));

vi.mock("next/navigation", () => ({
  redirect: vi.fn(),
}));

vi.mock("@/lib/api/payload-schemas", () => ({
  createPayloadSchema: vi.fn(),
  deletePayloadSchema: vi.fn(),
  updatePayloadSchema: vi.fn(),
}));

vi.mock("@/lib/session", () => ({
  requireSessionContext: vi.fn(),
}));

function validSchemaData(extra: Record<string, string | Blob> = {}): FormData {
  return formData({
    name: "temperature_reading",
    version: "1",
    definition: '{"temperature":{"type":"number"}}',
    ...extra,
  });
}

function permit(...permissions: string[]) {
  vi.mocked(requireSessionContext).mockResolvedValue({
    user: USER,
    permissions: new Set(permissions),
  });
}

describe("payload schema actions", () => {
  beforeEach(() => {
    vi.resetAllMocks();
  });

  it("never creates when an update id is missing", async () => {
    const result = await updatePayloadSchemaAction(
      EMPTY_SCHEMA_STATE,
      validSchemaData(),
    );

    expect(result).toMatchObject({
      status: "error",
      title: "Cannot update schema",
      fieldErrors: { payload_schema_id: expect.any(String) },
    });
    expect(requireSessionContext).not.toHaveBeenCalled();
    expect(createPayloadSchema).not.toHaveBeenCalled();
    expect(updatePayloadSchema).not.toHaveBeenCalled();
  });

  it("rejects a non-text update identity before authorization or transport", async () => {
    const result = await updatePayloadSchemaAction(
      "schema-1",
      EMPTY_SCHEMA_STATE,
      validSchemaData({
        payload_schema_id: new Blob(["schema-1"], { type: "text/plain" }),
      }),
    );

    expect(result.status).toBe("error");
    expect(requireSessionContext).not.toHaveBeenCalled();
    expect(createPayloadSchema).not.toHaveBeenCalled();
    expect(updatePayloadSchema).not.toHaveBeenCalled();
  });

  it("rejects a mismatched update identity before authorization or transport", async () => {
    const result = await updatePayloadSchemaAction(
      "schema-1",
      EMPTY_SCHEMA_STATE,
      validSchemaData({ payload_schema_id: "schema-2" }),
    );

    expect(result).toMatchObject({
      status: "error",
      title: "Cannot update schema",
      fieldErrors: { payload_schema_id: expect.any(String) },
    });
    expect(requireSessionContext).not.toHaveBeenCalled();
    expect(createPayloadSchema).not.toHaveBeenCalled();
    expect(updatePayloadSchema).not.toHaveBeenCalled();
  });

  it("keeps create and update as separate permission and transport paths", async () => {
    permit("payload_schema:set");

    const result = await updatePayloadSchemaAction(
      "schema-1",
      EMPTY_SCHEMA_STATE,
      validSchemaData({ payload_schema_id: "schema-1" }),
    );

    expect(result.status).toBe("success");
    expect(updatePayloadSchema).toHaveBeenCalledWith(
      "schema-1",
      expect.objectContaining({ name: "temperature_reading", version: 1 }),
    );
    expect(createPayloadSchema).not.toHaveBeenCalled();

    vi.resetAllMocks();
    permit("payload_schema:add");

    const createResult = await createPayloadSchemaAction(
      EMPTY_SCHEMA_STATE,
      validSchemaData(),
    );

    expect(createResult.status).toBe("success");
    expect(createPayloadSchema).toHaveBeenCalledOnce();
    expect(updatePayloadSchema).not.toHaveBeenCalled();
    expect(deletePayloadSchema).not.toHaveBeenCalled();
  });
});
