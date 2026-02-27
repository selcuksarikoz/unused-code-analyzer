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

	IMPORTANT: When adding/modifying analysis logic in any language file
	(javascript.go, python.go, go.go), you MUST apply the SAME changes to:
	- The single-file analysis function (e.g., findUsedNamesJS, findUsedNamesPython, findUsedGoNames)
	- The workspace analysis function in main.go (AnalyzeWorkspace)
	- All language-specific files to ensure consistent behavior

	Common bugs to avoid:
	- Checking usage INCLUDES the definition line itself (should skip the line where the name is defined)
	- This causes false negatives (unused items not detected)

	Scan File: Single file analysis
	Scan Folder: All files in folder, cross-file analysis
	Scan Workspace: All files in workspace, cross-file analysis
*/

func DetectLanguage(filename string) Language {
	ext := strings.ToLower(filepath.Ext(filename))
	switch ext {
	case ".py":
		return LangPython
	case ".go":
		return LangGo
	case ".rb":
		return LangRuby
	case ".php":
		return LangPHP
	default:
		return LangUnknown
	}
}
