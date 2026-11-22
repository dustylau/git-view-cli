import assert from "node:assert/strict";
import { readFile, readdir, stat } from "node:fs/promises";
import path from "node:path";
import { Given, Then } from "@cucumber/cucumber";

const root = process.cwd();
const read = (file) => readFile(path.join(root, file), "utf8");

Given("I read the project README", async function () {
  this.readme = await read("README.md");
});

Then(
  "the README describes {string}, {string}, and {string} modes",
  function (first, second, third) {
    for (const mode of [first, second, third])
      assert.ok(this.readme.includes(mode), mode);
  },
);

Then("the README states that the application is not implemented", function () {
  assert.match(
    this.readme,
    /No application, installable binary, or release exists/,
  );
});

Then("these documents are available:", async function (table) {
  for (const { file, topic } of table.hashes()) {
    const content = await read(file);
    assert.ok(content.includes(topic), `${file} must cover ${topic}`);
  }
});

Then(
  "the stack comparison keeps the application decision open",
  async function () {
    assert.match(await read("docs/STACK_COMPARISON.md"), /Decision: Open/);
  },
);

Then(
  "the contribution guide defines eleven ordered workflow steps",
  async function () {
    const content = await read("CONTRIBUTING.md");
    const steps = [...content.matchAll(/^(\d+)\. /gm)].map((match) =>
      Number(match[1]),
    );
    assert.deepEqual(
      steps,
      Array.from({ length: 11 }, (_, index) => index + 1),
    );
  },
);

async function markdownFiles(directory) {
  const files = [];
  for (const entry of await readdir(directory, { withFileTypes: true })) {
    if ([".git", "node_modules"].includes(entry.name)) continue;
    const file = path.join(directory, entry.name);
    if (entry.isDirectory()) files.push(...(await markdownFiles(file)));
    else if (entry.name.endsWith(".md")) files.push(file);
  }
  return files;
}

Then(
  "every local Markdown file link in project documents resolves",
  async function () {
    for (const file of await markdownFiles(root)) {
      const content = (await readFile(file, "utf8")).replace(
        /```[\s\S]*?```/g,
        "",
      );
      for (const match of content.matchAll(/\[[^\]]*\]\(([^)]+)\)/g)) {
        const target = match[1];
        if (/^(?:[a-z][a-z0-9+.-]*:|#)/i.test(target)) continue;
        const local = decodeURIComponent(target.split("#")[0]);
        assert.ok(
          (await stat(path.resolve(path.dirname(file), local))).isFile(),
          `${file}: ${target}`,
        );
      }
    }
  },
);
