import { cleanup, render, screen } from "@testing-library/react";
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";

import { getActionById } from "@/lib/api/actions";
import { listAllNodeClassActions } from "@/lib/api/node-classes";
import { listAllNodes } from "@/lib/api/nodes";
import { listAllPayloadSchemas } from "@/lib/api/payload-schemas";
import { requirePermission } from "@/lib/session";
import { nodeFixture, USER } from "@/test/fixtures";

import ActionDetailPage from "./page";

vi.mock("server-only", () => ({}));
vi.mock("next/navigation", () => ({ notFound: vi.fn() }));
vi.mock("@/app/(authenticated)/actions/_components/ActionForm", () => ({
  default: () => <div>action form</div>,
}));
vi.mock(
  "@/app/(authenticated)/actions/_components/DispatchActionDialog",
  () => ({
    default: ({ nodes }: { nodes: { id: string; name: string }[] }) => (
      <div>{nodes.map((node) => node.name).join(", ")}</div>
    ),
  }),
);
vi.mock("@/lib/api/actions", () => ({ getActionById: vi.fn() }));
vi.mock("@/lib/api/node-classes", () => ({
  listAllNodeClassActions: vi.fn(),
}));
vi.mock("@/lib/api/nodes", () => ({ listAllNodes: vi.fn() }));
vi.mock("@/lib/api/payload-schemas", () => ({
  listAllPayloadSchemas: vi.fn(),
}));
vi.mock("@/lib/session", () => ({ requirePermission: vi.fn() }));

const ACTION = {
  id: "action-1",
  name: "Restart",
  description: "Restart node",
  payload_schema_name: "empty",
  payload_schema_version: 1,
  preferences: {},
  created_at: "2026-08-10T00:00:00Z",
};

const NODE_CLASS = {
  id: "class-1",
  name: "Sensors",
  description: "Sensors",
  preferences: {},
  created_at: "2026-08-10T00:00:00Z",
};

describe("ActionDetailPage", () => {
  beforeEach(() => {
    vi.resetAllMocks();
    vi.mocked(getActionById).mockResolvedValue(ACTION);
    vi.mocked(listAllPayloadSchemas).mockResolvedValue([]);
    vi.mocked(listAllNodes).mockResolvedValue([
      nodeFixture({ id: "node-101", name: "Beyond first node" }),
    ]);
    vi.mocked(listAllNodeClassActions).mockResolvedValue([
      {
        node_class_action: {
          id: "assignment-1",
          node_class_id: NODE_CLASS.id,
          action_id: ACTION.id,
          created_at: "2026-08-10T00:00:00Z",
        },
        node_class: NODE_CLASS,
        action: ACTION,
      },
    ]);
  });

  afterEach(cleanup);

  it("loads every dispatch node and preserves the action compatibility filter", async () => {
    vi.mocked(requirePermission).mockResolvedValue({
      user: USER,
      permissions: new Set([
        "action:get",
        "action:dispatch",
        "node:get",
        "node_class_action:get",
      ]),
    });

    render(
      await ActionDetailPage({ params: Promise.resolve({ id: ACTION.id }) }),
    );

    expect(screen.getByText("Beyond first node")).toBeVisible();
    expect(listAllNodes).toHaveBeenCalledOnce();
    expect(listAllNodeClassActions).toHaveBeenCalledWith({
      action_id: ACTION.id,
    });
  });

  it("does not load dispatch resources without their exact permissions", async () => {
    vi.mocked(requirePermission).mockResolvedValue({
      user: USER,
      permissions: new Set(["action:get"]),
    });

    render(
      await ActionDetailPage({ params: Promise.resolve({ id: ACTION.id }) }),
    );

    expect(listAllNodes).not.toHaveBeenCalled();
    expect(listAllNodeClassActions).not.toHaveBeenCalled();
  });
});
