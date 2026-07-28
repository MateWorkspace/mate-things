import Button from "@/components/ui/button";

import { logoutAction } from "../_lib/actions";

export default function LogoutButton() {
  return (
    <form action={logoutAction}>
      <Button type="submit" variant="secondary">
        Log out
      </Button>
    </form>
  );
}
