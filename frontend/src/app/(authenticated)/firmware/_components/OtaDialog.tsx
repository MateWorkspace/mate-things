"use client";

import { useRouter } from "next/navigation";
import { useState, useTransition } from "react";

import Button from "@/components/ui/button";
import Dialog from "@/components/ui/dialog";
import Label from "@/components/ui/label";
import type { FirmwareResponse } from "@/lib/api/firmwares";
import type { NodeResponse } from "@/lib/api/nodes";

import { dispatchOtaAction, type FormActionState } from "../_lib/actions";

interface OtaDialogProps {
  availableFirmwares: readonly FirmwareResponse[];
  node: NodeResponse;
}

const INITIAL_STATE: FormActionState = { status: "idle" };

export default function OtaDialog({
  availableFirmwares,
  node,
}: OtaDialogProps) {
  const [open, setOpen] = useState(false);
  const [firmwareId, setFirmwareId] = useState("");
  const [confirmed, setConfirmed] = useState(false);
  const [state, setState] = useState<FormActionState>(INITIAL_STATE);
  const [isPending, startTransition] = useTransition();
  const router = useRouter();

  if (availableFirmwares.length === 0) {
    return (
      <p className="border-border bg-muted rounded-xl border border-dashed p-4 text-sm">
        No compatible firmware is currently available for OTA.
      </p>
    );
  }

  const close = () => {
    if (!isPending) {
      setOpen(false);
    }
  };

  return (
    <>
      <Button
        type="button"
        aria-label={`Dispatch OTA to ${node.name}`}
        onClick={() => {
          setFirmwareId("");
          setConfirmed(false);
          setState(INITIAL_STATE);
          setOpen(true);
        }}
      >
        Dispatch OTA
      </Button>

      <Dialog
        open={open}
        onClose={close}
        title="Dispatch firmware update"
        variant="sheet"
      >
        <form
          className="space-y-5"
          onSubmit={(event) => {
            event.preventDefault();
            if (!confirmed || !firmwareId || isPending) {
              return;
            }

            startTransition(async () => {
              const result = await dispatchOtaAction({
                nodeId: node.id,
                firmwareId,
              });
              setState(result);
              if (result.status === "success") {
                setFirmwareId("");
                setConfirmed(false);
                setOpen(false);
                router.refresh();
              }
            });
          }}
        >
          <dl className="border-border bg-muted grid gap-3 rounded-xl border p-4 text-sm sm:grid-cols-2">
            <div>
              <dt className="text-muted-foreground">Node</dt>
              <dd className="mt-1 font-semibold">{node.name}</dd>
            </div>
            <div>
              <dt className="text-muted-foreground">Device ID</dt>
              <dd className="mt-1 font-mono break-all">{node.device_id}</dd>
            </div>
          </dl>

          <div>
            <Label htmlFor={`ota-firmware-${node.id}`}>Firmware</Label>
            <select
              id={`ota-firmware-${node.id}`}
              name="firmware_id"
              value={firmwareId}
              required
              className="border-control-border bg-background text-foreground focus-visible:border-focus focus-visible:ring-focus min-h-11 w-full rounded-xl border px-3.5 py-2.5 text-sm focus-visible:ring-2 focus-visible:outline-none"
              onChange={(event) => {
                setFirmwareId(event.target.value);
                setConfirmed(false);
              }}
            >
              <option value="">Select compatible firmware</option>
              {availableFirmwares.map((firmware) => (
                <option key={firmware.id} value={firmware.id}>
                  {firmware.name}
                </option>
              ))}
            </select>
          </div>

          <label className="border-border flex min-h-11 items-start gap-3 rounded-xl border p-3 text-sm">
            <input
              type="checkbox"
              checked={confirmed}
              className="accent-primary mt-0.5 size-5 shrink-0"
              onChange={(event) => setConfirmed(event.target.checked)}
            />
            <span>
              I confirm this firmware update should be dispatched to{" "}
              <strong>{node.name}</strong>.
            </span>
          </label>

          <p
            aria-live={state.status === "error" ? "assertive" : "polite"}
            className={
              state.status === "error"
                ? "text-critical text-sm"
                : "text-success text-sm"
            }
          >
            {state.message}
          </p>

          <div className="flex flex-wrap justify-end gap-2">
            <Button
              type="button"
              variant="secondary"
              disabled={isPending}
              onClick={close}
            >
              Cancel
            </Button>
            <Button
              type="submit"
              disabled={!firmwareId || !confirmed || isPending}
            >
              {isPending ? "Dispatching…" : "Confirm OTA dispatch"}
            </Button>
          </div>
        </form>
      </Dialog>
    </>
  );
}
