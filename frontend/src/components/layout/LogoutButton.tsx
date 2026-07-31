import Button from "@/components/ui/button";
import { logoutAction } from "@/lib/api/session-actions";

export default function LogoutButton() {
  return (
    <form action={logoutAction}>
      <Button type="submit" variant="critical">
        Log out
      </Button>
    </form>
  );
}
