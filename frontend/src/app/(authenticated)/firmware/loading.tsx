import PageHeader from "@/components/ui/page-header";
import { LoadingState } from "@/components/ui/states";

const FIRMWARE_SKELETONS = ["firmware-1", "firmware-2", "firmware-3"] as const;

export default function FirmwareLoading() {
  return (
    <main className="mx-auto w-full max-w-7xl space-y-6 px-4 py-6 sm:px-6 lg:px-8">
      <PageHeader
        title="Firmware"
        description="Loading firmware binaries and compatibility filters."
      />
      <LoadingState title="Loading firmware filters" />
      <div className="grid grid-cols-1 gap-5 sm:grid-cols-2 xl:grid-cols-3">
        {FIRMWARE_SKELETONS.map((firmware) => (
          <LoadingState key={firmware} title="Loading firmware record" />
        ))}
      </div>
    </main>
  );
}
