package main

import (
	"path/filepath"
	"strings"
)

/*
	Language Detection Rules:
	- .ts, .tsx -> TypeScript
	- .js, .jsx, .mjs, .cjs, .vue, .svelte -> JavaScript
	- .py -> Python
	- .go -> Go

	IMPORTANT: All languages (JS/TS, Python, Go) MUST support cross-file workspace analysis.
	When analyzing workspace, ALL files are analyzed together to detect:
	- Unused imports (check if imported in any other file)
	- Unused variables/functions (check if used in any other file)
	- Unused parameters

	Scan File: Single file analysis
	Scan Folder: All files in folder, cross-file analysis
	Scan Workspace: All files in workspace, cross-file analysis
*/

func DetectLanguage(filename string) Language {
	ext := strings.ToLower(filepath.Ext(filename))
	switch ext {
	case ".ts", ".tsx":
		return LangTypeScript
	case ".js", ".jsx", ".mjs", ".cjs", ".vue", ".svelte":
		return LangJavaScript
	case ".py":
		return LangPython
	case ".go":
		return LangGo
	default:
		return LangUnknown
	}
}
