package main

import (
	"regexp"
	"strings"
)

func analyzeRuby(content, filename string) AnalysisResult {
	imports := findRubyImports(content)
	variables := findRubyVariables(content)
	parameters := findRubyParameters(content)
	used := findUsedNamesRuby(content, imports, variables)
	paramUsed := findUsedParameterNames(content, parameters)

	var unusedImports, unusedVars, unusedParams []CodeIssue

	for _, imp := range imports {
		if !used[imp.Name] {
			unusedImports = append(unusedImports, CodeIssue{
				ID:   generateUUID(),
				Line: imp.Line,
				Text: "require '" + imp.Source + "'",
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

func analyzeRubyForWorkspace(content, filename string) ([]Definition, []Import, []CodeIssue, []CodeIssue) {
	imports := findRubyImports(content)
	variables := findRubyDefinitions(content, filename)
	parameters := findRubyParametersForWorkspace(content, filename)
	used := findUsedNamesRuby(content, imports, toRubyDefSlice(variables))
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

	return toRubyDefSlice(variables), importsToRubySlice(imports, filename), unusedVars, unusedParams
}

func findRubyImports(content string) []Import {
	var imports []Import
	lines := strings.Split(content, "\n")

	importRe := regexp.MustCompile(`^\s*require\s+['"]([^'"]+)['"]\s*$`)
	requireRelRe := regexp.MustCompile(`^\s*require_relative\s+['"]([^'"]+)['"]\s*$`)
	requireDotRe := regexp.MustCompile(`^\s*require\s+\.([^'\"]+)\s*$`)

	for i, line := range lines {
		lineNum := i + 1

		if matches := importRe.FindStringSubmatch(line); len(matches) > 1 {
			imports = append(imports, Import{
				Name:   matches[1],
				File:   "",
				Line:   lineNum,
				Source: matches[1],
			})
		} else if matches := requireRelRe.FindStringSubmatch(line); len(matches) > 1 {
			imports = append(imports, Import{
				Name:   matches[1],
				File:   "",
				Line:   lineNum,
				Source: matches[1],
			})
		} else if matches := requireDotRe.FindStringSubmatch(line); len(matches) > 1 {
			imports = append(imports, Import{
				Name:   matches[1],
				File:   "",
				Line:   lineNum,
				Source: matches[1],
			})
		}
	}

	return imports
}

func findRubyVariables(content string) []Definition {
	var vars []Definition
	lines := strings.Split(content, "\n")

	// Match: variable_name = value
	assignRe := regexp.MustCompile(`^\s*([a-z_][a-zA-Z0-9_]*)\s*=\s*.+\s*$`)
	// Match: def method_name
	defRe := regexp.MustCompile(`^\s*def\s+([a-z_][a-zA-Z0-9_]*)\s*`)
	// Match: class ClassName
	classRe := regexp.MustCompile(`^\s*class\s+([A-Z][a-zA-Z0-9_]*)\s*`)
	// Match: CONSTANT_NAME = value
	constRe := regexp.MustCompile(`^\s*([A-Z][A-Z0-9_]*)\s*=\s*.+\s*$`)

	seen := make(map[string]bool)

	for i, line := range lines {
		lineNum := i + 1
		trimmed := strings.TrimSpace(line)

		// Skip comments and empty lines
		if strings.HasPrefix(trimmed, "#") || trimmed == "" {
			continue
		}

		if matches := assignRe.FindStringSubmatch(line); len(matches) > 1 {
			name := matches[1]
			if !seen[name] && !isRubyBuiltin(name) && !strings.HasPrefix(name, "_") {
				seen[name] = true
				vars = append(vars, Definition{
					Name: name,
					File: "",
					Line: lineNum,
					Type: "var",
				})
			}
		}

		if matches := defRe.FindStringSubmatch(line); len(matches) > 1 {
			name := matches[1]
			if !seen[name] && !isRubyBuiltin(name) {
				seen[name] = true
				vars = append(vars, Definition{
					Name: name,
					File: "",
					Line: lineNum,
					Type: "def",
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

func findRubyDefinitions(content, filename string) []Definition {
	return findRubyVariables(content)
}

func findRubyParameters(content string) []CodeIssue {
	return findRubyParametersForWorkspace(content, "")
}

func findRubyParametersForWorkspace(content, filename string) []CodeIssue {
	var params []CodeIssue
	lines := strings.Split(content, "\n")

	defRe := regexp.MustCompile(`^\s*def\s+[a-zA-Z_][a-zA-Z0-9_]*\s*\(([^)]*)\)`)

	for i, line := range lines {
		lineNum := i + 1

		if matches := defRe.FindStringSubmatch(line); len(matches) > 1 {
			args := matches[1]
			if args == "" {
				continue
			}

			for _, arg := range strings.Split(args, ",") {
				arg = strings.TrimSpace(arg)
				// Skip block args and default values
				if strings.Contains(arg, "&") || strings.Contains(arg, "*") {
					continue
				}
				arg = strings.Split(arg, "=")[0]
				arg = strings.TrimSpace(arg)

				if arg != "" && arg != "self" {
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

func findUsedNamesRuby(content string, imports []Import, variables []Definition) map[string]bool {
	var items []NamedItem
	for _, imp := range imports {
		items = append(items, NamedItem{Name: imp.Name, Line: imp.Line})
	}
	for _, v := range variables {
		items = append(items, NamedItem{Name: v.Name, Line: v.Line})
	}
	return FindUsedNames(content, items)
}

func isRubyBuiltin(name string) bool {
	builtins := map[string]bool{
		"nil": true, "true": true, "false": true, "self": true, "super": true,
		"puts": true, "print": true, "printf": true, "gets": true, "chomp": true,
		"each": true, "map": true, "select": true, "reject": true, "find": true,
		"include": true, "extend": true, "prepend": true, "private": true, "protected": true, "public": true,
		"attr_reader": true, "attr_writer": true, "attr_accessor": true,
		"require": true, "require_relative": true,
		"new": true, "class": true, "module": true, "def": true, "end": true,
		"if": true, "else": true, "elsif": true, "unless": true, "case": true, "when": true,
		"while": true, "until": true, "for": true, "do": true, "begin": true, "rescue": true, "ensure": true,
		"raise": true, "return": true, "break": true, "next": true, "redo": true, "retry": true,
		"yield": true, "block_given?": true, "lambda": true, "proc": true,
	}
	return builtins[name]
}

func toRubyDefSlice(vars []Definition) []Definition {
	return vars
}

func importsToRubySlice(imports []Import, filename string) []Import {
	return imports
}

func buildResultRuby(file AnalyzeFile, defs []Definition, imports []Import, usedNames map[string]bool) AnalysisResult {
	var unusedImports []CodeIssue
	for _, imp := range imports {
		key := imp.Name + "@" + file.Filename
		if !usedNames[key] {
			unusedImports = append(unusedImports, CodeIssue{
				ID:   generateUUID(),
				Line: imp.Line,
				Text: "require '" + imp.Source + "'",
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

	parameters := findRubyParametersForWorkspace(file.Content, file.Filename)
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
