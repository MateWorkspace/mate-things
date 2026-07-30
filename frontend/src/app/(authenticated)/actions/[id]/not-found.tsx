import { EmptyState } from "@/components/ui/states";

export default function NotFound() {
  return (
    <main className="mx-auto max-w-5xl p-6">
      <EmptyState
        title="Action not found"
        description="This action may have been deleted or the link is incorrect."
      />
    </main>
  );
}
