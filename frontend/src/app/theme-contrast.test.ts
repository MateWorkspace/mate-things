import { readFileSync } from "node:fs";
import { join } from "node:path";
import { describe, expect, it } from "vitest";

const themeCss = readFileSync(
  join(process.cwd(), "src/app/globals.css"),
  "utf8",
);

function tokenPair(name: string): readonly [string, string] {
  const values = [
    ...themeCss.matchAll(
      new RegExp(`--${name}:\\s*(#[0-9a-f]{6})(?:;|\\s)`, "gi"),
    ),
  ].map((match) => match[1]);

  if (values.length !== 2) {
    throw new Error(`Expected light and dark values for --${name}`);
  }

  return [values[0], values[1]];
}

function relativeLuminance(hex: string): number {
  const channels = hex
    .slice(1)
    .match(/../g)!
    .map((value) => Number.parseInt(value, 16) / 255)
    .map((value) =>
      value <= 0.04045 ? value / 12.92 : Math.pow((value + 0.055) / 1.055, 2.4),
    );

  return 0.2126 * channels[0] + 0.7152 * channels[1] + 0.0722 * channels[2];
}

function contrastRatio(first: string, second: string): number {
  const firstLuminance = relativeLuminance(first);
  const secondLuminance = relativeLuminance(second);
  const lighter = Math.max(firstLuminance, secondLuminance);
  const darker = Math.min(firstLuminance, secondLuminance);

  return (lighter + 0.05) / (darker + 0.05);
}

describe("semantic contrast tokens", () => {
  it("keeps muted text at 4.5:1 or better in both modes", () => {
    const [lightText, darkText] = tokenPair("muted-foreground");
    const [lightBackground, darkBackground] = tokenPair("background");

    expect(contrastRatio(lightText, lightBackground)).toBeGreaterThanOrEqual(
      4.5,
    );
    expect(contrastRatio(darkText, darkBackground)).toBeGreaterThanOrEqual(4.5);
  });

  it("keeps control boundaries at 3:1 or better in both modes", () => {
    const [lightBorder, darkBorder] = tokenPair("control-border");
    const [lightBackground, darkBackground] = tokenPair("background");

    expect(contrastRatio(lightBorder, lightBackground)).toBeGreaterThanOrEqual(
      3,
    );
    expect(contrastRatio(darkBorder, darkBackground)).toBeGreaterThanOrEqual(3);
  });
});
