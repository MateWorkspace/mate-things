import { refresh } from "next/cache";
import { redirect } from "next/navigation";
import { beforeEach, describe, expect, it, vi } from "vitest";

import {
  createFirmware,
  deleteFirmware,
  getFirmwareBinaryUrlById,
  listAvailableFirmwaresByNodeId,
  replaceFirmwareBinary,
  updateFirmware,
} from "@/lib/api/firmwares";
import { dispatchOtaByNodeId } from "@/lib/api/ota";
import { requireSessionContext } from "@/lib/session";
import { formData } from "@/test/form-data";
import { USER } from "@/test/fixtures";

import {
  createFirmwareAction,
  deleteFirmwareAction,
  dispatchOtaAction,
  replaceFirmwareBinaryAction,
  updateFirmwareAction,
  type FormActionState,
} from "./actions";

vi.mock("server-only", () => ({}));

vi.mock("next/cache", () => ({
  refresh: vi.fn(),
}));

vi.mock("next/navigation", () => ({
  redirect: vi.fn(),
}));

vi.mock("@/lib/api/firmwares", () => ({
  createFirmware: vi.fn(),
  deleteFirmware: vi.fn(),
  getFirmwareBinaryUrlById: vi.fn(),
  listAvailableFirmwaresByNodeId: vi.fn(),
  replaceFirmwareBinary: vi.fn(),
  updateFirmware: vi.fn(),
}));

vi.mock("@/lib/api/ota", () => ({
  dispatchOtaByNodeId: vi.fn(),
}));

vi.mock("@/lib/session", () => ({
  requireSessionContext: vi.fn(),
}));

const EMPTY_STATE: FormActionState = { status: "idle" };
const FILE = new File(["binary"], "freezer.bin", {
  type: "application/octet-stream",
});

function permit(...permissions: string[]) {
  vi.mocked(requireSessionContext).mockResolvedValue({
    user: USER,
    permissions: new Set(permissions),
  });
}

function firmwareFormData(
  overrides: Record<string, string | File> = {},
): FormData {
  const data = formData({
    node_class_id: "class-1",
    name: "freezer-v2",
    file: FILE,
    ...overrides,
  });
  data.append("schema_key", "sample_rate");
  data.append("schema_value_type", "uint32");
  return data;
}

describe("firmware actions", () => {
  beforeEach(() => {
    vi.resetAllMocks();
  });

  it("rechecks firmware:add before uploading", async () => {
    permit("firmware:get");

    const result = await createFirmwareAction(EMPTY_STATE, firmwareFormData());

    expect(result).toMatchObject({
      status: "error",
      title: "Permission denied",
    });
    expect(createFirmware).not.toHaveBeenCalled();
  });

  it("validates duplicate configuration keys before uploading", async () => {
    permit("firmware:add");
    const data = firmwareFormData();
    data.append("schema_key", "sample_rate");
    data.append("schema_value_type", "bool");

    const result = await createFirmwareAction(EMPTY_STATE, data);

    expect(result).toMatchObject({
      status: "error",
      fieldErrors: { config_schema: expect.stringMatching(/unique/i) },
    });
    expect(createFirmware).not.toHaveBeenCalled();
  });

  it("uploads trimmed metadata, binary, and normalized schema", async () => {
    permit("firmware:add");

    const result = await createFirmwareAction(
      EMPTY_STATE,
      firmwareFormData({
        node_class_id: "  class-1  ",
        name: "  freezer-v2  ",
      }),
    );

    expect(createFirmware).toHaveBeenCalledWith("class-1", "freezer-v2", FILE, [
      { key: "sample_rate", value_type: "uint32" },
    ]);
    expect(refresh).toHaveBeenCalledOnce();
    expect(result.status).toBe("success");
  });

  it("accepts an intentionally empty configuration schema", async () => {
    permit("firmware:add");
    const data = formData({
      node_class_id: "class-1",
      name: "freezer-v2",
      file: FILE,
    });
    data.append("schema_key", "");
    data.append("schema_value_type", "");

    const result = await createFirmwareAction(EMPTY_STATE, data);

    expect(createFirmware).toHaveBeenCalledWith(
      "class-1",
      "freezer-v2",
      FILE,
      [],
    );
    expect(result.status).toBe("success");
  });

  it("rechecks firmware:set before editing metadata", async () => {
    permit("firmware:get");

    const result = await updateFirmwareAction(
      EMPTY_STATE,
      formData({
        firmware_id: "firmware-1",
        node_class_id: "class-1",
        name: "freezer-v2",
      }),
    );

    expect(result.title).toBe("Permission denied");
    expect(updateFirmware).not.toHaveBeenCalled();
  });

  it("rechecks firmware:set before replacing a binary", async () => {
    permit("firmware:get");

    const result = await replaceFirmwareBinaryAction(
      EMPTY_STATE,
      firmwareFormData({ firmware_id: "firmware-1" }),
    );

    expect(result.title).toBe("Permission denied");
    expect(replaceFirmwareBinary).not.toHaveBeenCalled();
  });

  it("updates trimmed firmware metadata", async () => {
    permit("firmware:set");

    const result = await updateFirmwareAction(
      EMPTY_STATE,
      formData({
        firmware_id: "  firmware-1  ",
        node_class_id: "  class-2  ",
        name: "  freezer-v3  ",
      }),
    );

    expect(updateFirmware).toHaveBeenCalledWith("firmware-1", {
      node_class_id: "class-2",
      name: "freezer-v3",
    });
    expect(refresh).toHaveBeenCalledOnce();
    expect(result.status).toBe("success");
  });

  it("replaces the binary with the submitted schema", async () => {
    permit("firmware:set");

    const result = await replaceFirmwareBinaryAction(
      EMPTY_STATE,
      firmwareFormData({ firmware_id: "firmware-1" }),
    );

    expect(replaceFirmwareBinary).toHaveBeenCalledWith("firmware-1", FILE, [
      { key: "sample_rate", value_type: "uint32" },
    ]);
    expect(refresh).toHaveBeenCalledOnce();
    expect(result.status).toBe("success");
  });

  it("rechecks firmware:remove before deleting", async () => {
    permit("firmware:get");

    const result = await deleteFirmwareAction(
      EMPTY_STATE,
      formData({
        firmware_id: "firmware-1",
        firmware_name: "freezer-v2",
        confirmation: "freezer-v2",
      }),
    );

    expect(result.title).toBe("Permission denied");
    expect(deleteFirmware).not.toHaveBeenCalled();
  });

  it("requires the exact firmware name before deleting", async () => {
    permit("firmware:remove");

    const result = await deleteFirmwareAction(
      EMPTY_STATE,
      formData({
        firmware_id: "firmware-1",
        firmware_name: "freezer-v2",
        confirmation: "Freezer-v2",
      }),
    );

    expect(result.fieldErrors?.confirmation).toMatch(/exactly match/i);
    expect(deleteFirmware).not.toHaveBeenCalled();
  });

  it("deletes and redirects only after backend success", async () => {
    permit("firmware:remove");

    await deleteFirmwareAction(
      EMPTY_STATE,
      formData({
        firmware_id: "firmware-1",
        firmware_name: "freezer-v2",
        confirmation: "freezer-v2",
      }),
    );

    expect(deleteFirmware).toHaveBeenCalledWith("firmware-1");
    expect(redirect).toHaveBeenCalledWith("/firmware");
    expect(vi.mocked(deleteFirmware).mock.invocationCallOrder[0]).toBeLessThan(
      vi.mocked(redirect).mock.invocationCallOrder[0],
    );
  });
});

