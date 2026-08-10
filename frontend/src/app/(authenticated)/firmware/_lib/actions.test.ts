import { beforeEach, describe, expect, it, vi } from "vitest";

import { replaceFirmwareBinary } from "@/lib/api/firmwares";
import { requireSessionContext } from "@/lib/session";
import { formData } from "@/test/form-data";
import { USER } from "@/test/fixtures";

import { replaceFirmwareBinaryAction, type FormActionState } from "./actions";

vi.mock("server-only", () => ({}));

vi.mock("next/navigation", () => ({
  redirect: vi.fn(),
}));

vi.mock("@/lib/api/firmwares", () => ({
  createFirmware: vi.fn(),
  deleteFirmware: vi.fn(),
  replaceFirmwareBinary: vi.fn(),
  updateFirmware: vi.fn(),
}));

vi.mock("@/lib/api/ota", () => ({
  dispatchOtaByNodeId: vi.fn(),
}));

vi.mock("@/lib/session", () => ({
  requireSessionContext: vi.fn(),
}));

const INITIAL_STATE: FormActionState = { status: "idle" };
const BINARY = new File(["firmware"], "firmware.bin", {
  type: "application/octet-stream",
});

function permit(...permissions: string[]): void {
  vi.mocked(requireSessionContext).mockResolvedValue({
    user: USER,
    permissions: new Set(permissions),
  });
}

function replacementData(mode: string, configSchema?: string | Blob): FormData {
  const data = formData({
    firmware_id: "firmware-1",
    file: BINARY,
    schema_intent: mode,
  });
  if (configSchema !== undefined) {
    data.set("config_schema", configSchema);
  }
  return data;
}

describe("replaceFirmwareBinaryAction", () => {
  beforeEach(() => {
    vi.resetAllMocks();
  });

  it("rejects malformed direct-call input without reaching transport", async () => {
    permit("firmware:set");

    const result = await replaceFirmwareBinaryAction(
      INITIAL_STATE,
      null as unknown as FormData,
    );

    expect(result).toMatchObject({
      status: "error",
      title: "Check the replacement",
    });
    expect(replaceFirmwareBinary).not.toHaveBeenCalled();
  });

  it("rejects an unknown schema intent", async () => {
    permit("firmware:set");

    const result = await replaceFirmwareBinaryAction(
      INITIAL_STATE,
      replacementData("merge"),
    );

    expect(result).toMatchObject({
      status: "error",
      fieldErrors: { schema_intent: expect.stringMatching(/choose/i) },
    });
    expect(replaceFirmwareBinary).not.toHaveBeenCalled();
  });

  it("rejects invalid replacement schema JSON", async () => {
    permit("firmware:set");

    const result = await replaceFirmwareBinaryAction(
      INITIAL_STATE,
      replacementData("replace", "not-json"),
    );

    expect(result).toMatchObject({
      status: "error",
      fieldErrors: { config_schema: expect.stringMatching(/valid json/i) },
    });
    expect(replaceFirmwareBinary).not.toHaveBeenCalled();
  });

  it.each([
    ["keep", undefined, undefined],
    [
      "replace",
      '[{"key":"rate","value_type":"uint32"}]',
      [{ key: "rate", value_type: "uint32" }],
    ],
    ["clear", undefined, []],
  ] as const)(
    "maps %s intent to the transport contract",
    async (mode, rawSchema, expectedSchema) => {
      permit("firmware:set");

      const result = await replaceFirmwareBinaryAction(
        INITIAL_STATE,
        replacementData(mode, rawSchema),
      );

      expect(result.status).toBe("success");
      expect(replaceFirmwareBinary).toHaveBeenCalledWith({
        id: "firmware-1",
        binary: BINARY,
        configSchema: expectedSchema,
      });
      expect(
        vi.mocked(requireSessionContext).mock.invocationCallOrder[0],
      ).toBeLessThan(
        vi.mocked(replaceFirmwareBinary).mock.invocationCallOrder[0],
      );
    },
  );

  it("requires exactly firmware:set before transport", async () => {
    permit("firmware:get", "firmware:add");

    const result = await replaceFirmwareBinaryAction(
      INITIAL_STATE,
      replacementData("keep"),
    );

    expect(result).toMatchObject({
      status: "error",
      title: "Permission denied",
    });
    expect(requireSessionContext).toHaveBeenCalledOnce();
    expect(replaceFirmwareBinary).not.toHaveBeenCalled();
  });
});
