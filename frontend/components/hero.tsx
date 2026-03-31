import React from "react";
import { TradeLabAppShell } from "@/components/tradelab-app-shell";

/** Hero renders the default dashboard entrypoint used on the homepage. */
export function Hero() {
  return <TradeLabAppShell requestedPath="/" />;
}
