import type { Metadata } from "next";
import Link from "next/link";
import { notFound } from "next/navigation";
import { cache } from "react";

import PreferencesDialog from "@/components/preferences/PreferencesDialog";
import Card from "@/components/ui/card";
import PageHeader from "@/components/ui/page-header";
import StatusBadge from "@/components/ui/status-badge";
import { ApiError } from "@/lib/api/client";
import {
  getFirmwareBinaryUrlById,
  getFirmwareById,
  getFirmwareConfigParameters,
} from "@/lib/api/firmwares";
import { listAllNodeClasses } from "@/lib/api/node-classes";
import { listNodes } from "@/lib/api/nodes";
import { requirePermission } from "@/lib/session";

import FirmwareForm from "../_components/FirmwareForm";
import { formatBytes } from "../_lib/format";

interface FirmwareDetailPageProps {
  params: Promise<{ id: string }>;
}

const getFirmware = cache(getFirmwareById);

export async function generateMetadata({
  params,
}: FirmwareDetailPageProps): Promise<Metadata> {
  const { id } = await params;

  try {
    const firmware = await getFirmware(id);
    return {
      title: `${firmware.name} — Mate Things`,
      description:
        "Firmware binary, compatibility, configuration schema, and fleet assignment details.",
    };
  } catch {
    return { title: "Firmware details — Mate Things" };
  }
}

const DATE_FORMATTER = new Intl.DateTimeFormat("en", {
  dateStyle: "medium",
  timeStyle: "short",
  timeZone: "UTC",
});

