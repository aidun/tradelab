export type BuildInfo = {
  release: string;
  branch: string;
  commit: string;
  shortCommit: string;
};

function cleanValue(value: string | undefined, fallback: string) {
  const trimmed = value?.trim();
  return trimmed ? trimmed : fallback;
}

export function resolveBuildInfo(env: NodeJS.ProcessEnv = process.env): BuildInfo {
  const release = cleanValue(env.NEXT_PUBLIC_BUILD_RELEASE, "local-dev");
  const branch = cleanValue(env.NEXT_PUBLIC_BUILD_BRANCH, "workspace");
  const commit = cleanValue(env.NEXT_PUBLIC_BUILD_COMMIT, "uncommitted");

  return {
    release,
    branch,
    commit,
    shortCommit: commit === "uncommitted" ? commit : commit.slice(0, 12)
  };
}
