package main

import (
	"fmt"
	"regexp"
	"strings"
)

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

var (
	reJSImport     = regexp.MustCompile(`^\s*import\s+(?:(?:\{[^}]*\}|\*\s+as\s+[^}]+|[a-zA-Z_$][a-zA-Z0-9_$]*)(?:\s*,\s*(?:\{[^}]*\}|\*\s+as\s+[^}]+|[a-zA-Z_$][a-zA-Z0-9_$]*))*\s+from\s+)?['"]([^'"]+)['"]`)
	reJSImportAlt  = regexp.MustCompile(`^\s*import\s*\(\s*['"]([^'"]+)['"]\s*\)`)
	reJSFuncDecl   = regexp.MustCompile(`(?:^|\n)\s*function\s+([a-zA-Z_$][a-zA-Z0-9_$]*)\s*\(`)
	reJSClassDecl  = regexp.MustCompile(`(?:^|\n)\s*(?:export\s+)?(?:default\s+)?class\s+([a-zA-Z_$][a-zA-Z0-9_$]*)`)
	reJSVarDecl    = regexp.MustCompile(`(?:^|\n)\s*(?:const|let|var)\s+([a-zA-Z_$][a-zA-Z0-9_$]*)\s*=`)
	reJSArrowFunc  = regexp.MustCompile(`(?:^|\n)\s*(?:const|let|var)\s+([a-zA-Z_$][a-zA-Z0-9_$]*)\s*=\s*(?:\([^)]*\)|[a-zA-Z_$][a-zA-Z0-9_$]*)\s*=>`)
	reJSMethodDecl = regexp.MustCompile(`(?:^|\n)\s*([a-zA-Z_$][a-zA-Z0-9_$]*)\s*\([^)]*\)\s*\{`)
)

func analyzeJavaScript(content, filename string) AnalysisResult {
	fmt.Printf("[JS] Analyzing: %s\n", filename)
	defs := findJSDefinitions(content, filename)
	imports := findJSImports(content, filename)
	parameters := findJSParameters(content, filename)

	fmt.Printf("[JS] Found defs: %d, imports: %d, params: %d\n", len(defs), len(imports), len(parameters))

	if len(imports) == 0 && len(defs) > 0 {
		fmt.Printf("[JS] WARNING: No imports found in %s! First 500 chars:\n%s\n", filename, content[:min(500, len(content))])
	}

	used := findUsedNamesJS(content, defs)

	fmt.Printf("[JS] Used names: %v\n", used)

	var unusedImports, unusedVars, unusedParams []CodeIssue

	for _, imp := range imports {
		if !used[imp.Name] {
			unusedImports = append(unusedImports, CodeIssue{
				ID:   generateUUID(),
				Line: imp.Line,
				Text: "import " + imp.Name,
				File: filename,
			})
		}
	}

	for _, v := range defs {
		if v.Exported {
			continue
		}
		if !used[v.Name] {
			unusedVars = append(unusedVars, CodeIssue{
				ID:   generateUUID(),
				Line: v.Line,
				Text: v.Type + " " + v.Name,
				File: filename,
			})
		}
	}

	for _, p := range parameters {
		paramName := strings.TrimPrefix(p.Text, "parameter ")
		if !used[paramName] {
			unusedParams = append(unusedParams, CodeIssue{
				ID:   generateUUID(),
				Line: p.Line,
				Text: p.Text,
				File: filename,
			})
		}
	}

	fmt.Printf("[JS] Result - unused imports: %d, vars: %d, params: %d\n", len(unusedImports), len(unusedVars), len(unusedParams))

	return AnalysisResult{
		Imports:    unusedImports,
		Variables:  unusedVars,
		Parameters: unusedParams,
	}
}

func analyzeJavaScriptForWorkspace(content, filename string) ([]Definition, []Import, []CodeIssue, []CodeIssue) {
	defs := findJSDefinitions(content, filename)
	imports := findJSImports(content, filename)
	parameters := findJSParameters(content, filename)

	used := findUsedNamesJS(content, defs)

	var unusedVars []CodeIssue
	for _, v := range defs {
		if v.Exported {
			continue
		}
		if !used[v.Name] {
			unusedVars = append(unusedVars, CodeIssue{
				ID:   generateUUID(),
				Line: v.Line,
				Text: v.Type + " " + v.Name,
				File: filename,
			})
		}
	}

	var unusedParams []CodeIssue
	for _, p := range parameters {
		paramName := strings.TrimPrefix(p.Text, "parameter ")
		if !used[paramName] {
			unusedParams = append(unusedParams, CodeIssue{
				ID:   generateUUID(),
				Line: p.Line,
				Text: p.Text,
				File: filename,
			})
		}
	}

	return defs, imports, unusedVars, unusedParams
}

