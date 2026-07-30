import PageHeader from "@/components/ui/page-header";
import { LoadingState } from "@/components/ui/states";

export default function AuthenticatedLoading() {
  return (
    <main
      className="mx-auto w-full max-w-7xl space-y-6 px-4 py-6 sm:px-6 lg:px-8"
      aria-busy="true"
    >
      <PageHeader
        title="Loading workspace"
        description="Retrieving the latest authorized data."
      />
      <div className="grid gap-4 sm:grid-cols-2 xl:grid-cols-3">
        <LoadingState title="Loading workspace summary" />
        <LoadingState title="Loading workspace records" />
        <LoadingState title="Loading workspace controls" />
      </div>
    </main>
  );
}
