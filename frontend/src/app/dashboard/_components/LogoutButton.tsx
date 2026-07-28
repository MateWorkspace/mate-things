import Button from "@/components/ui/button";

import { logoutAction } from "../_lib/actions";

// A plain <form action> - genuinely no client JS needed for a logout
// button (no pending-state UI required), so this stays a Server Component.
export default function LogoutButton() {
  return (
    <form action={logoutAction}>
      <Button type="submit" variant="secondary">
        Log out
      </Button>
    </form>
  );
}
