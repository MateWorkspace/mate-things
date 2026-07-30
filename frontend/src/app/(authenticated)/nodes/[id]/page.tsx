import type { Metadata } from "next";
import { notFound } from "next/navigation";
import { cache } from "react";

import PreferencesDialog from "@/components/preferences/PreferencesDialog";
import PageHeader from "@/components/ui/page-header";
import { EmptyState } from "@/components/ui/states";
import { listActions } from "@/lib/api/actions";
import { ApiError } from "@/lib/api/client";
import {
  getFirmwareConfigParameters,
  listAvailableFirmwaresByNodeId,
} from "@/lib/api/firmwares";
import { getNodeConfig } from "@/lib/api/node-config";
import { listNodeLogs } from "@/lib/api/node-logs";
import { getNodeById } from "@/lib/api/nodes";
import { listTelemetryRecords } from "@/lib/api/telemetry";
import { parsePageQuery } from "@/lib/collection-query";
import { requirePermission } from "@/lib/session";

import DeleteNodeDialog from "./_components/DeleteNodeDialog";
import NodeConfigForm from "./_components/NodeConfigForm";
import NodeEditForm from "./_components/NodeEditForm";
import NodeFirmwareWorkspace from "./_components/NodeFirmwareWorkspace";
import NodeOverview from "./_components/NodeOverview";
import {
  NodeActionsWorkspace,
  NodeLogsWorkspace,
  NodeTelemetryWorkspace,
} from "./_components/NodeOperationsWorkspace";
import NodeTabs from "./_components/NodeTabs";
import {
  getNodeTabButtonId,
  getNodeTabPanelId,
  normalizeNodeTab,
} from "./_lib/node-tabs";

type RawSearchParams = Record<string, string | string[] | undefined>;

interface NodeDetailPageProps {
  params: Promise<{ id: string }>;
  searchParams: Promise<RawSearchParams>;
}

const getNode = cache(getNodeById);

export async function generateMetadata({
  params,
}: Pick<NodeDetailPageProps, "params">): Promise<Metadata> {
  const { id } = await params;

  try {
    const node = await getNode(id);
    return {
      title: `${node.name} — Mate Things`,
      description: `Fleet details and operations for node ${node.device_id}.`,
    };
  } catch {
    return { title: "Node details — Mate Things" };
  }
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
    node = await getNode(id);
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
  const showActions = activeTab === "actions";
  const showTelemetry = activeTab === "telemetry";
  const showLogs = activeTab === "logs";
  const firmwareQuery = parsePageQuery(rawSearchParams);
  const canLoadConfiguration = showConfiguration && canReadConfig;
  const canLoadSchema =
    canLoadConfiguration && canReadFirmware && Boolean(node.firmware_id);
  const [values, parameters, availableFirmwares, actions, telemetry, logs] =
    await Promise.all([
      canLoadConfiguration ? getNodeConfig(node.id) : Promise.resolve([]),
      canLoadSchema
        ? getFirmwareConfigParameters(node.firmware_id)
        : Promise.resolve([]),
      showFirmware && canReadFirmware
        ? listAvailableFirmwaresByNodeId(node.id, firmwareQuery)
        : Promise.resolve(null),
      showActions && permissions.has("action:get")
        ? listActions({
            page: 1,
            limit: 48,
            node_class_id: node.node_class_id,
          })
        : Promise.resolve(null),
      showTelemetry && permissions.has("telemetry_record:get")
        ? listTelemetryRecords({ node_device_id: node.device_id })
        : Promise.resolve(null),
      showLogs && permissions.has("node_log:get")
        ? listNodeLogs({ node_device_id: node.device_id })
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
        {showActions && !permissions.has("action:get") ? (
          <EmptyState
            title="Action access required"
            description="action:get permission is required to view actions compatible with this node."
          />
        ) : null}
        {showActions && actions ? (
          <NodeActionsWorkspace
            actions={actions.data}
            canDispatch={permissions.has("action:dispatch")}
            node={node}
          />
        ) : null}
        {showTelemetry && !permissions.has("telemetry_record:get") ? (
          <EmptyState
            title="Telemetry access required"
            description="telemetry_record:get permission is required to inspect this node's telemetry."
          />
        ) : null}
        {showTelemetry && telemetry ? (
          <NodeTelemetryWorkspace records={telemetry.data} node={node} />
        ) : null}
        {showLogs && !permissions.has("node_log:get") ? (
          <EmptyState
            title="Node log access required"
            description="node_log:get permission is required to inspect this node's logs."
          />
        ) : null}
        {showLogs && logs ? (
          <NodeLogsWorkspace logs={logs.data} node={node} />
        ) : null}
      </section>
    </main>
  );
}
