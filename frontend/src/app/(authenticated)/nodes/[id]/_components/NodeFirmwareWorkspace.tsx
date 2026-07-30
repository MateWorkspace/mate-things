import Link from "next/link";

import Pagination from "@/components/collection/Pagination";
import Card from "@/components/ui/card";
import StatusBadge from "@/components/ui/status-badge";
import type { PageDataResponse } from "@/lib/api/types";
import type { FirmwareResponse } from "@/lib/api/firmwares";
import type { NodeResponse } from "@/lib/api/nodes";

import OtaDialog from "../../../firmware/_components/OtaDialog";

interface NodeFirmwareWorkspaceProps {
  availableFirmwares: PageDataResponse<FirmwareResponse>;
  canDispatch: boolean;
  node: NodeResponse;
}

export default function NodeFirmwareWorkspace({
  availableFirmwares,
  canDispatch,
  node,
}: NodeFirmwareWorkspaceProps) {
  return (
    <section aria-labelledby="node-firmware-heading" className="space-y-5">
      <div className="flex flex-wrap items-start justify-between gap-4">
        <div>
          <h2
            id="node-firmware-heading"
            className="font-display text-primary text-2xl tracking-wide"
          >
            Firmware & OTA
          </h2>
          <p className="text-muted-foreground mt-1 text-sm">
            Only firmware the backend reports as compatible with this node is
            shown.
          </p>
        </div>
        {canDispatch ? (
          <OtaDialog availableFirmwares={availableFirmwares.data} node={node} />
        ) : null}
      </div>

      <Card className="p-4">
        <dl className="grid gap-3 text-sm sm:grid-cols-2">
          <div>
            <dt className="text-muted-foreground">Current firmware ID</dt>
            <dd className="mt-1 font-mono break-all">
              {node.firmware_id || "Unassigned"}
            </dd>
          </div>
          <div>
            <dt className="text-muted-foreground">Node state</dt>
            <dd className="mt-1">
              <StatusBadge variant={node.is_connected ? "success" : "critical"}>
                {node.is_connected ? "Connected" : "Disconnected"}
              </StatusBadge>
            </dd>
          </div>
        </dl>
      </Card>

      {availableFirmwares.data.length > 0 ? (
        <div className="grid gap-3 sm:grid-cols-2 xl:grid-cols-3">
          {availableFirmwares.data.map((firmware) => (
            <Card key={firmware.id} className="p-4">
              <div className="flex flex-wrap items-start justify-between gap-2">
                <h3 className="font-semibold">{firmware.name}</h3>
                {firmware.id === node.firmware_id ? (
                  <StatusBadge variant="info">Current</StatusBadge>
                ) : null}
              </div>
              <p className="text-muted-foreground mt-2 font-mono text-xs break-all">
                {firmware.checksum}
              </p>
              <Link
                href={`/firmware/${firmware.id}`}
                className="text-primary focus-visible:ring-focus mt-3 inline-block rounded text-sm font-semibold underline-offset-4 hover:underline focus-visible:ring-2 focus-visible:outline-none"
              >
                View firmware details
              </Link>
            </Card>
          ))}
        </div>
      ) : !canDispatch ? (
        <p className="border-border bg-muted rounded-xl border border-dashed p-4 text-sm">
          No compatible firmware is currently available for this node.
        </p>
      ) : null}

      <div className="border-border border-t pt-4">
        <Pagination
          page={availableFirmwares.page}
          pathname={`/nodes/${node.id}`}
          searchParams={{
            tab: "firmware",
            limit: String(availableFirmwares.page.limit),
          }}
        />
      </div>
    </section>
  );
}
