import type { Metadata } from "next";
import { notFound } from "next/navigation";

import PreferencesDialog from "@/components/preferences/PreferencesDialog";
import PageHeader from "@/components/ui/page-header";
import { EmptyState } from "@/components/ui/states";
import { ApiError } from "@/lib/api/client";
import {
  getFirmwareConfigParameters,
  listAvailableFirmwaresByNodeId,
} from "@/lib/api/firmwares";
import { getNodeConfig } from "@/lib/api/node-config";
import { getNodeById } from "@/lib/api/nodes";
import { parsePageQuery } from "@/lib/collection-query";
import { requirePermission } from "@/lib/session";

import DeleteNodeDialog from "./_components/DeleteNodeDialog";
import NodeConfigForm from "./_components/NodeConfigForm";
import NodeEditForm from "./_components/NodeEditForm";
import NodeFirmwareWorkspace from "./_components/NodeFirmwareWorkspace";
import NodeOverview from "./_components/NodeOverview";
import NodeTabs from "./_components/NodeTabs";
import {
  getNodeTabButtonId,
  getNodeTabPanelId,
  normalizeNodeTab,
} from "./_lib/node-tabs";

export const metadata: Metadata = {
  title: "Node details — Mate Things",
};

type RawSearchParams = Record<string, string | string[] | undefined>;

interface NodeDetailPageProps {
  params: Promise<{ id: string }>;
  searchParams: Promise<RawSearchParams>;
}

function WorkspacePlaceholder({ tab }: { tab: string }) {
  const copy: Record<string, { title: string; description: string }> = {
    actions: {
      title: "Action data is not connected yet",
      description:
        "This workspace will show available actions and dispatch history when action data is connected.",
    },
    telemetry: {
      title: "Telemetry is not connected yet",
      description:
        "This workspace will show node telemetry when telemetry data is connected.",
    },
    logs: {
      title: "Node logs are not connected yet",
      description:
        "This workspace will show device log records when node-log data is connected.",
    },
  };

  const content = copy[tab];
  return <EmptyState title={content.title} description={content.description} />;
}

export default async function NodeDetailPage({
  params,
  searchParams,
}: NodeDetailPageProps) {
  const [{ id }, rawSearchParams, { permissions }] = await Promise.all([
    params,
    searchParams,
    requirePermission("node:get"),
  ]);
  const activeTab = normalizeNodeTab(rawSearchParams.tab);

  let node;
  try {
    node = await getNodeById(id);
  } catch (error) {
    if (error instanceof ApiError && error.status === 404) {
      notFound();
    }

    throw error;
  }

  const canReadConfig = permissions.has("node_config:get");
  const canReadFirmware = permissions.has("firmware:get");
  const showConfiguration = activeTab === "configuration";
  const showFirmware = activeTab === "firmware";
  const firmwareQuery = parsePageQuery(rawSearchParams);
  const canLoadConfiguration = showConfiguration && canReadConfig;
  const canLoadSchema =
    canLoadConfiguration && canReadFirmware && Boolean(node.firmware_id);
  const [values, parameters, availableFirmwares] = await Promise.all([
    canLoadConfiguration ? getNodeConfig(node.id) : Promise.resolve([]),
    canLoadSchema
      ? getFirmwareConfigParameters(node.firmware_id)
      : Promise.resolve([]),
    showFirmware && canReadFirmware
      ? listAvailableFirmwaresByNodeId(node.id, firmwareQuery)
      : Promise.resolve(null),
  ]);
  const panelId = getNodeTabPanelId(node.id, activeTab);

  return (
    <main className="mx-auto w-full max-w-7xl space-y-6 px-4 py-6 sm:px-6 lg:px-8">
      <PageHeader
        title={node.name}
        description={`Device ID: ${node.device_id}`}
        actions={
          <div className="flex flex-wrap justify-end gap-2">
            {permissions.has("node:set") ? <NodeEditForm node={node} /> : null}
            {permissions.has("node:remove") ? (
              <DeleteNodeDialog
                deviceId={node.device_id}
                nodeId={node.id}
                nodeName={node.name}
              />
            ) : null}
            <PreferencesDialog
              resource="node"
              id={node.id}
              preferences={node.preferences}
              permissions={[...permissions]}
            />
          </div>
        }
      />

      <NodeTabs activeTab={activeTab} nodeId={node.id} />

      <section
        id={panelId}
        role="tabpanel"
        aria-labelledby={getNodeTabButtonId(node.id, activeTab)}
      >
        {activeTab === "overview" ? <NodeOverview node={node} /> : null}
        {activeTab === "configuration" && !canReadConfig ? (
          <EmptyState
            title="Configuration access required"
            description="node_config:get permission is required to view this node's configuration values."
          />
        ) : null}
        {activeTab === "configuration" && canReadConfig && !canReadFirmware ? (
          <EmptyState
            title="Firmware schema access required"
            description="firmware:get permission is required to load the configuration schema for this node's current firmware."
          />
        ) : null}
        {activeTab === "configuration" &&
        canReadConfig &&
        canReadFirmware &&
        !node.firmware_id ? (
          <EmptyState
            title="No firmware assigned"
            description="Assign firmware before its configuration schema can be shown."
          />
        ) : null}
        {activeTab === "configuration" && canLoadSchema ? (
          <NodeConfigForm
            canSet={permissions.has("node_config:set")}
            nodeId={node.id}
            parameters={parameters}
            values={values}
          />
        ) : null}
        {showFirmware && !canReadFirmware ? (
          <EmptyState
            title="Firmware access required"
            description="firmware:get permission is required to load firmware compatible with this node."
          />
        ) : null}
        {showFirmware && availableFirmwares ? (
          <NodeFirmwareWorkspace
            availableFirmwares={availableFirmwares}
            canDispatch={permissions.has("ota:dispatch")}
            node={node}
          />
        ) : null}
        {activeTab !== "overview" &&
        activeTab !== "configuration" &&
        activeTab !== "firmware" ? (
          <WorkspacePlaceholder tab={activeTab} />
        ) : null}
      </section>
    </main>
  );
}
