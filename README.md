# Get Unused Imports

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

- **Multi-language Support**: TypeScript, JavaScript, Python, Go
- **Auto Analyzer**: Automatically analyzes files on save and when TreeView is opened
- **Tree View**: Results displayed in Explorer sidebar
- **Webview**: Detailed view with tabbed interface
- **Quick Navigation**: Click to jump to unused code
- **High Performance**: Go backend compiled to WebAssembly with intelligent caching

## Installation

1. Open VS Code
2. Go to Extensions (`Cmd+Shift+P` → "Extensions: Open")
3. Search for "Get Unused Imports"
4. Click Install

Or install from [VS Code Marketplace](https://marketplace.visualstudio.com/items?itemName=selcuksarikoz.get-unused-imports)

## Usage

### Command Palette
- `Get Unused: Scan Workspace` - Analyze entire workspace
- `Get Unused: Scan File` - Analyze current file
- `Get Unused: Scan Folder` - Analyze selected folder
- `Get Unused: Analyzer` - Open webview

### Auto Analyzer
The extension automatically analyzes files when:
- A file is saved
- The "Unused Code" TreeView is opened
- Results are updated in real-time

### Tree View
Open "Unused Code" view in Explorer sidebar. Results are shown with:
- File names and issue counts
- Issue types (Imports, Variables, Parameters)
- Click to jump to the issue

### Settings

| Setting | Default | Description |
|---------|---------|-------------|
| `get-unused-imports.autoAnalyzer` | `true` | Automatically analyze on save |
| `get-unused-imports.autoAnalyzeDelay` | `500` | Delay in ms before auto-analyzing |
| `get-unused-imports.fileExtensions` | `["ts", "tsx", "js", "jsx", "vue", "svelte", "py", "go"]` | File extensions to scan |
| `get-unused-imports.excludeFolders` | `["node_modules", ".next", "dist", "build", "out", ".git"]` | Folders to exclude |

## Supported Languages

| Language | Extensions |
|----------|------------|
| TypeScript | .ts, .tsx |
| JavaScript | .js, .jsx, .vue, .svelte |
| Python | .py |
| Go | .go |

## Architecture

- **Backend**: Go → WebAssembly
- **Frontend**: TypeScript + VS Code API
- **UI**: Tailwind CSS + Material Icons
- **Caching**: SHA256 hash-based caching for fast incremental analysis

## Development

```bash
# Install dependencies
npm install

# Build extension
npm run build
```

## License

MIT License - see [LICENSE](LICENSE)
