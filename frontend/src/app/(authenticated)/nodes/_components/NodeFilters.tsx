"use client";

import { Search, X } from "lucide-react";
import Form from "next/form";
import Link from "next/link";

import Button from "@/components/ui/button";
import Input from "@/components/ui/input";

export type ConnectionFilter = "all" | "connected" | "disconnected";

interface FilterOption {
  id: string;
  name: string;
}

interface NodeFiltersProps {
  classes: readonly FilterOption[];
  connection: ConnectionFilter;
  firmwares: readonly FilterOption[];
  limit: number;
  nodeClassId?: string;
  firmwareId?: string;
  search?: string;
  showClassFilter: boolean;
  showFirmwareFilter: boolean;
}

const SELECT_CLASS_NAME =
  "border-control-border bg-background text-foreground focus-visible:border-focus focus-visible:ring-focus min-h-11 w-full rounded-xl border px-3.5 py-2.5 text-sm transition-colors focus-visible:ring-2 focus-visible:outline-none";

export default function NodeFilters({
  classes,
  connection,
  firmwareId,
  firmwares,
  limit,
  nodeClassId,
  search,
  showClassFilter,
  showFirmwareFilter,
}: NodeFiltersProps) {
  return (
    <Form action="/nodes" className="grid gap-4 lg:grid-cols-12 lg:items-end">
      <input name="page" type="hidden" value="1" />
      <input name="limit" type="hidden" value={limit} />

      <label className="lg:col-span-4">
        <span className="text-foreground/70 mb-1.5 block text-xs font-semibold">
          Search
        </span>
        <span className="relative block">
          <Search
            aria-hidden="true"
            className="text-muted-foreground pointer-events-none absolute top-1/2 left-3.5 size-4 -translate-y-1/2"
          />
          <Input
            className="pl-10"
            defaultValue={search}
            name="search"
            placeholder="Name or device ID"
            type="search"
          />
        </span>
      </label>

      {showClassFilter ? (
        <label className="lg:col-span-2">
          <span className="text-foreground/70 mb-1.5 block text-xs font-semibold">
            Node class
          </span>
          <select
            className={SELECT_CLASS_NAME}
            defaultValue={nodeClassId ?? ""}
            name="node_class_id"
          >
            <option value="">All classes</option>
            {classes.map((nodeClass) => (
              <option key={nodeClass.id} value={nodeClass.id}>
                {nodeClass.name}
              </option>
            ))}
          </select>
        </label>
      ) : null}

      {showFirmwareFilter ? (
        <label className="lg:col-span-2">
          <span className="text-foreground/70 mb-1.5 block text-xs font-semibold">
            Firmware
          </span>
          <select
            className={SELECT_CLASS_NAME}
            defaultValue={firmwareId ?? ""}
            name="firmware_id"
          >
            <option value="">All firmware</option>
            {firmwares.map((firmware) => (
              <option key={firmware.id} value={firmware.id}>
                {firmware.name}
              </option>
            ))}
          </select>
        </label>
      ) : null}

      <label className="lg:col-span-2">
        <span className="text-foreground/70 mb-1.5 flex items-center gap-2 text-xs font-semibold">
          Connection
          <span className="bg-muted text-muted-foreground rounded-full px-2 py-0.5 text-[0.65rem] uppercase">
            This page
          </span>
        </span>
        <select
          className={SELECT_CLASS_NAME}
          defaultValue={connection}
          name="connection"
        >
          <option value="all">All states</option>
          <option value="connected">Connected</option>
          <option value="disconnected">Disconnected</option>
        </select>
      </label>

      <div className="flex gap-2 lg:col-span-2">
        <Button className="min-h-11 flex-1" type="submit">
          Apply
        </Button>
        <Link
          aria-label="Clear node filters"
          className="border-border text-primary hover:bg-highlight/40 focus-visible:ring-focus focus-visible:ring-offset-background inline-flex size-11 shrink-0 items-center justify-center rounded-xl border transition-colors focus-visible:ring-2 focus-visible:ring-offset-2 focus-visible:outline-none"
          href={`/nodes?limit=${limit}`}
        >
          <X aria-hidden="true" className="size-4" />
        </Link>
      </div>
    </Form>
  );
}
