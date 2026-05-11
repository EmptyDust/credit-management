import { readdirSync } from "node:fs";
import { resolve } from "node:path";

import { describe, expect, it } from "vitest";

describe("legacy page cleanup", () => {
  it("keeps src/pages as a 404-only compatibility boundary", () => {
    const pagesDir = resolve(process.cwd(), "src/pages");
    const entries = readdirSync(pagesDir, { recursive: true, withFileTypes: true })
      .filter((entry) => entry.isFile())
      .map((entry) => resolve(entry.parentPath, entry.name))
      .map((path) => path.replace(`${pagesDir}/`, ""))
      .sort();

    expect(entries).toEqual(["NotFound.tsx"]);
  });
});
