import { LoadingState } from "@/components/ui/states";

export default function Loading() {
  return (
    <main className="mx-auto max-w-7xl space-y-4 p-6">
      <LoadingState title="Loading actions" />
      <LoadingState />
      <LoadingState />
    </main>
  );
}