func findJSImports(content string, filename string) []Import {
	fmt.Printf("[JS Imports] Starting for: %s\n", filename)
	var imports []Import
	seen := make(map[string]bool)
	lines := strings.Split(content, "\n")

	fmt.Printf("[JS Imports] Total lines: %d\n", len(lines))

	// Print first 10 lines for debugging
	for i := 0; i < min(10, len(lines)); i++ {
		fmt.Printf("[JS Imports] Line %d: %s\n", i+1, lines[i])
	}

	testLine := `import { useState } from "react";`
	matches := reJSImport.FindStringSubmatch(testLine)
	fmt.Printf("[JS Imports] Test regex on '%s': matches=%d\n", testLine, len(matches))

	for i, line := range lines {
		line = strings.TrimSpace(line)
		lineNum := i + 1

		if strings.HasPrefix(line, "import ") {
			fmt.Printf("[JS Imports] Import line %d: %s\n", lineNum, line)
			matches := reJSImport.FindStringSubmatch(line)
			fmt.Printf("[JS Imports] Regex matches: %d\n", len(matches))
		}

		if strings.HasPrefix(line, "//") {
			continue
		}

		if matches := reJSImport.FindStringSubmatch(line); len(matches) > 2 {
			fmt.Printf("[JS Imports] Matched import at line %d: %s\n", lineNum, line)
			modulePath := matches[2]
			importPart := matches[1]
			names := extractJSImportNames(importPart)
			fmt.Printf("[JS Imports] Found names: %v from import part: %s\n", names, importPart)

			for _, name := range names {
				key := name + "|" + modulePath
				if name != "" && name != "_" && !seen[key] {
					seen[key] = true
					imports = append(imports, Import{
						Name:   name,
						File:   filename,
						Line:   lineNum,
						Source: modulePath,
					})
				}
			}
		} else if matches := reJSImportAlt.FindStringSubmatch(line); len(matches) > 1 {
			modulePath := matches[1]
			imports = append(imports, Import{
				Name:   modulePath,
				File:   filename,
				Line:   lineNum,
				Source: modulePath,
			})
		}
	}
	fmt.Printf("[JS Imports] Final count: %d\n", len(imports))
	return imports
}

func extractJSImportNames(importStr string) []string {
	var names []string

	importStr = strings.TrimSpace(importStr)

	if strings.HasPrefix(importStr, "{") {
		inside := strings.Trim(importStr, "{}")
		parts := strings.Split(inside, ",")
		for _, part := range parts {
			part = strings.TrimSpace(part)
			if part == "" {
				continue
			}
			if strings.Contains(part, " as ") {
				asParts := strings.Split(part, " as ")
				if len(asParts) >= 2 {
					names = append(names, strings.TrimSpace(asParts[len(asParts)-1]))
				}
			} else {
				names = append(names, part)
			}
		}
	} else if strings.HasPrefix(importStr, "*") {
		parts := strings.Split(importStr, " as ")
		if len(parts) >= 2 {
			names = append(names, strings.TrimSpace(parts[1]))
		}
	} else if importStr != "" {
		names = append(names, strings.TrimSpace(importStr))
	}

	return names
}

