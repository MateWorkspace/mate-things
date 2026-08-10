import "server-only";

import { listActionLogs, type ActionLogResponse } from "@/lib/api/action-logs";
import { listNodeLogs, type NodeLogResponse } from "@/lib/api/node-logs";
import { listNodes, type NodeResponse } from "@/lib/api/nodes";

const NODE_SAMPLE_SIZE = 48;
const RECENT_WINDOW_HOURS = 6;

export interface DashboardData {
  nodes?: {
    /** Exact backend collection total. */
    total: number;
    /** Number of nodes inspected in the bounded first-page sample. */
    sampled: number;
    /** Connected count within the bounded sample, not the full fleet. */
    connected: number;
    /** Disconnected nodes within the bounded sample, not the full fleet. */
    disconnected: NodeResponse[];
  };
  failedActions?: ActionLogResponse[];
  recentWarnings?: NodeLogResponse[];
  windowStartedAt: string;
  loadedAt: string;
}

function newestFirst<T>(
  items: readonly T[],
  dateFor: (item: T) => string,
): T[] {
  return [...items].sort(
    (left, right) =>
      new Date(dateFor(right)).getTime() - new Date(dateFor(left)).getTime(),
  );
}

export async function loadDashboardData(
  permissions: ReadonlySet<string>,
  now = new Date(),
): Promise<DashboardData> {
  const windowStartedAt = new Date(
    now.getTime() - RECENT_WINDOW_HOURS * 60 * 60 * 1_000,
  ).toISOString();

  const nodeRequest = permissions.has("node:get")
    ? listNodes({ page: 1, limit: NODE_SAMPLE_SIZE }).then((response) => ({
        total: response.page.total_items,
        sampled: response.data.length,
        connected: response.data.filter((node) => node.is_connected).length,
        disconnected: response.data.filter((node) => !node.is_connected),
      }))
    : Promise.resolve(undefined);

  const actionRequest = permissions.has("action_log:get")
    ? listActionLogs({ executed_at_start: windowStartedAt }).then((response) =>
        response.data.filter(
          (action) =>
            action.action_status === "FAILED" ||
            action.action_status === "UNRESPONDED",
        ),
      )
    : Promise.resolve(undefined);

  const warningRequest = permissions.has("node_log:get")
    ? Promise.all([
        listNodeLogs({ logged_at_start: windowStartedAt, level: "ERROR" }),
        listNodeLogs({ logged_at_start: windowStartedAt, level: "WARN" }),
      ]).then(([errors, warnings]) =>
        newestFirst(
          [...errors.data, ...warnings.data],
          (record) => record.logged_at,
        ),
      )
    : Promise.resolve(undefined);

  const [nodes, failedActions, recentWarnings] = await Promise.all([
    nodeRequest,
    actionRequest,
    warningRequest,
  ]);

  return {
    nodes,
    failedActions,
    recentWarnings,
    windowStartedAt,
    loadedAt: now.toISOString(),
  };
}

export const DASHBOARD_NODE_SAMPLE_SIZE = NODE_SAMPLE_SIZE;
export const DASHBOARD_RECENT_WINDOW_HOURS = RECENT_WINDOW_HOURS;
