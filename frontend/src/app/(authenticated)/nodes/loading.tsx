import PageHeader from "@/components/ui/page-header";
import { LoadingState } from "@/components/ui/states";

const CARD_SKELETONS = ["node-1", "node-2", "node-3", "node-4"] as const;

export default function Loading() {
  return (
    <main className="mx-auto w-full max-w-7xl space-y-6 px-4 py-6 sm:px-6 lg:px-8">
      <PageHeader
        title="Nodes fleet"
        description="Loading self-registered devices and fleet filters."
      />
      <LoadingState title="Loading node filters" />
      <div className="grid grid-cols-1 gap-5 sm:grid-cols-2 xl:grid-cols-3 2xl:grid-cols-4">
        {CARD_SKELETONS.map((card) => (
          <LoadingState key={card} title="Loading node" />
        ))}
      </div>
    </main>
  );
}