export default async function FirmwareDetailPage({
  params,
}: FirmwareDetailPageProps) {
  const [{ id }, { permissions }] = await Promise.all([
    params,
    requirePermission("firmware:get"),
  ]);

  let firmware;
  try {
    firmware = await getFirmware(id);
  } catch (error) {
    if (error instanceof ApiError && error.status === 404) {
      notFound();
    }
    throw error;
  }

  const canReadClass = permissions.has("node_class:get");
  const canReadNodes = permissions.has("node:get");
  const [configSchema, nodeClasses, nodes, downloadUrl] = await Promise.all([
    getFirmwareConfigParameters(firmware.id),
    canReadClass ? listAllNodeClasses() : Promise.resolve(null),
    canReadNodes
      ? listNodes({ firmware_id: firmware.id, limit: 12 })
      : Promise.resolve(null),
    firmware.binary_path && firmware.size > 0
      ? getFirmwareBinaryUrlById(firmware.id).catch(() => null)
      : Promise.resolve(null),
  ]);
  const nodeClassOptions =
    nodeClasses?.map(({ id: nodeClassId, name }) => ({
      id: nodeClassId,
      name,
    })) ?? [];
  const nodeClass = nodeClasses?.find(
    (item) => item.id === firmware.node_class_id,
  );
  const freshness = firmware.updated_at ?? firmware.created_at;
  const hasBinary = Boolean(downloadUrl);

  return (
    <main className="mx-auto w-full max-w-7xl space-y-8 px-4 py-6 sm:px-6 lg:px-8">
      <PageHeader
        title={firmware.name}
        description="Firmware binary, compatibility, configuration schema, and lifecycle details."
        actions={
          permissions.has("firmware:set") ||
          permissions.has("firmware:remove") ||
          permissions.has("preferences:set") ? (
            <div className="flex flex-wrap gap-2">
              <FirmwareForm
                firmware={firmware}
                configSchema={configSchema}
                nodeClasses={nodeClassOptions}
                canEdit={permissions.has("firmware:set")}
                canReplace={permissions.has("firmware:set")}
                canDelete={permissions.has("firmware:remove")}
              />
              <PreferencesDialog
                resource="firmware"
                id={firmware.id}
                preferences={firmware.preferences}
                permissions={[...permissions]}
              />
            </div>
          ) : undefined
        }
      />

      <section className="grid gap-5 lg:grid-cols-2">
        <Card>
          <div className="flex flex-wrap items-start justify-between gap-3">
            <h2 className="font-display text-primary text-xl tracking-wide">
              Binary
            </h2>
            <StatusBadge variant={hasBinary ? "success" : "warning"}>
              {hasBinary ? "Available" : "Unavailable"}
            </StatusBadge>
          </div>
          <dl className="mt-4 space-y-3 text-sm">
            <div>
              <dt className="text-muted-foreground">Size</dt>
              <dd className="mt-1 font-semibold">
                {formatBytes(firmware.size)}
              </dd>
            </div>
            <div>
              <dt className="text-muted-foreground">Checksum</dt>
              <dd className="mt-1 font-mono break-all">
                {firmware.checksum || "Not reported"}
              </dd>
            </div>
            <div>
              <dt className="text-muted-foreground">Storage path</dt>
              <dd className="mt-1 font-mono break-all">
                {firmware.binary_path || "Not reported"}
              </dd>
            </div>
          </dl>
          {downloadUrl ? (
            <a
              href={downloadUrl}
              rel="noreferrer"
              className="bg-primary text-surface focus-visible:ring-focus focus-visible:ring-offset-background mt-5 inline-flex min-h-11 items-center justify-center rounded-xl px-4 py-2.5 text-sm font-semibold focus-visible:ring-2 focus-visible:ring-offset-2 focus-visible:outline-none"
            >
              Download binary
            </a>
          ) : null}
        </Card>

        <Card>
          <h2 className="font-display text-primary text-xl tracking-wide">
            Compatibility and audit
          </h2>
          <dl className="mt-4 space-y-3 text-sm">
            <div>
              <dt className="text-muted-foreground">Node class</dt>
              <dd className="mt-1 font-semibold">
                {nodeClass ? (
                  <Link
                    href={`/node-classes/${nodeClass.id}`}
                    className="text-primary focus-visible:ring-focus rounded underline-offset-4 hover:underline focus-visible:ring-2 focus-visible:outline-none"
                  >
                    {nodeClass.name}
                  </Link>
                ) : (
                  <span className="font-mono">{firmware.node_class_id}</span>
                )}
              </dd>
            </div>
            <div>
              <dt className="text-muted-foreground">Firmware ID</dt>
              <dd className="mt-1 font-mono break-all">{firmware.id}</dd>
            </div>
            <div>
              <dt className="text-muted-foreground">
                {firmware.updated_at ? "Updated" : "Created"}
              </dt>
              <dd className="mt-1">
                <time dateTime={freshness}>
                  {DATE_FORMATTER.format(new Date(freshness))} UTC
                </time>
              </dd>
            </div>
            {firmware.created_by ? (
              <div>
                <dt className="text-muted-foreground">Created by</dt>
                <dd className="mt-1 font-mono break-all">
                  {firmware.created_by}
                </dd>
              </div>
            ) : null}
          </dl>
        </Card>
      </section>

      <section aria-labelledby="firmware-schema-heading">
        <h2
          id="firmware-schema-heading"
          className="font-display text-primary text-2xl tracking-wide"
        >
          Configuration schema
        </h2>
        {configSchema.length > 0 ? (
          <div className="mt-4 grid gap-3 sm:grid-cols-2 lg:grid-cols-3">
            {configSchema.map((parameter) => (
              <Card key={parameter.key} className="p-4">
                <h3 className="font-mono font-semibold break-all">
                  {parameter.key}
                </h3>
                <p className="text-muted-foreground mt-2 text-sm">
                  {parameter.value_type}
                </p>
              </Card>
            ))}
          </div>
        ) : (
          <p className="border-border bg-muted mt-4 rounded-xl border border-dashed p-4 text-sm">
            This firmware does not declare configuration parameters.
          </p>
        )}
      </section>

      <Card>
        <h2 className="font-display text-primary text-xl tracking-wide">
          Preferences
        </h2>
        <pre className="border-border bg-muted mt-4 overflow-x-auto rounded-xl border p-4 text-sm">
          {JSON.stringify(firmware.preferences, null, 2)}
        </pre>
      </Card>

      {nodes ? (
        <section aria-labelledby="firmware-nodes-heading">
          <div className="flex flex-wrap items-end justify-between gap-3">
            <div>
              <h2
                id="firmware-nodes-heading"
                className="font-display text-primary text-2xl tracking-wide"
              >
                Assigned nodes
              </h2>
              <p className="text-muted-foreground mt-1 text-sm">
                {nodes.page.total_items}{" "}
                {nodes.page.total_items === 1 ? "node" : "nodes"} currently
                report this firmware.
              </p>
            </div>
            <Link
              href={`/nodes?firmware_id=${encodeURIComponent(firmware.id)}`}
              className="text-primary focus-visible:ring-focus rounded-lg text-sm font-semibold underline-offset-4 hover:underline focus-visible:ring-2 focus-visible:outline-none"
            >
              View all nodes
            </Link>
          </div>
          {nodes.data.length > 0 ? (
            <div className="mt-4 grid gap-3 sm:grid-cols-2 xl:grid-cols-3">
              {nodes.data.map((node) => (
                <Card key={node.id} className="p-4">
                  <h3 className="font-semibold">{node.name}</h3>
                  <p className="text-muted-foreground mt-1 font-mono text-sm">
                    {node.device_id}
                  </p>
                  <Link
                    href={`/nodes/${node.id}?tab=firmware`}
                    className="text-primary focus-visible:ring-focus mt-3 inline-block rounded text-sm font-semibold underline-offset-4 hover:underline focus-visible:ring-2 focus-visible:outline-none"
                  >
                    Open firmware workspace
                  </Link>
                </Card>
              ))}
            </div>
          ) : (
            <p className="border-border bg-muted mt-4 rounded-xl border border-dashed p-4 text-sm">
              No nodes currently report this firmware.
            </p>
          )}
        </section>
      ) : null}
    </main>
  );
}
