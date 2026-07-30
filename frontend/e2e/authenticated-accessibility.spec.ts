import AxeBuilder from "@axe-core/playwright";
import { expect, test } from "@playwright/test";

const ACCESS_TOKEN = "e30.eyJleHAiOjQxMDI0NDQ4MDB9.signature";

test.beforeEach(async ({ context }) => {
  await context.addCookies([
    {
      name: "mate_access_token",
      value: ACCESS_TOKEN,
      domain: "localhost",
      path: "/",
      httpOnly: true,
      sameSite: "Lax",
    },
  ]);
});

test("authenticated shell and profile dialog have no detectable accessibility violations", async ({
  page,
}) => {
  await page.goto("/dashboard");
  await expect(
    page.getByRole("heading", { name: "Fleet overview" }),
  ).toBeVisible();

  const shellResults = await new AxeBuilder({ page }).analyze();
  expect(shellResults.violations).toEqual([]);

  await page
    .getByRole("button", { name: "Open profile for Alex Morgan" })
    .click();
  await expect(
    page.getByRole("dialog", { name: "Your profile" }),
  ).toBeVisible();

  const profileResults = await new AxeBuilder({ page }).analyze();
  expect(profileResults.violations).toEqual([]);
});
