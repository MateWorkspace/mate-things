import { beforeEach, describe, expect, it, vi } from "vitest";

import { apiFetch } from "@/lib/api/client";
import { getFirmwareConfigParameters } from "@/lib/api/firmwares";

import { getNodeConfig, setNodeConfig } from "./node-config";

vi.mock("@/lib/api/client", () => ({
  apiFetch: vi.fn(),
  apiRequest: vi.fn(),
  buildQuery: vi.fn(),
}));

describe("node configuration API", () => {
  beforeEach(() => {
    vi.resetAllMocks();
  });

  it("reads the selected node's configuration values", async () => {
    const response = [
      {
        key: "sample_rate",
        value: "30",
        updated_at: "2026-07-30T10:00:00Z",
      },
    ];
    vi.mocked(apiFetch).mockResolvedValue(response);

    await expect(getNodeConfig("node-1")).resolves.toEqual(response);

    expect(apiFetch).toHaveBeenCalledWith("/nodes/node-1/config");
  });

  it("sets one node configuration value through the string wire contract", async () => {
    vi.mocked(apiFetch).mockResolvedValue(undefined);

    await setNodeConfig("node-1", "sample_rate", "30");

    expect(apiFetch).toHaveBeenCalledWith("/nodes/node-1/config", {
      method: "PUT",
      body: { key: "sample_rate", value: "30" },
    });
  });

  it("reads a firmware's configuration parameter schema", async () => {
    const response = [{ key: "sample_rate", value_type: "uint32" }];
    vi.mocked(apiFetch).mockResolvedValue(response);

    await expect(getFirmwareConfigParameters("firmware-1")).resolves.toEqual(
      response,
    );

    expect(apiFetch).toHaveBeenCalledWith(
      "/firmwares/firmware-1/config-parameters",
    );
  });
});
