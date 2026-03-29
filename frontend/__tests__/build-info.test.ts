import { describe, expect, it } from "vitest";

import { resolveBuildInfo } from "@/lib/build-info";

describe("resolveBuildInfo", () => {
  it("returns explicit build metadata when provided", () => {
    expect(
      resolveBuildInfo({
        NEXT_PUBLIC_BUILD_RELEASE: "v0.1.22",
        NEXT_PUBLIC_BUILD_BRANCH: "master",
        NEXT_PUBLIC_BUILD_COMMIT: "1234567890abcdef"
      })
    ).toEqual({
      release: "v0.1.22",
      branch: "master",
      commit: "1234567890abcdef",
      shortCommit: "1234567890ab"
    });
  });

  it("falls back to local development metadata when values are missing", () => {
    expect(resolveBuildInfo({})).toEqual({
      release: "local-dev",
      branch: "workspace",
      commit: "uncommitted",
      shortCommit: "uncommitted"
    });
  });
});
