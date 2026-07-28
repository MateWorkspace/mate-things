import type { Metadata } from "next";
import Image from "next/image";

import mark from "@/assets/vertical.svg";

import LoginForm from "./_components/LoginForm";

export const metadata: Metadata = {
  title: "Sign in — Mate Things",
};

export default function LoginPage() {
  return (
    <div className="flex min-h-screen flex-col lg:flex-row">
      <div className="relative isolate flex min-h-[240px] items-end overflow-hidden bg-surface px-6 py-8 sm:min-h-[320px] sm:px-10 lg:min-h-screen lg:w-[46%] lg:items-center lg:px-16">
        <Image
          src={mark}
          alt=""
          priority
          className="pointer-events-none absolute -bottom-14 -left-14 h-[300px] w-[300px] select-none sm:h-[380px] sm:w-[380px] lg:-bottom-20 lg:-left-20 lg:h-[520px] lg:w-[520px]"
        />
        <div className="relative z-10 max-w-xs sm:max-w-sm">
          <p className="text-base text-primary/80 sm:text-lg lg:text-xl">
            Fleet control for every Mate device.
          </p>
        </div>
      </div>

      <div className="flex flex-1 items-center justify-center bg-background px-6 py-12 sm:px-10">
        <div className="w-full max-w-sm">
          <p className="font-display text-sm tracking-[0.2em] text-accent uppercase">
            Mate Things
          </p>
          <h1 className="mt-2 font-display text-4xl tracking-wide text-primary sm:text-5xl">
            Sign in
          </h1>
          <p className="mt-3 text-sm text-foreground/70">
            Enter your operator credentials to access the fleet console.
          </p>

          <div className="mt-8">
            <LoginForm />
          </div>
        </div>
      </div>
    </div>
  );
}
