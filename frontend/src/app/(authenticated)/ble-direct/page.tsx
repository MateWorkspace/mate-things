import type { Metadata } from "next";

import PageHeader from "@/components/ui/page-header";
import { requireSessionContext } from "@/lib/session";

import BleDirectRoot from "./_components/BleDirectRoot";

export const metadata: Metadata = { title: "BLE Direct — Mate Things" };

export default async function BleDirectPage() {
  await requireSessionContext();

  return (
    <main className="mx-auto w-full max-w-5xl space-y-8 px-4 py-6 sm:px-6 lg:px-8">
      <PageHeader
        title="BLE Direct"
        description="Connect directly to a device over Bluetooth to inspect its system info, manage WiFi, edit settings, and view a live log — no backend involved."
      />
      <BleDirectRoot />
    </main>
  );
}
