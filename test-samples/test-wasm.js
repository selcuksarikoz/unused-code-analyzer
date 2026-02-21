const fs = require('fs');
const path = require('path');

const wasmPath = path.join(__dirname, 'out', 'main.wasm');
const wasmExecPath = path.join(__dirname, 'out', 'wasm_exec.js');

if (!fs.existsSync(wasmPath)) {
  console.error('WASM not found:', wasmPath);
  process.exit(1);
}

require(wasmExecPath);
const go = new Go();

WebAssembly.instantiate(fs.readFileSync(wasmPath), go.importObject)
  .then(result => {
    go.run(result.instance);
    
    const files = [
      { filename: 'test-samples/main.ts', content: fs.readFileSync('test-samples/main.ts', 'utf-8'), hash: '1' },
      { filename: 'test-samples/utils.ts', content: fs.readFileSync('test-samples/utils.ts', 'utf-8'), hash: '2' },
      { filename: 'test-samples/reexports.ts', content: fs.readFileSync('test-samples/reexports.ts', 'utf-8'), hash: 'reex' },
      { filename: 'test-samples/feature.ts', content: fs.readFileSync('test-samples/feature.ts', 'utf-8'), hash: 'feat' },
      { filename: 'test-samples/main.jsx', content: fs.readFileSync('test-samples/main.jsx', 'utf-8'), hash: 'jsx' },
      { filename: 'test-samples/main.vue', content: fs.readFileSync('test-samples/main.vue', 'utf-8'), hash: 'vue' },
      { filename: 'test-samples/main.py', content: fs.readFileSync('test-samples/main.py', 'utf-8'), hash: '3' },
      { filename: 'test-samples/utils.py', content: fs.readFileSync('test-samples/utils.py', 'utf-8'), hash: '4' },
      { filename: 'test-samples/feature.py', content: fs.readFileSync('test-samples/feature.py', 'utf-8'), hash: 'featpy' },
      { filename: 'test-samples/main.rb', content: fs.readFileSync('test-samples/main.rb', 'utf-8'), hash: '5' },
      { filename: 'test-samples/utils.rb', content: fs.readFileSync('test-samples/utils.rb', 'utf-8'), hash: '6' },
      { filename: 'test-samples/feature.rb', content: fs.readFileSync('test-samples/feature.rb', 'utf-8'), hash: 'featruby' },
      { filename: 'test-samples/main.php', content: fs.readFileSync('test-samples/main.php', 'utf-8'), hash: '7' },
      { filename: 'test-samples/utils.php', content: fs.readFileSync('test-samples/utils.php', 'utf-8'), hash: '8' },
      { filename: 'test-samples/feature.php', content: fs.readFileSync('test-samples/feature.php', 'utf-8'), hash: 'featphp' },
      { filename: 'test-samples/main.go', content: fs.readFileSync('test-samples/main.go', 'utf-8'), hash: 'go' },
    ];

    const analysisResult = globalThis.analyzeWorkspace(JSON.stringify({ files }));
    console.log(analysisResult);
  })
  .catch(console.error);
