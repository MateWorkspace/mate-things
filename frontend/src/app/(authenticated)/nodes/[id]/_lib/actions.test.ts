import { refresh } from "next/cache";
import { redirect } from "next/navigation";
import { beforeEach, describe, expect, it, vi } from "vitest";

import { setNodeConfig } from "@/lib/api/node-config";
import { deleteNode, getNodeById, updateNode } from "@/lib/api/nodes";
import { requireSessionContext } from "@/lib/session";
import { formData } from "@/test/form-data";
import { nodeFixture, USER } from "@/test/fixtures";

import {
  deleteNodeAction,
  saveNodeAction,
  saveNodeConfigAction,
  type NodeActionState,
} from "./actions";

vi.mock("server-only", () => ({}));

vi.mock("next/cache", () => ({
  refresh: vi.fn(),
}));

vi.mock("next/navigation", () => ({
  redirect: vi.fn(),
}));

vi.mock("@/lib/api/node-config", () => ({
  setNodeConfig: vi.fn(),
}));

vi.mock("@/lib/api/nodes", () => ({
  deleteNode: vi.fn(),
  getNodeById: vi.fn(),
  updateNode: vi.fn(),
}));

vi.mock("@/lib/session", () => ({
  requireSessionContext: vi.fn(),
}));

const IDLE_STATE: NodeActionState = { status: "idle" };

function permit(...permissions: string[]) {
  vi.mocked(requireSessionContext).mockResolvedValue({
    user: USER,
    permissions: new Set(permissions),
  });
}

describe("node workspace actions", () => {
  beforeEach(() => {
    vi.resetAllMocks();
  });

  it("rechecks node_config:set before validating or setting a value", async () => {
    permit("node_config:get");

    const result = await saveNodeConfigAction(
      IDLE_STATE,
      formData({ node_id: "", key: "", value: "" }),
    );

    expect(result).toMatchObject({
      status: "error",
      title: "Permission denied",
    });
    expect(setNodeConfig).not.toHaveBeenCalled();
  });

  it("sets exactly one configuration key using its string value", async () => {
    permit("node_config:set");

    const result = await saveNodeConfigAction(
      IDLE_STATE,
      formData({
        node_id: "node-1",
        key: "sample_rate",
        value: "30",
      }),
    );

    expect(setNodeConfig).toHaveBeenCalledWith("node-1", "sample_rate", "30");
    expect(refresh).toHaveBeenCalledOnce();
    expect(result.status).toBe("success");
  });

  it("rechecks node:set before updating editable node metadata", async () => {
    permit("node:get");

    const result = await saveNodeAction(
      IDLE_STATE,
      formData({
        node_id: "node-1",
        name: "Freezer 07",
        device_id: "AC276E5E030C",
        description: "Cold room",
      }),
    );

    expect(result).toMatchObject({
      status: "error",
      title: "Permission denied",
    });
    expect(updateNode).not.toHaveBeenCalled();
  });

  it("updates only the editable node metadata after authorization", async () => {
    permit("node:set");

    const result = await saveNodeAction(
      IDLE_STATE,
      formData({
        node_id: "node-1",
        name: "  Freezer 07  ",
        device_id: "  AC276E5E030C  ",
        description: "  Cold room  ",
      }),
    );

    expect(updateNode).toHaveBeenCalledWith("node-1", {
      name: "Freezer 07",
      device_id: "AC276E5E030C",
      description: "Cold room",
    });
    expect(refresh).toHaveBeenCalledOnce();
    expect(result.status).toBe("success");
  });

  it("compares delete confirmation with the backend's current device ID", async () => {
    permit("node:remove");
    vi.mocked(getNodeById).mockResolvedValue(nodeFixture());

    const result = await deleteNodeAction(
      IDLE_STATE,
      formData({
        node_id: "node-1",
        confirmation: "wrong-device",
      }),
    );

    expect(getNodeById).toHaveBeenCalledWith("node-1");
    expect(result).toMatchObject({
      status: "error",
      fieldErrors: { confirmation: expect.any(String) },
    });
    expect(deleteNode).not.toHaveBeenCalled();
    expect(redirect).not.toHaveBeenCalled();
  });

  it("redirects to the fleet only after the confirmed delete succeeds", async () => {
    permit("node:remove");
    vi.mocked(getNodeById).mockResolvedValue(nodeFixture());

    await deleteNodeAction(
      IDLE_STATE,
      formData({
        node_id: "node-1",
        confirmation: "AC276E5E030C",
      }),
    );

    expect(deleteNode).toHaveBeenCalledWith("node-1");
    expect(redirect).toHaveBeenCalledWith("/nodes");
    expect(vi.mocked(deleteNode).mock.invocationCallOrder[0]).toBeLessThan(
      vi.mocked(redirect).mock.invocationCallOrder[0],
    );
  });

  it("does not redirect when the backend rejects deletion", async () => {
    permit("node:remove");
    vi.mocked(getNodeById).mockResolvedValue(nodeFixture());
    vi.mocked(deleteNode).mockRejectedValue(new Error("backend unavailable"));

    const result = await deleteNodeAction(
      IDLE_STATE,
      formData({
        node_id: "node-1",
        confirmation: "AC276E5E030C",
      }),
    );

    expect(result.status).toBe("error");
    expect(redirect).not.toHaveBeenCalled();
  });
});
