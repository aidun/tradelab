import { defineConfig } from "vitest/config";
import { fileURLToPath } from "node:url";

export default defineConfig({
  resolve: {
    alias: {
      "@": fileURLToPath(new URL(".", import.meta.url))
    }
  },
  test: {
    globals: true,
    environment: "jsdom",
    setupFiles: ["./vitest.setup.ts"],
    include: ["__tests__/**/*.test.ts?(x)"],
    coverage: {
      provider: "v8",
      reporter: ["text", "json-summary"],
      include: ["components/**/*.{ts,tsx}", "lib/**/*.{ts,tsx}"],
      exclude: ["**/*.test.ts", "**/*.test.tsx"],
      thresholds: {
        statements: 45,
        lines: 45,
        functions: 50,
        branches: 50
      }
    }
  }
});
