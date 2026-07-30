import PageHeader from "@/components/ui/page-header";
import { LoadingState } from "@/components/ui/states";

export default function FirmwareDetailLoading() {
  return (
    <main className="mx-auto w-full max-w-7xl space-y-6 px-4 py-6 sm:px-6 lg:px-8">
      <PageHeader
        title="Firmware details"
        description="Loading binary, compatibility, and lifecycle data."
      />
      <LoadingState title="Loading firmware details" />
      <LoadingState title="Loading configuration schema" />
    </main>
  );
}
