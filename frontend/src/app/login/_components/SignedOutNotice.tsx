"use client";

import { useRouter, useSearchParams } from "next/navigation";
import { useEffect, useRef } from "react";

import { useToast } from "@/hooks/use-toast";

// Renders nothing - fires a one-shot success toast when arriving from
// logout, then strips the query param so a page refresh doesn't re-fire it.
export default function SignedOutNotice() {
  const searchParams = useSearchParams();
  const router = useRouter();
  const toast = useToast();
  const fired = useRef(false);

  useEffect(() => {
    if (fired.current || searchParams.get("signedOut") !== "1") {
      return;
    }
    fired.current = true;
    toast.success("Signed out", "You've been signed out successfully.");
    router.replace("/login");
  }, [searchParams, router, toast]);

  return null;
}
