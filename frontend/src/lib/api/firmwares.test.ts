import { beforeEach, describe, expect, it, vi } from "vitest";

import { apiFetch } from "@/lib/api/client";

import {
  replaceFirmwareBinary,
  type FirmwareConfigParameter,
} from "./firmwares";

vi.mock("server-only", () => ({}));

vi.mock("@/lib/api/client", () => ({
  apiFetch: vi.fn(),
  apiRequest: vi.fn(),
  buildQuery: vi.fn(() => ""),
}));

describe("replaceFirmwareBinary", () => {
  beforeEach(() => {
    vi.resetAllMocks();
  });

  const cases: Array<
    [string, FirmwareConfigParameter[] | undefined, boolean, string | undefined]
  > = [
    ["keep", undefined, false, undefined],
    [
      "replace",
      [{ key: "rate", value_type: "uint32" }],
      true,
      '[{"key":"rate","value_type":"uint32"}]',
    ],
    ["clear", [], true, "[]"],
  ];

  it.each(cases)(
    "serializes %s schema intent",
    async (_mode, configSchema, present, expected) => {
      const binary = new File(["firmware"], "firmware.bin", {
        type: "application/octet-stream",
      });

      await replaceFirmwareBinary({
        id: "firmware-1",
        binary,
        configSchema,
      });

      const [path, options] = vi.mocked(apiFetch).mock.calls[0];
      expect(path).toBe("/firmwares/firmware-1/binary");
      expect(options?.method).toBe("PUT");

      const body = options?.body as FormData;
      expect(body.get("file")).toBe(binary);
      expect(body.has("config_schema")).toBe(present);
      if (present) {
        expect(body.get("config_schema")).toBe(expected);
      }
    },
  );
});
