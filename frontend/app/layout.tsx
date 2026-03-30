import React from "react";
import type { Metadata } from "next";
import { Space_Grotesk, IBM_Plex_Mono } from "next/font/google";
import { TradeLabAuthProvider, type TradeLabAuthRuntimeConfig } from "@/lib/tradelab-auth";
import "./globals.css";

const displayFont = Space_Grotesk({
  subsets: ["latin"],
  variable: "--font-display"
});

const monoFont = IBM_Plex_Mono({
  subsets: ["latin"],
  weight: ["400", "500"],
  variable: "--font-mono"
});

export const metadata: Metadata = {
  title: "TradeLab",
  description: "A multi-asset paper trading platform for strategy testing and automated demo execution."
};

function resolveAuthRuntimeConfig(): TradeLabAuthRuntimeConfig {
  return {
    mockMode: process.env.NEXT_PUBLIC_AUTH_MOCK_MODE === "true",
    clerkPublishableKey: process.env.NEXT_PUBLIC_CLERK_PUBLISHABLE_KEY ?? null
  };
}

export default function RootLayout({
  children
}: Readonly<{
  children: React.ReactNode;
}>) {
  const authConfig = resolveAuthRuntimeConfig();

  return (
    <html lang="en">
      <body className={`${displayFont.variable} ${monoFont.variable}`}>
        <TradeLabAuthProvider config={authConfig}>{children}</TradeLabAuthProvider>
      </body>
    </html>
  );
}
