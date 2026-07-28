import type { Metadata } from "next";
import { Anton, Inter } from "next/font/google";

import ToastProvider from "@/components/ui/toast-provider";

import "./globals.css";

// Body face: humanist/geometric, deliberately neutral so the display face
// carries the brand personality - see frontend/AGENTS.md's Typography rule.
const inter = Inter({
  subsets: ["latin"],
  variable: "--font-inter",
});

// Display face: closest free substitute for the wordmark's Coolvetica-style
// bold/condensed/geometric character (no licensed Coolvetica file exists in
// this repo yet - see frontend/AGENTS.md). Anton only ships one weight.
const anton = Anton({
  subsets: ["latin"],
  weight: "400",
  variable: "--font-anton",
});

export const metadata: Metadata = {
  title: "Mate Things",
  description: "All-in-one IoT service by Mate.",
};

export default function RootLayout({
  children,
}: Readonly<{
  children: React.ReactNode;
}>) {
  return (
    <html
      lang="en"
      className={`h-full antialiased ${inter.variable} ${anton.variable}`}
    >
      <body className="min-h-full flex flex-col">
        <ToastProvider>{children}</ToastProvider>
      </body>
    </html>
  );
}
