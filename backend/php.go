package main

import (
	"regexp"
	"strings"
)

func analyzePHP(content, filename string) AnalysisResult {
	imports := findPHPImports(content)
	variables := findPHPVariables(content)
	parameters := findPHPParameters(content)
	used := findUsedNamesPHP(content, imports, variables)
	paramUsed := findUsedParameterNames(content, parameters)

	var unusedImports, unusedVars, unusedParams []CodeIssue

	for _, imp := range imports {
		if !used[imp.Name] {
			unusedImports = append(unusedImports, CodeIssue{
				ID:   generateUUID(),
				Line: imp.Line,
				Text: "use " + imp.Source + ";",
				File: filename,
			})
		}
	}

	for _, v := range variables {
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
		if !paramUsed[paramName] {
			unusedParams = append(unusedParams, CodeIssue{
				ID:   generateUUID(),
				Line: p.Line,
				Text: p.Text,
				File: filename,
			})
		}
	}

	return AnalysisResult{
		Imports:    unusedImports,
		Variables:  unusedVars,
		Parameters: unusedParams,
	}
}

func analyzePHPForWorkspace(content, filename string) ([]Definition, []Import, []CodeIssue, []CodeIssue) {
	imports := findPHPImports(content)
	variables := findPHPDefinitions(content, filename)
	parameters := findPHPParametersForWorkspace(content, filename)
	used := findUsedNamesPHP(content, imports, toPHPDefSlice(variables))
	paramUsed := findUsedParameterNames(content, parameters)

	var unusedVars, unusedParams []CodeIssue

	for _, v := range variables {
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
		if !paramUsed[paramName] {
			unusedParams = append(unusedParams, CodeIssue{
				ID:   generateUUID(),
				Line: p.Line,
				Text: p.Text,
				File: filename,
			})
		}
	}

	return toPHPDefSlice(variables), importsToPHPSlice(imports, filename), unusedVars, unusedParams
}

func findPHPImports(content string) []Import {
	var imports []Import
	lines := strings.Split(content, "\n")

	useRe := regexp.MustCompile(`^\s*use\s+([^;]+);`)
	useFunctionRe := regexp.MustCompile(`^\s*use\s+function\s+([^;]+);`)
	useConstRe := regexp.MustCompile(`^\s*use\s+const\s+([^;]+);`)

	for i, line := range lines {
		lineNum := i + 1
		trimmed := strings.TrimSpace(line)

		if strings.HasPrefix(trimmed, "//") || strings.HasPrefix(trimmed, "/*") || trimmed == "" {
			continue
		}

		if matches := useRe.FindStringSubmatch(line); len(matches) > 1 {
			name := matches[1]
			name = strings.TrimSpace(name)
			if strings.Contains(name, "\\") {
				parts := strings.Split(name, "\\")
				name = parts[len(parts)-1]
			}
			imports = append(imports, Import{
				Name:   name,
				File:   "",
				Line:   lineNum,
				Source: matches[1],
			})
		} else if matches := useFunctionRe.FindStringSubmatch(line); len(matches) > 1 {
			name := matches[1]
			name = strings.TrimSpace(name)
			if strings.Contains(name, "\\") {
				parts := strings.Split(name, "\\")
				name = parts[len(parts)-1]
			}
			imports = append(imports, Import{
				Name:   name,
				File:   "",
				Line:   lineNum,
				Source: "function " + matches[1],
			})
		} else if matches := useConstRe.FindStringSubmatch(line); len(matches) > 1 {
			name := matches[1]
			name = strings.TrimSpace(name)
			if strings.Contains(name, "\\") {
				parts := strings.Split(name, "\\")
				name = parts[len(parts)-1]
			}
			imports = append(imports, Import{
				Name:   name,
				File:   "",
				Line:   lineNum,
				Source: "const " + matches[1],
			})
		}
	}

	return imports
}

func findPHPVariables(content string) []Definition {
	var vars []Definition
	lines := strings.Split(content, "\n")

	varRe := regexp.MustCompile(`^\s*\$([a-zA-Z_][a-zA-Z0-9_]*)\s*=`)
	funcRe := regexp.MustCompile(`^\s*function\s+([a-zA-Z_][a-zA-Z0-9_]*)\s*\(`)
	classRe := regexp.MustCompile(`^\s*class\s+([a-zA-Z_][a-zA-Z0-9_]*)\s*`)
	interfaceRe := regexp.MustCompile(`^\s*interface\s+([a-zA-Z_][a-zA-Z0-9_]*)\s*`)
	traitRe := regexp.MustCompile(`^\s*trait\s+([a-zA-Z_][a-zA-Z0-9_]*)\s*`)
	constRe := regexp.MustCompile(`^\s*const\s+([a-zA-Z_][a-zA-Z0-9_]*)\s*=`)

	seen := make(map[string]bool)

	for i, line := range lines {
		lineNum := i + 1
		trimmed := strings.TrimSpace(line)

		if strings.HasPrefix(trimmed, "//") || strings.HasPrefix(trimmed, "/*") || trimmed == "" {
			continue
		}

		if matches := varRe.FindStringSubmatch(line); len(matches) > 1 {
			name := "$" + matches[1]
			if !seen[name] && !isPHPBuiltin(name) {
				seen[name] = true
				vars = append(vars, Definition{
					Name: name,
					File: "",
					Line: lineNum,
					Type: "var",
				})
			}
		}

		if matches := funcRe.FindStringSubmatch(line); len(matches) > 1 {
			name := matches[1]
			if !seen[name] && !isPHPBuiltin(name) {
				seen[name] = true
				vars = append(vars, Definition{
					Name: name,
					File: "",
					Line: lineNum,
					Type: "function",
				})
			}
		}

		if matches := classRe.FindStringSubmatch(line); len(matches) > 1 {
			name := matches[1]
			if !seen[name] {
				seen[name] = true
				vars = append(vars, Definition{
					Name: name,
					File: "",
					Line: lineNum,
					Type: "class",
				})
			}
		}

		if matches := interfaceRe.FindStringSubmatch(line); len(matches) > 1 {
			name := matches[1]
			if !seen[name] {
				seen[name] = true
				vars = append(vars, Definition{
					Name: name,
					File: "",
					Line: lineNum,
					Type: "interface",
				})
			}
		}

		if matches := traitRe.FindStringSubmatch(line); len(matches) > 1 {
			name := matches[1]
			if !seen[name] {
				seen[name] = true
				vars = append(vars, Definition{
					Name: name,
					File: "",
					Line: lineNum,
					Type: "trait",
				})
			}
		}

		if matches := constRe.FindStringSubmatch(line); len(matches) > 1 {
			name := matches[1]
			if !seen[name] {
				seen[name] = true
				vars = append(vars, Definition{
					Name: name,
					File: "",
					Line: lineNum,
					Type: "const",
				})
			}
		}
	}

	return vars
}

