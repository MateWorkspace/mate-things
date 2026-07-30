import Link from "next/link";
export default function NotFound() {
  return (
    <main className="mx-auto max-w-3xl p-6 text-center">
      <h1 className="font-display text-primary text-3xl">Schema not found</h1>
      <p className="text-muted-foreground my-4">
        This payload schema version no longer exists.
      </p>
      <Link
        href="/admin/payload-schemas"
        className="text-primary font-semibold underline"
      >
        Return to payload schemas
      </Link>
    </main>
  );
}
