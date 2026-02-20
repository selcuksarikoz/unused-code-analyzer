# Development Rules

## Architecture

### Cross-File Analysis Only
- **NEVER** analyze a single file in isolation
- **ALWAYS** analyze the entire workspace together
- For each import in each file, check if it's used in ANY other file in the workspace
- Import is UNUSED only if it's not used in ANY file (including its own file)

### Algorithm
1. Parse ALL files in workspace, extract imports and definitions
2. For each file's imports, check ALL other files in workspace for usage
3. Mark import as USED if found in ANY file
4. Mark import as UNUSED only if NOT found in ANY file

### Data Flow
```
Workspace (all files)
    │
    ▼
Parse all files → Extract: imports[], definitions[]
    │
    ▼
For each file:
  For each import in file:
    Check ALL other files for usage
    → usedNames[importName + "@" + filename] = true/false
    │
    ▼
Build result using usedNames map (cross-file data)
```

## SOLID & DRY Principles

### SOLID
- **S**ingle Responsibility: Each parser handles only one language
- **O**pen/Closed: Add new languages without modifying existing code
- **L**iskov Substitution: All parsers implement same interface
- **I**nterface Segregation: Simple, focused functions
- **D**ependency Inversion: Use abstractions, not concrete implementations

### DRY
- Common functionality goes to shared utilities
- No duplicated logic across language parsers
- Reuse `FindUsedNames`, tokenizer patterns, etc.

## Code Organization

### File Structure
```
backend/
├── main.go          # Entry point, workspace orchestration
├── jsparser.go      # JavaScript/TypeScript parser (tokenizer-based)
├── pythonparser.go  # Python parser
├── goparser.go      # Go parser
├── rubyparser.go    # Ruby parser
├── phpparser.go     # PHP parser
├── analyzer.go      # Generic builder utilities
├── language.go     # Language detection
├── types.go        # Shared types
└── utils.go        # Shared utilities
```

### Key Functions
- `AnalyzeWorkspace()` - Main entry, coordinates cross-file analysis
- `getParsedWorkspaceData()` - Extract imports/defs from all files
- `buildResult*()` - Uses cross-file `usedNames` map, NOT local analysis

## Common Mistakes to Avoid

### ❌ Wrong: Local-only analysis
```go
// This is WRONG - only looks at one file
func buildResultJSTS(file AnalyzeFile, ...) {
    localUsed := FindUsedJSNames(file.Content, ...)
    // Uses localUsed instead of cross-file usedNames!
}
```

### ✅ Correct: Cross-file analysis
```go
// This is CORRECT - uses workspace-wide usedNames
func buildResultJSTS(file AnalyzeFile, defs [], imports [], usedNames map[string]bool) {
    for _, imp := range imports {
        if !usedNames[imp.Name+"@"+file.Filename] {
            // unused!
        }
    }
}
```

### ❌ Wrong: Finding usage in same file only
```go
func FindUsedNames(content string, items []NamedItem) map[string]bool {
    // Only searches one content string!
}
```

## Testing

Before submitting:
1. Test with workspace containing multiple files
2. Verify unused imports are correctly detected
3. Verify used imports are NOT flagged as unused
4. Test cross-file: file A imports X, file B uses X → X should NOT be unused
