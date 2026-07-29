import type { Metadata } from "next";
import Image from "next/image";
import { Suspense } from "react";

import mark from "@/assets/matethings-horizontal.svg";

import LoginForm from "./_components/LoginForm";
import SignedOutNotice from "./_components/SignedOutNotice";

export const metadata: Metadata = {
  title: "Sign in — Mate Things",
};

export default function LoginPage() {
  return (
    <div className="flex min-h-screen flex-col lg:flex-row">
      <div
        className="bg-background flex min-h-60 flex-col justify-center gap-6 px-6 py-10 sm:min-h-80 sm:px-10 sm:py-12 lg:min-h-screen lg:w-[46%] lg:px-16"
        style={{
          backgroundImage:
            "linear-gradient(color-mix(in srgb, var(--color-primary) 12%, transparent) 1px, transparent 1px), linear-gradient(90deg, color-mix(in srgb, var(--color-primary) 12%, transparent) 1px, transparent 1px)",
          backgroundSize: "40px 40px",
        }}
      >
        <Image
          src={mark}
          alt="Mate Things"
          priority
          className="h-auto w-48 sm:w-56 lg:w-72"
        />
        <p className="text-primary/80 max-w-xs text-base sm:max-w-sm sm:text-lg lg:text-xl">
          Fleet control for every MATE device
        </p>
      </div>

      <div className="bg-background flex flex-1 items-center justify-center px-6 py-12 sm:px-10">
        <div className="w-full max-w-sm">
          <p className="font-display text-accent text-sm tracking-[0.2em] uppercase">
            Mate Things
          </p>
          <h1 className="font-display text-primary mt-2 text-4xl tracking-wide sm:text-5xl">
            Sign in
          </h1>

          <div className="mt-8">
            <LoginForm />
          </div>

          <Suspense fallback={null}>
            <SignedOutNotice />
          </Suspense>
        </div>
      </div>
    </div>
  );
}