func findPHPDefinitions(content, filename string) []Definition {
	return findPHPVariables(content)
}

func findPHPParameters(content string) []CodeIssue {
	return findPHPParametersForWorkspace(content, "")
}

func findPHPParametersForWorkspace(content, filename string) []CodeIssue {
	var params []CodeIssue
	lines := strings.Split(content, "\n")

	funcRe := regexp.MustCompile(`^\s*function\s+[a-zA-Z_][a-zA-Z0-9_]*\s*\(([^)]*)\)`)

	for i, line := range lines {
		lineNum := i + 1

		if matches := funcRe.FindStringSubmatch(line); len(matches) > 1 {
			args := matches[1]
			if args == "" {
				continue
			}

			for _, arg := range strings.Split(args, ",") {
				arg = strings.TrimSpace(arg)
				arg = strings.Split(arg, "=")[0]
				arg = strings.TrimSpace(arg)

				if strings.HasPrefix(arg, "$") && arg != "$this" {
					params = append(params, CodeIssue{
						ID:   generateUUID(),
						Line: lineNum,
						Text: "parameter " + arg,
						File: filename,
					})
				}
			}
		}
	}

	return params
}

func findUsedNamesPHP(content string, imports []Import, variables []Definition) map[string]bool {
	var items []NamedItem
	for _, imp := range imports {
		items = append(items, NamedItem{Name: imp.Name, Line: imp.Line})
	}
	for _, v := range variables {
		items = append(items, NamedItem{Name: v.Name, Line: v.Line})
	}
	return FindUsedNames(content, items)
}

func isPHPBuiltin(name string) bool {
	builtins := map[string]bool{
		"$this": true,
		"echo":  true, "print": true, "printf": true, "sprintf": true,
		"function": true, "class": true, "interface": true, "trait": true, "extends": true, "implements": true,
		"public": true, "private": true, "protected": true, "static": true, "final": true, "abstract": true,
		"new": true, "clone": true, "instanceof": true, "use": true, "namespace": true,
		"if": true, "else": true, "elseif": true, "switch": true, "case": true, "default": true,
		"for": true, "foreach": true, "while": true, "do": true, "break": true, "continue": true,
		"return": true, "yield": true, "throw": true, "try": true, "catch": true, "finally": true,
		"true": true, "false": true, "null": true, "undefined": true,
		"array": true, "list": true, "isset": true, "empty": true, "unset": true,
		"include": true, "include_once": true, "require": true, "require_once": true,
	}
	return builtins[name]
}

func toPHPDefSlice(vars []Definition) []Definition {
	return vars
}

func importsToPHPSlice(imports []Import, filename string) []Import {
	return imports
}

func buildResultPHP(file AnalyzeFile, defs []Definition, imports []Import, usedNames map[string]bool) AnalysisResult {
	var unusedImports []CodeIssue
	for _, imp := range imports {
		key := imp.Name + "@" + file.Filename
		if !usedNames[key] {
			unusedImports = append(unusedImports, CodeIssue{
				ID:   generateUUID(),
				Line: imp.Line,
				Text: "use " + imp.Source + ";",
				File: file.Filename,
			})
		}
	}

	var unusedVars []CodeIssue
	for _, v := range defs {
		key := v.Name + "@" + file.Filename
		if !usedNames[key] {
			unusedVars = append(unusedVars, CodeIssue{
				ID:   generateUUID(),
				Line: v.Line,
				Text: v.Type + " " + v.Name,
				File: file.Filename,
			})
		}
	}

	parameters := findPHPParametersForWorkspace(file.Content, file.Filename)
	paramUsed := findUsedParameterNames(file.Content, parameters)
	var unusedParams []CodeIssue
	for _, p := range parameters {
		paramName := strings.TrimPrefix(p.Text, "parameter ")
		if !paramUsed[paramName] {
			unusedParams = append(unusedParams, CodeIssue{
				ID:   generateUUID(),
				Line: p.Line,
				Text: p.Text,
				File: file.Filename,
			})
		}
	}

	return AnalysisResult{
		Imports:    unusedImports,
		Variables:  unusedVars,
		Parameters: unusedParams,
	}
}
