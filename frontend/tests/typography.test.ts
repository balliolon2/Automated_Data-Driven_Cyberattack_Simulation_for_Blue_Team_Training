import test from "node:test";
import assert from "node:assert/strict";
import fs from "node:fs";
import path from "node:path";
import { fileURLToPath } from "node:url";

const __dirname = path.dirname(fileURLToPath(import.meta.url));
const frontendRoot = path.resolve(__dirname, "..");
const srcRoot = path.join(frontendRoot, "src");

const MIN_PX = 12;
const MIN_REM = 0.75;

function readUtf8(relPath: string): string {
  return fs.readFileSync(path.join(frontendRoot, relPath), "utf8");
}

function isSubFloor(val: number, unit?: string): boolean {
  const normalized = (unit ?? "px").toLowerCase();
  if (normalized === "rem" || normalized === "em") {
    return val < MIN_REM;
  }
  return val < MIN_PX;
}

function collectAllSourceFiles(dir: string): string[] {
  const results: string[] = [];
  for (const entry of fs.readdirSync(dir, { withFileTypes: true })) {
    const abs = path.join(dir, entry.name);
    if (entry.isDirectory()) {
      results.push(...collectAllSourceFiles(abs));
    } else if (/\.(tsx|ts|css)$/.test(entry.name)) {
      results.push(abs);
    }
  }
  return results.sort();
}

function scanTypographyViolations(files: string[]): string[] {
  const violations: string[] = [];

  for (const absPath of files) {
    const relPath = path.relative(frontendRoot, absPath).replace(/\\/g, "/");
    const content = fs.readFileSync(absPath, "utf8");
    const lines = content.split(/\r?\n/);

    lines.forEach((line, idx) => {
      const lineNo = idx + 1;

      // 1. Zero arbitrary pixel font sizes (rem-only enforcement)
      for (const m of line.matchAll(/\btext-\[\d+(?:\.\d+)?px\]/g)) {
        violations.push(`${relPath}:${lineNo} forbidden arbitrary px font size "${m[0]}"`);
      }

      // 2. Zero sub-floor rem/em arbitrary font sizes (< 0.75)
      for (const m of line.matchAll(/\btext-\[(\d*\.?\d+)(rem|em)\]/g)) {
        if (isSubFloor(Number.parseFloat(m[1]), m[2])) {
          violations.push(`${relPath}:${lineNo} sub-floor font size "${m[0]}" (< ${MIN_REM}${m[2]})`);
        }
      }

      // 3. Zero non-standard micro utility classes
      for (const m of line.matchAll(/\btext-(?:2xs|3xs)\b/g)) {
        violations.push(`${relPath}:${lineNo} forbidden micro font class "${m[0]}"`);
      }

      // 4. Zero inline JS fontSize below 12px / 0.75rem (quoted or unquoted, with or without unit)
      for (const m of line.matchAll(/\bfontSize\s*:\s*['"]?(\d*\.?\d+)(px|rem|em)?['"]?/gi)) {
        if (isSubFloor(Number.parseFloat(m[1]), m[2])) {
          violations.push(`${relPath}:${lineNo} sub-floor inline fontSize "${m[0]}"`);
        }
      }

      // 5. Zero CSS font-size below 12px / 0.75rem
      for (const m of line.matchAll(/\bfont-size\s*:\s*(\d*\.?\d+)(px|rem|em)\b/gi)) {
        if (isSubFloor(Number.parseFloat(m[1]), m[2])) {
          violations.push(`${relPath}:${lineNo} sub-floor CSS font-size "${m[0]}"`);
        }
      }

      // 6. Zero invalid graphite-550 palette class
      for (const m of line.matchAll(/\btext-graphite-550\b/g)) {
        violations.push(`${relPath}:${lineNo} invalid palette class "${m[0]}"`);
      }
    });
  }

  return violations;
}

test("tailwind.config.js defines Thai fallback in sans & mono stacks and rem-based log fontSize token", async () => {
  const configMod = await import("../tailwind.config.js");
  const config = configMod.default;

  const sansStack: string[] = config?.theme?.extend?.fontFamily?.sans ?? [];
  assert.ok(
    sansStack.some((f) => f.includes("IBM Plex Sans Thai")),
    `fontFamily.sans must include "IBM Plex Sans Thai" fallback, got: ${JSON.stringify(sansStack)}`,
  );

  const monoStack: string[] = config?.theme?.extend?.fontFamily?.mono ?? [];
  assert.ok(
    monoStack.some((f) => f.includes("IBM Plex Sans Thai")),
    `fontFamily.mono must include "IBM Plex Sans Thai" fallback, got: ${JSON.stringify(monoStack)}`,
  );

  const logToken = config?.theme?.extend?.fontSize?.log;
  assert.ok(Array.isArray(logToken), "theme.extend.fontSize.log token must be a [size, options] tuple");
  const [logSize, logOpts] = logToken as [string, { lineHeight?: string }];
  assert.equal(logSize, "0.8125rem", `fontSize.log must be "0.8125rem", got "${logSize}"`);
  assert.equal(logOpts?.lineHeight, "1.5", `fontSize.log lineHeight must be "1.5", got "${logOpts?.lineHeight}"`);
  assert.ok(!isSubFloor(Number.parseFloat(logSize), "rem"), `fontSize.log must be >= ${MIN_REM}rem`);
});

test("index.css sets body default typography to text-base with leading-relaxed", () => {
  const css = readUtf8("src/index.css");
  const bodyBlock = css.match(/body\s*\{[\s\S]*?\}/);
  assert.ok(bodyBlock, "src/index.css must contain a body rule");
  assert.match(bodyBlock[0], /\btext-base\b/, "body rule must apply text-base");
  assert.match(bodyBlock[0], /\bleading-relaxed\b/, "body rule must apply leading-relaxed");
});

test("unused scaffold App.css is removed", () => {
  assert.equal(
    fs.existsSync(path.join(srcRoot, "App.css")),
    false,
    "unreferenced src/App.css should be deleted",
  );
});

test("entire frontend/src tree enforces >= 12px hard floor, rem-only typography, and valid graphite shades", () => {
  const files = collectAllSourceFiles(srcRoot);
  assert.ok(files.length >= 26, `expected at least 26 frontend source files, found ${files.length}`);
  const violations = scanTypographyViolations(files);
  assert.deepEqual(violations, [], `Typography violations found:\n${violations.join("\n")}`);
});
