import Card from "@/components/ui/card";

import type { SystemInfo } from "../../_lib/ble/protocol";

function InfoRow({ label, value }: { label: string; value: string | number }) {
  return (
    <div className="flex items-start justify-between gap-4">
      <dt className="text-muted-foreground">{label}</dt>
      <dd className="text-right font-medium">{value}</dd>
    </div>
  );
}

export default function SystemInfoBlock({ info }: { info: SystemInfo }) {
  return (
    <Card>
      <h2 className="font-display text-primary text-xl tracking-wide">
        System info
      </h2>
      <div className="mt-4 grid gap-6 sm:grid-cols-2">
        <div>
          <h3 className="text-muted-foreground text-xs font-semibold tracking-wider uppercase">
            Project
          </h3>
          <dl className="mt-2 space-y-2">
            <InfoRow label="Name" value={info.project.name} />
            <InfoRow label="Project" value={info.project.project_name} />
            <InfoRow label="Version" value={info.project.project_version} />
            <InfoRow label="Type" value={info.project.type} />
            <InfoRow label="Firmware" value={info.project.firmware_version} />
          </dl>
        </div>
        <div>
          <h3 className="text-muted-foreground text-xs font-semibold tracking-wider uppercase">
            Chip
          </h3>
          <dl className="mt-2 space-y-2">
            <InfoRow label="Model" value={info.chip.model} />
            <InfoRow label="Revision" value={info.chip.revision} />
            <InfoRow label="Cores" value={info.chip.cores} />
            <InfoRow label="MAC" value={info.chip.hardware_mac} />
          </dl>
        </div>
      </div>
    </Card>
  );
}
