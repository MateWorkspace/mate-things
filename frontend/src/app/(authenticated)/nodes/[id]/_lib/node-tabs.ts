export const NODE_TABS = [
  { id: "overview", label: "Overview" },
  { id: "configuration", label: "Configuration" },
  { id: "firmware", label: "Firmware & OTA" },
  { id: "actions", label: "Actions" },
  { id: "telemetry", label: "Telemetry" },
  { id: "logs", label: "Node Logs" },
] as const;

export type NodeTab = (typeof NODE_TABS)[number]["id"];

function isNodeTab(value: string): value is NodeTab {
  return NODE_TABS.some((tab) => tab.id === value);
}

export function normalizeNodeTab(
  value: string | string[] | undefined,
): NodeTab {
  const candidate = Array.isArray(value) ? value[0] : value;
  return candidate && isNodeTab(candidate) ? candidate : "overview";
}

export function getNodeTabPanelId(nodeId: string, tab: NodeTab): string {
  return `node-${nodeId}-${tab}-panel`;
}

export function getNodeTabButtonId(nodeId: string, tab: NodeTab): string {
  return `tab-${getNodeTabPanelId(nodeId, tab)}`;
}
