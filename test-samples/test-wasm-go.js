const assert = require("assert");
const fs = require("fs");
const path = require("path");

const projectRoot = path.resolve(__dirname, "..");
const wasmPath = path.join(projectRoot, "out", "main.wasm");
const wasmExecPath = path.join(projectRoot, "out", "wasm_exec.js");

if (!fs.existsSync(wasmPath) || !fs.existsSync(wasmExecPath)) {
  console.error("WASM artifacts are missing. Run: bash scripts/build.sh");
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
        filename: "a.go",
        hash: "1",
        content: [
          "package main",
          "",
          'import "foo"',
          "",
          "func main() {}",
        ].join("\n"),
      },
      {
        filename: "b.go",
        hash: "2",
        content: [
          "package main",
          "",
          "func test() {",
          "  foobar := 1",
          "  _ = foobar",
          "}",
        ].join("\n"),
      },
      {
        filename: "c.go",
        hash: "3",
        content: [
          "package main",
          "",
          "func usesImport() {",
          "  foo.Println()",
          "}",
        ].join("\n"),
      },
    ],
  };

  const results = normalizeResult(globalThis.analyzeWorkspace(JSON.stringify(workspace)));

  assert(results["a.go"], "missing result for a.go");

  const importIssues = results["a.go"].imports.map((x) => x.text);
  assert.strictEqual(
    importIssues.length,
    0,
    "import foo should be treated as used when another file references exact identifier foo",
  );

  const bIssues = results["b.go"].imports.map((x) => x.text);
  assert.strictEqual(
    bIssues.length,
    0,
    "b.go has no imports and should not receive import issues",
  );

  console.log("test-wasm-go.js passed");
}

main().catch((err) => {
  console.error(err);
  process.exit(1);
});
