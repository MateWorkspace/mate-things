import { beforeEach, describe, expect, it, vi } from "vitest";

import { apiFetch } from "@/lib/api/client";

import { createFirmware, replaceFirmwareBinary } from "./firmwares";

vi.mock("@/lib/api/client", () => ({
  apiFetch: vi.fn(),
  apiRequest: vi.fn(),
  buildQuery: vi.fn(() => ""),
}));

describe("firmware multipart wrappers", () => {
  beforeEach(() => {
    vi.resetAllMocks();
  });

  it("uploads the firmware binary with its configuration schema", async () => {
    const file = new File(["binary"], "freezer.bin", {
      type: "application/octet-stream",
    });

    await createFirmware("class-1", "freezer-v2", file, [
      { key: "sample_rate", value_type: "uint32" },
      { key: "enabled", value_type: "bool" },
    ]);

    const [path, options] = vi.mocked(apiFetch).mock.calls[0];
    expect(path).toBe("/firmwares");
    expect(options?.method).toBe("POST");
    expect(options?.body).toBeInstanceOf(FormData);

    const body = options?.body as FormData;
    expect(body.get("node_class_id")).toBe("class-1");
    expect(body.get("name")).toBe("freezer-v2");
    expect(body.get("file")).toBe(file);
    expect(body.get("config_schema")).toBe(
      '[{"key":"sample_rate","value_type":"uint32"},{"key":"enabled","value_type":"bool"}]',
    );
  });

  it("replaces the binary and sends the complete updated schema", async () => {
    const file = new File(["replacement"], "freezer-v3.bin");

    await replaceFirmwareBinary("firmware-1", file, [
      { key: "broker", value_type: "string" },
    ]);

    const [path, options] = vi.mocked(apiFetch).mock.calls[0];
    expect(path).toBe("/firmwares/firmware-1/binary");
    expect(options?.method).toBe("PUT");

    const body = options?.body as FormData;
    expect(body.get("file")).toBe(file);
    expect(body.get("config_schema")).toBe(
      '[{"key":"broker","value_type":"string"}]',
    );
  });
});
