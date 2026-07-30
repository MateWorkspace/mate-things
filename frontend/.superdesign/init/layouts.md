# Shared layouts

The current app only has a root layout. The authenticated app shell, app bar, sidebar, profile dialog, breadcrumb, and mobile drawer are missing.

## `src/app/layout.tsx`

Root HTML layout. It loads Inter for body copy, Anton for display copy, and installs the global toast provider.

```tsx
import type { Metadata } from "next";
import { Anton, Inter } from "next/font/google";

import ToastProvider from "@/components/ui/toast-provider";

import "./globals.css";

const inter = Inter({
  subsets: ["latin"],
  variable: "--font-inter",
});

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
      <body className="flex min-h-full flex-col">
        <ToastProvider>{children}</ToastProvider>
      </body>
    </html>
  );
}
```
