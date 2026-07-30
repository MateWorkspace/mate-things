import Link from "next/link";

export default function NotFound() {
  return (
    <main className="mx-auto max-w-3xl p-6 text-center">
      <h1 className="font-display text-primary text-3xl">User not found</h1>
      <p className="text-muted-foreground my-4">
        This account no longer exists or is unavailable.
      </p>
      <Link
        className="text-primary font-semibold underline"
        href="/admin/users"
      >
        Return to users
      </Link>
    </main>
  );
}
