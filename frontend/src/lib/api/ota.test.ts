import { beforeEach, describe, expect, it, vi } from "vitest";

import { apiFetch } from "@/lib/api/client";

import { dispatchOtaByNodeId } from "./ota";

vi.mock("@/lib/api/client", () => ({
  apiFetch: vi.fn(),
}));

describe("OTA API", () => {
  beforeEach(() => {
    vi.resetAllMocks();
  });

  it("dispatches by firmware identity without accepting a caller-controlled URL", async () => {
    await dispatchOtaByNodeId("node-1", { firmware_id: "firmware-2" });

    expect(apiFetch).toHaveBeenCalledWith("/nodes/node-1/ota", {
      method: "POST",
      body: { firmware_id: "firmware-2" },
    });
  });
});
