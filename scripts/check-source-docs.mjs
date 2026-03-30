import { readFileSync } from "node:fs";
import path from "node:path";

const repoRoot = process.cwd();

const rules = [
  {
    file: "backend/internal/config/config.go",
    kind: "go",
    symbols: ["Config", "Load"]
  },
  {
    file: "backend/internal/domain/account.go",
    kind: "go",
    symbols: ["Principal", "RegisteredAccount"]
  },
  {
    file: "backend/internal/domain/market.go",
    kind: "go",
    symbols: ["Market", "CandleFeed"]
  },
  {
    file: "backend/internal/domain/order.go",
    kind: "go",
    symbols: ["Order", "PortfolioSummary", "ActivityLog", "NormalizeAccountingMode"]
  },
  {
    file: "backend/internal/domain/strategy.go",
    kind: "go",
    symbols: ["StrategyConfig", "Strategy", "StrategyRun"]
  },
  {
    file: "backend/internal/http/router.go",
    kind: "go",
    symbols: ["MarketLister", "OrderPlacer", "StrategyManager", "NewRouter"]
  },
  {
    file: "frontend/components/hero.tsx",
    kind: "ts",
    symbols: ["Hero"]
  },
  {
    file: "frontend/components/market-dashboard.tsx",
    kind: "ts",
    symbols: ["MarketDashboard"]
  },
  {
    file: "frontend/lib/api.ts",
    kind: "ts",
    symbols: [
      "ApiError",
      "resolveApiErrorMessage",
      "createDemoSession",
      "fetchPortfolio",
      "fetchOrders",
      "fetchActivity",
      "fetchStrategies",
      "placeMarketOrder"
    ]
  },
  {
    file: "frontend/lib/build-info.ts",
    kind: "ts",
    symbols: ["resolveBuildInfo"]
  },
  {
    file: "frontend/lib/tradelab-auth.tsx",
    kind: "ts",
    symbols: ["TradeLabAuthProvider", "useTradeLabAuth", "AuthEntryActions", "AuthStatusControls"]
  },
  {
    file: "frontend/lib/use-account-session.ts",
    kind: "ts",
    symbols: ["useAccountSession"]
  }
];

const failures = [];

for (const rule of rules) {
  const absolutePath = path.join(repoRoot, rule.file);
  const source = readFileSync(absolutePath, "utf8");

  for (const symbol of rule.symbols) {
    const documented = rule.kind === "go" ? hasGoDoc(source, symbol) : hasTsDoc(source, symbol);
    if (!documented) {
      failures.push(`${rule.file}: missing source-code documentation for ${symbol}`);
    }
  }
}

if (failures.length > 0) {
  console.error("Source-code documentation check failed:");
  for (const failure of failures) {
    console.error(`- ${failure}`);
  }
  process.exit(1);
}

console.log(`Source-code documentation check passed for ${rules.length} curated public surfaces.`);

function hasGoDoc(source, symbol) {
  const pattern = new RegExp(`//\\s+${escapeRegex(symbol)}\\b[\\s\\S]*?\\n(?:type|func)\\s+${escapeRegex(symbol)}\\b`, "m");
  return pattern.test(source);
}

function hasTsDoc(source, symbol) {
  const jsDocPattern = new RegExp(`/\\*\\*[\\s\\S]*?\\*/\\s*export\\s+(?:async\\s+)?(?:function|class)\\s+${escapeRegex(symbol)}\\b`, "m");
  const lineCommentPattern = new RegExp(`//\\s+${escapeRegex(symbol)}\\b.*\\nexport\\s+(?:async\\s+)?(?:function|class)\\s+${escapeRegex(symbol)}\\b`, "m");
  return jsDocPattern.test(source) || lineCommentPattern.test(source);
}

function escapeRegex(value) {
  return value.replace(/[.*+?^${}()|[\]\\]/g, "\\$&");
}
