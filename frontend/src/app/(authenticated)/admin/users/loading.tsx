export default function Loading() {
  return (
    <main className="mx-auto max-w-7xl animate-pulse space-y-5 p-6">
      <div className="bg-muted h-12 rounded-xl" />
      <div className="grid gap-5 sm:grid-cols-2 xl:grid-cols-3">
        {[1, 2, 3, 4, 5, 6].map((item) => (
          <div key={item} className="bg-muted h-64 rounded-2xl" />
        ))}
      </div>
    </main>
  );
}
