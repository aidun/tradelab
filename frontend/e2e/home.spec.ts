import { expect, test } from "@playwright/test";

test("shows the TradeLab landing page", async ({ page }) => {
  await page.goto("/");

  await expect(page.getByRole("heading", { name: /paper trading with/i })).toBeVisible();
  await expect(page.getByText(/xrp \/ usdt/i)).toBeVisible();
});
