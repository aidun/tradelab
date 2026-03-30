import { afterEach, describe, expect, it, vi } from "vitest";

import { resolveApiErrorMessage } from "@/lib/api";

describe("resolveApiErrorMessage", () => {
  afterEach(() => {
    vi.restoreAllMocks();
  });

  it("keeps detailed backend messages in development mode", () => {
    expect(
      resolveApiErrorMessage(
        "create order: ERROR: column \"strategy_id\" is of type uuid but expression is of type text",
        "We couldn't place that order right now.",
        "development"
      )
    ).toContain("column \"strategy_id\"");
  });

  it("returns the friendly fallback and logs detail in production mode", () => {
    const errorSpy = vi.spyOn(console, "error").mockImplementation(() => undefined);

    expect(
      resolveApiErrorMessage(
        "create order: ERROR: column \"strategy_id\" is of type uuid but expression is of type text",
        "We couldn't place that order right now.",
        "production"
      )
    ).toBe("We couldn't place that order right now.");
    expect(errorSpy).toHaveBeenCalledWith(
      "TradeLab API error",
      "create order: ERROR: column \"strategy_id\" is of type uuid but expression is of type text"
    );
  });

  it("falls back cleanly when the backend does not send a detail string", () => {
    expect(resolveApiErrorMessage(undefined, "We couldn't load activity right now.", "production")).toBe(
      "We couldn't load activity right now."
    );
  });
});
