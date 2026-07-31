import { redirect } from "next/navigation";
import { beforeEach, describe, expect, it, vi } from "vitest";

import {
  createNodeClass,
  deleteNodeClass,
  getNodeClassById,
  updateNodeClass,
} from "@/lib/api/node-classes";
import { requireSessionContext } from "@/lib/session";
import { formData } from "@/test/form-data";
import { USER } from "@/test/fixtures";

import {
  createNodeClassAction,
  deleteNodeClassAction,
  updateNodeClassAction,
  type FormActionState,
} from "./actions";

vi.mock("server-only", () => ({}));

vi.mock("next/navigation", () => ({
  redirect: vi.fn(),
}));

vi.mock("@/lib/api/node-classes", () => ({
  createNodeClass: vi.fn(),
  deleteNodeClass: vi.fn(),
  getNodeClassById: vi.fn(),
  updateNodeClass: vi.fn(),
}));

vi.mock("@/lib/session", () => ({
  requireSessionContext: vi.fn(),
}));

const EMPTY_STATE: FormActionState = { status: "idle" };

function permit(...permissions: string[]) {
  vi.mocked(requireSessionContext).mockResolvedValue({
    user: USER,
    permissions: new Set(permissions),
  });
}

describe("node class actions", () => {
  beforeEach(() => {
    vi.resetAllMocks();
  });

  it("rejects creation without node_class:add before calling the API", async () => {
    permit("node_class:get");

    const state = await createNodeClassAction(
      EMPTY_STATE,
      formData({
        name: "Cold Storage",
        description: "Temperature nodes",
      }),
    );

    expect(state).toMatchObject({
      status: "error",
      title: "Permission denied",
    });
    expect(createNodeClass).not.toHaveBeenCalled();
  });

  it("rejects a whitespace-only class name after authorization", async () => {
    permit("node_class:add");

    const state = await createNodeClassAction(
      EMPTY_STATE,
      formData({ name: "   ", description: "Temperature nodes" }),
    );

    expect(state).toMatchObject({
      status: "error",
      fieldErrors: { name: "This field is required." },
    });
    expect(createNodeClass).not.toHaveBeenCalled();
  });

  it("creates a class with trimmed form values and refreshes the page", async () => {
    permit("node_class:add");

    const state = await createNodeClassAction(
      EMPTY_STATE,
      formData({
        name: "  Cold Storage  ",
        description: "  Temperature nodes  ",
      }),
    );

    expect(createNodeClass).toHaveBeenCalledWith({
      name: "Cold Storage",
      description: "Temperature nodes",
    });
    expect(state.status).toBe("success");
  });

  it("rechecks node_class:set before updating a class", async () => {
    permit("node_class:get");

    const state = await updateNodeClassAction(
      EMPTY_STATE,
      formData({
        node_class_id: "class-1",
        name: "Cold Storage",
        description: "Temperature nodes",
      }),
    );

    expect(state).toMatchObject({
      status: "error",
      title: "Permission denied",
    });
    expect(updateNodeClass).not.toHaveBeenCalled();
  });

  it("updates a class with trimmed editable values", async () => {
    permit("node_class:set");

    const state = await updateNodeClassAction(
      EMPTY_STATE,
      formData({
        node_class_id: "  class-1  ",
        name: "  Cold Storage  ",
        description: "  Temperature nodes  ",
      }),
    );

    expect(updateNodeClass).toHaveBeenCalledWith("class-1", {
      name: "Cold Storage",
      description: "Temperature nodes",
    });
    expect(state.status).toBe("success");
  });

  it("rejects a typed-name mismatch without reading the class", async () => {
    permit("node_class:remove");

    const state = await deleteNodeClassAction(
      EMPTY_STATE,
      formData({
        node_class_id: "class-1",
        node_class_name: "Cold Storage",
        confirmation: "Cold storage",
      }),
    );

    expect(getNodeClassById).not.toHaveBeenCalled();
    expect(state).toMatchObject({
      status: "error",
      fieldErrors: { confirmation: expect.any(String) },
    });
    expect(deleteNodeClass).not.toHaveBeenCalled();
    expect(redirect).not.toHaveBeenCalled();
  });

  it("rechecks node_class:remove before reading or deleting a class", async () => {
    permit("node_class:get");

    const state = await deleteNodeClassAction(
      EMPTY_STATE,
      formData({
        node_class_id: "class-1",
        node_class_name: "Cold Storage",
        confirmation: "Cold Storage",
      }),
    );

    expect(state).toMatchObject({
      status: "error",
      title: "Permission denied",
    });
    expect(getNodeClassById).not.toHaveBeenCalled();
    expect(deleteNodeClass).not.toHaveBeenCalled();
  });

  it("deletes with remove-only permission without making a read request", async () => {
    permit("node_class:remove");
    vi.mocked(getNodeClassById).mockRejectedValue(
      new Error("GET requires node_class:get"),
    );

    await deleteNodeClassAction(
      EMPTY_STATE,
      formData({
        node_class_id: "class-1",
        node_class_name: "Cold Storage",
        confirmation: "Cold Storage",
      }),
    );

    expect(getNodeClassById).not.toHaveBeenCalled();
    expect(deleteNodeClass).toHaveBeenCalledWith("class-1");
    expect(redirect).toHaveBeenCalledWith("/node-classes");
    expect(vi.mocked(deleteNodeClass).mock.invocationCallOrder[0]).toBeLessThan(
      vi.mocked(redirect).mock.invocationCallOrder[0],
    );
  });

  it("does not redirect when the backend rejects a dependent deletion", async () => {
    permit("node_class:remove");
    vi.mocked(deleteNodeClass).mockRejectedValue(
      new Error("dependent resources"),
    );

    const state = await deleteNodeClassAction(
      EMPTY_STATE,
      formData({
        node_class_id: "class-1",
        node_class_name: "Cold Storage",
        confirmation: "Cold Storage",
      }),
    );

    expect(state.status).toBe("error");
    expect(redirect).not.toHaveBeenCalled();
  });
});
