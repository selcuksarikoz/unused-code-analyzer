const assert = require("assert");
const fs = require("fs");
const path = require("path");

const projectRoot = path.resolve(__dirname, "..");
const wasmPath = path.join(projectRoot, "out", "main.wasm");
const wasmExecPath = path.join(projectRoot, "out", "wasm_exec.js");

if (!fs.existsSync(wasmPath)) {
  console.error("WASM not found:", wasmPath);
  process.exit(1);
}

if (!fs.existsSync(wasmExecPath)) {
  console.error("wasm_exec.js not found:", wasmExecPath);
  process.exit(1);
}

require(wasmExecPath);

function normalizeResult(rawResult) {
  const parsed = JSON.parse(rawResult);
  const results = parsed.Results || parsed.results || {};
  const normalized = {};

  for (const [filename, issues] of Object.entries(results)) {
    normalized[filename] = {
      imports: issues?.imports || issues?.Imports || [],
      variables: issues?.variables || issues?.Variables || [],
      parameters: issues?.parameters || issues?.Parameters || [],
    };
  }

  return normalized;
}

async function loadWasmAnalyzer() {
  const go = new Go();
  const wasmBytes = fs.readFileSync(wasmPath);
  const { instance } = await WebAssembly.instantiate(wasmBytes, go.importObject);
  go.run(instance);
}

async function main() {
  await loadWasmAnalyzer();

  const workspace = {
    files: [
      {
        filename: "a.ts",
        hash: "1",
        content: [
          "import { foo } from './x';",
          "const local = 1;",
          "console.log(local);",
        ].join("\n"),
      },
      {
        filename: "b.ts",
        hash: "2",
        content: [
          "const foobar = 2;",
          "console.log(foobar);",
        ].join("\n"),
      },
      {
        filename: "c.rb",
        hash: "3",
        content: [
          "def add(a, b)",
          "  a + b",
          "end",
        ].join("\n"),
      },
      {
        filename: "d.php",
        hash: "4",
        content: [
          "<?php",
          "function add($a, $b) {",
          "  return $a + $b;",
          "}",
        ].join("\n"),
      },
    ],
  };

  const results = normalizeResult(globalThis.analyzeWorkspace(JSON.stringify(workspace)));

  assert(results["a.ts"], "missing result for a.ts");
  assert(results["c.rb"], "missing result for c.rb");
  assert(results["d.php"], "missing result for d.php");

  assert.strictEqual(results["a.ts"].imports.length, 1, "foo import should be unused when only foobar exists");
  assert.strictEqual(results["c.rb"].parameters.length, 0, "ruby params used in same function must not be flagged");
  assert.strictEqual(results["d.php"].parameters.length, 0, "php params used in same function must not be flagged");

  console.log("test-wasm.js passed");
}

main().catch((err) => {
  console.error(err);
  process.exit(1);
});