describe("OTA action", () => {
  beforeEach(() => {
    vi.resetAllMocks();
  });

  it("rechecks ota:dispatch before reading compatibility", async () => {
    permit("firmware:get");

    const result = await dispatchOtaAction({
      nodeId: "node-1",
      firmwareId: "firmware-2",
    });

    expect(result.title).toBe("Permission denied");
    expect(listAvailableFirmwaresByNodeId).not.toHaveBeenCalled();
    expect(dispatchOtaByNodeId).not.toHaveBeenCalled();
  });

  it("rejects a malformed direct Server Action input", async () => {
    permit("ota:dispatch");

    const result = await dispatchOtaAction(
      null as unknown as Parameters<typeof dispatchOtaAction>[0],
    );

    expect(result).toMatchObject({
      status: "error",
      title: "Check the OTA request",
    });
    expect(listAvailableFirmwaresByNodeId).not.toHaveBeenCalled();
  });

  it("does not dispatch incompatible firmware", async () => {
    permit("ota:dispatch");
    vi.mocked(listAvailableFirmwaresByNodeId).mockResolvedValue({
      data: [],
      page: { page: 1, limit: 48, total_items: 0 },
    });

    const result = await dispatchOtaAction({
      nodeId: "node-1",
      firmwareId: "firmware-2",
    });

    expect(result.message).toMatch(/not available/i);
    expect(dispatchOtaByNodeId).not.toHaveBeenCalled();
    expect(getFirmwareBinaryUrlById).not.toHaveBeenCalled();
  });

  it("rechecks compatibility across available-firmware pages and resolves the backend URL", async () => {
    permit("ota:dispatch");
    vi.mocked(listAvailableFirmwaresByNodeId)
      .mockResolvedValueOnce({
        data: [
          {
            id: "firmware-1",
            node_class_id: "class-1",
            name: "freezer-v1",
            size: 1024,
            checksum: "aaa",
            binary_path: "firmware/freezer-v1.bin",
            preferences: {},
            created_at: "2026-07-30T00:00:00Z",
          },
        ],
        page: { page: 1, limit: 1, total_items: 2 },
      })
      .mockResolvedValueOnce({
        data: [
          {
            id: "firmware-2",
            node_class_id: "class-1",
            name: "freezer-v2",
            size: 2048,
            checksum: "bbb",
            binary_path: "firmware/freezer-v2.bin",
            preferences: {},
            created_at: "2026-07-30T00:00:00Z",
          },
        ],
        page: { page: 2, limit: 1, total_items: 2 },
      });
    vi.mocked(getFirmwareBinaryUrlById).mockResolvedValue(
      "https://storage.example/signed-firmware",
    );

    const result = await dispatchOtaAction({
      nodeId: "node-1",
      firmwareId: "firmware-2",
    });

    expect(listAvailableFirmwaresByNodeId).toHaveBeenNthCalledWith(
      1,
      "node-1",
      { page: 1, limit: 48 },
    );
    expect(listAvailableFirmwaresByNodeId).toHaveBeenNthCalledWith(
      2,
      "node-1",
      { page: 2, limit: 48 },
    );
    expect(getFirmwareBinaryUrlById).toHaveBeenCalledWith("firmware-2");
    expect(dispatchOtaByNodeId).toHaveBeenCalledWith("node-1", {
      firmware_id: "firmware-2",
      firmware_url: "https://storage.example/signed-firmware",
    });
    expect(result.status).toBe("success");
  });
});