func findJSDefinitions(content, filename string) []Definition {
	var defs []Definition
	seen := make(map[string]bool)

	lines := strings.Split(content, "\n")
	for i, line := range lines {
		line = strings.TrimSpace(line)
		lineNum := i + 1

		if strings.HasPrefix(line, "//") || strings.HasPrefix(line, "/*") {
			continue
		}

		isExported := strings.HasPrefix(line, "export ")
		isExportDefault := strings.HasPrefix(line, "export default ")

		if matches := reJSFuncDecl.FindStringSubmatch(line); len(matches) > 1 {
			name := matches[1]
			if name != "" && !seen[name] && !isJSKeyword(name) {
				seen[name] = true
				defs = append(defs, Definition{
					Name:     name,
					File:     filename,
					Line:     lineNum,
					Type:     "function",
					Exported: isExported || isExportDefault,
				})
			}
		}

		if matches := reJSClassDecl.FindStringSubmatch(line); len(matches) > 1 {
			name := matches[1]
			if name != "" && !seen[name] && !isJSKeyword(name) {
				seen[name] = true
				defs = append(defs, Definition{
					Name:     name,
					File:     filename,
					Line:     lineNum,
					Type:     "class",
					Exported: isExported || isExportDefault,
				})
			}
		}

		if matches := reJSVarDecl.FindStringSubmatch(line); len(matches) > 1 {
			name := matches[1]
			if name != "" && !seen[name] && isValidJSIdent(name) {
				seen[name] = true
				defs = append(defs, Definition{
					Name:     name,
					File:     filename,
					Line:     lineNum,
					Type:     "var",
					Exported: isExported,
				})
			}
		}

		if matches := reJSArrowFunc.FindStringSubmatch(line); len(matches) > 1 {
			name := matches[1]
			if name != "" && !seen[name] && isValidJSIdent(name) {
				seen[name] = true
				defs = append(defs, Definition{
					Name:     name,
					File:     filename,
					Line:     lineNum,
					Type:     "const",
					Exported: isExported,
				})
			}
		}

		if matches := reJSMethodDecl.FindStringSubmatch(line); len(matches) > 1 {
			name := matches[1]
			if name != "" && !seen[name] && isValidJSIdent(name) && !isJSKeyword(name) {
				seen[name] = true
				defs = append(defs, Definition{
					Name:     name,
					File:     filename,
					Line:     lineNum,
					Type:     "method",
					Exported: false,
				})
			}
		}
	}
	return defs
}

func isJSKeyword(name string) bool {
	keywords := map[string]bool{
		"if": true, "else": true, "for": true, "while": true, "do": true,
		"switch": true, "case": true, "default": true, "break": true,
		"return": true, "continue": true, "try": true, "catch": true,
		"finally": true, "throw": true, "new": true, "delete": true,
		"typeof": true, "instanceof": true, "void": true, "yield": true,
		"await": true, "async": true, "import": true, "export": true,
		"from": true, "as": true, "class": true, "extends": true,
		"super": true, "this": true, "static": true, "get": true, "set": true,
		"const": true, "let": true, "var": true, "function": true,
		"true": true, "false": true, "null": true, "undefined": true,
		"in": true, "of": true, "NaN": true, "Infinity": true,
	}
	return keywords[name]
}

func isValidJSIdent(s string) bool {
	if s == "" {
		return false
	}
	re := regexp.MustCompile(`^[a-zA-Z_$][a-zA-Z0-9_$]*$`)
	return re.MatchString(s)
}

func findJSParameters(content, filename string) []CodeIssue {
	var params []CodeIssue
	seen := make(map[string]bool)
	reFuncParams := regexp.MustCompile(`function\s*(?:[a-zA-Z_$][a-zA-Z0-9_$]*)?\s*\(([^)]*)\)`)
	reArrowParams := regexp.MustCompile(`=>\s*\(([^)]*)\)|([a-zA-Z_$][a-zA-Z0-9_$]*)\s*=>\s*[{(]`)

	lines := strings.Split(content, "\n")
	for i, line := range lines {
		line = strings.TrimSpace(line)
		lineNum := i + 1

		if matches := reFuncParams.FindStringSubmatch(line); len(matches) > 1 {
			paramsList := matches[1]
			paramNames := extractJSParams(paramsList)
			for _, name := range paramNames {
				if name != "" && name != "this" && !seen[name] {
					seen[name] = true
					params = append(params, CodeIssue{
						Line: lineNum,
						Text: "parameter " + name,
						File: filename,
						ID:   generateUUID(),
					})
				}
			}
		}

		if matches := reArrowParams.FindStringSubmatch(line); len(matches) > 1 {
			paramsList := matches[1]
			if paramsList == "" {
				paramsList = matches[2]
			}
			if paramsList != "" {
				paramNames := extractJSParams(paramsList)
				for _, name := range paramNames {
					if name != "" && name != "this" && !seen[name] {
						seen[name] = true
						params = append(params, CodeIssue{
							Line: lineNum,
							Text: "parameter " + name,
							File: filename,
							ID:   generateUUID(),
						})
					}
				}
			}
		}
	}
	return params
}

func extractJSParams(params string) []string {
	var names []string
	if params == "" {
		return names
	}

	parts := strings.Split(params, ",")
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}

		re := regexp.MustCompile(`^([a-zA-Z_$][a-zA-Z0-9_$]*)`)
		match := re.FindStringSubmatch(part)
		if len(match) > 1 {
			names = append(names, match[1])
		}
	}
	return names
}

func findUsedNamesJS(content string, defs []Definition) map[string]bool {
	var items []NamedItem
	for _, def := range defs {
		items = append(items, NamedItem{Name: def.Name, Line: def.Line})
	}
	return FindUsedNames(content, items)
}
