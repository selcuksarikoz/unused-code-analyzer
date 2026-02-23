# Unused Code Analyzer

<p align="center">
  <img src="assets/icons/logo.png" alt="Logo" width="128" height="128">
</p>

<p align="center">
  <a href="https://marketplace.visualstudio.com/items?itemName=selcuksarikoz.get-unused-imports">
    <img src="https://img.shields.io/visual-studio-marketplace/v/selcuksarikoz.get-unused-imports" alt="VS Code Marketplace">
  </a>
  <a href="https://marketplace.visualstudio.com/items?itemName=selcuksarikoz.get-unused-imports">
    <img src="https://img.shields.io/visual-studio-marketplace/d/selcuksarikoz.get-unused-imports" alt="Downloads">
  </a>
  <a href="LICENSE">
    <img src="https://img.shields.io/github/license/selcuksarikoz/get-unused-imports" alt="License">
  </a>
</p>

A VS Code extension to detect unused imports, variables, and parameters in your codebase. Built with Go (WASM) for high performance.

## Features

- **Multi-language Support**: TypeScript, JavaScript, Python, Go, Ruby, PHP
- **Auto Analyzer**: Starts only after you open the "Unused Code" activity view, then analyzes on file changes/saves
- **Tree View**: Results displayed in Explorer sidebar
- **Quick Navigation**: Click to jump to unused code
- **High Performance**: Go backend compiled to WebAssembly with intelligent caching

## Installation

1. Open VS Code
2. Go to Extensions (`Cmd+Shift+P` → "Extensions: Open")
3. Search for "Unused Code Analyzer"
4. Click Install

Or install from [VS Code Marketplace](https://marketplace.visualstudio.com/items?itemName=selcuksarikoz.get-unused-imports)

## Usage

### Command Palette
- `Get Unused: Scan Workspace` - Analyze entire workspace
- `Get Unused: Scan File` - Analyze current file
- `Get Unused: Scan Folder` - Analyze selected folder

### Auto Analyzer
The extension automatically analyzes files when:
- The "Unused Code" activity view is opened at least once
- A relevant file is changed or saved after that
- Results are updated in real-time

### Tree View
Open "Unused Code" view in Explorer sidebar. Results are shown with:
- File names and issue counts
- Issue types (Imports, Variables, Parameters)
- Click to jump to the issue

### Settings

| Setting | Default | Description |
|---------|---------|-------------|
| `get-unused-imports.autoAnalyzer` | `true` | Automatically analyze on save after the "Unused Code" activity view is opened |
| `get-unused-imports.autoAnalyzeDelay` | `500` | Delay in ms before auto-analyzing |
| `get-unused-imports.fileExtensions` | `["ts", "tsx", "js", "jsx", "vue", "svelte", "py", "go", "rb", "php"]` | File extensions to scan |
| `get-unused-imports.excludeFolders` | `["node_modules", ".next", "dist", "build", "out", ".git"]` | Folders to exclude |

## Supported Languages

| Language | Extensions | Backend |
|----------|------------|---------|
| TypeScript | .ts, .tsx | Native (ts-morph) |
| JavaScript | .js, .jsx, .vue, .svelte | Native (ts-morph) |
| Python | .py | WASM (go-python) |
| Go | .go | WASM (go/parser) |
| Ruby | .rb | WASM (regex-based) |
| PHP | .php | WASM (regex-based) |

## Architecture

- **Backend**: Go → WebAssembly
- **Frontend**: TypeScript + VS Code API
- **UI**: Tailwind CSS
- **Caching**: MD5 hash-based caching for fast incremental analysis

## Development

```bash
# Install dependencies
npm install

# Build WASM (from backend directory)
cd backend && GOOS=js GOARCH=wasm go build -o ../out/main.wasm .

# Build extension
npm run build
```

## License

MIT License - see [LICENSE](LICENSE)

---

<p align="center">
  <strong>☕ Enjoying Unused Code Analyzer? <a href="https://buymeacoffee.com/funnyturkishdude">Buy me a coffee</a> or leave a rating on the <a href="https://marketplace.visualstudio.com/items?itemName=selcuksarikoz.get-unused-imports">Marketplace</a>!</strong>
</p>
