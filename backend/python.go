package main

import (
	"github.com/go-python/gpython/ast"
	"github.com/go-python/gpython/parser"
	"github.com/go-python/gpython/py"
)

func analyzePython(content, filename string) AnalysisResult {
	tree, err := parser.ParseString(content, py.ExecMode)
	if err != nil {
		return AnalysisResult{}
	}

	imports := findPythonImportsAST(tree)
	variables := findPythonVariablesAST(tree)
	parameters := findPythonParametersAST(tree)
	used := findUsedNamesPython(content, imports, variables)

	var unusedImports, unusedVars, unusedParams []CodeIssue

	for _, imp := range imports {
		if !used[imp.name] {
			unusedImports = append(unusedImports, CodeIssue{
				ID:   generateUUID(),
				Line: imp.line,
				Text: imp.text,
				File: filename,
			})
		}
	}

	for _, v := range variables {
		if !used[v.name] {
			unusedVars = append(unusedVars, CodeIssue{
				ID:   generateUUID(),
				Line: v.line,
				Text: v.text,
				File: filename,
			})
		}
	}

	for _, p := range parameters {
		if !used[p.name] {
			unusedParams = append(unusedParams, CodeIssue{
				ID:   generateUUID(),
				Line: p.line,
				Text: p.text,
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

func analyzePythonForWorkspace(content, filename string) ([]Definition, []Import, []CodeIssue, []CodeIssue) {
	tree, err := parser.ParseString(content, py.ExecMode)
	if err != nil {
		return []Definition{}, []Import{}, []CodeIssue{}, []CodeIssue{}
	}

	imports := findPythonImportsForWorkspace(tree, filename)
	variables := findPythonDefinitionsForWorkspace(tree, filename)
	parameters := findPythonParametersForWorkspace(tree, filename)
	used := findUsedNamesPython(content, toPyItemSlice(imports), toPyItemSliceVars(variables))

	var unusedVars, unusedParams []CodeIssue

	for _, v := range variables {
		if !used[v.name] {
			unusedVars = append(unusedVars, CodeIssue{
				ID:   generateUUID(),
				Line: v.line,
				Text: v.text,
				File: filename,
			})
		}
	}

	for _, p := range parameters {
		if !used[p.name] {
			unusedParams = append(unusedParams, CodeIssue{
				ID:   generateUUID(),
				Line: p.line,
				Text: p.text,
				File: filename,
			})
		}
	}

	return toDefinitionSlice(variables, filename), imports, unusedVars, unusedParams
}

func toPyItemSlice(imports []Import) []pyItem {
	var items []pyItem
	for _, imp := range imports {
		items = append(items, pyItem{name: imp.Name, line: imp.Line, text: imp.Name})
	}
	return items
}

func toPyItemSliceVars(vars []pyItem) []pyItem {
	return vars
}

func toDefinitionSlice(vars []pyItem, filename string) []Definition {
	var defs []Definition
	for _, v := range vars {
		defs = append(defs, Definition{
			Name: v.name,
			File: filename,
			Line: v.line,
			Type: v.text,
		})
	}
	return defs
}

func findPythonImportsForWorkspace(tree ast.Ast, filename string) []Import {
	var imports []Import
	seen := make(map[string]bool)

	ast.Walk(tree, func(node ast.Ast) bool {
		switch n := node.(type) {
		case *ast.ImportFrom:
			module := string(n.Module)
			for _, alias := range n.Names {
				name := string(alias.Name)
				if name == "" || name == "_" {
					continue
				}
				asName := string(alias.AsName)

				line := n.GetLineno()
				key := name + "|" + module
				if !seen[key] {
					seen[key] = true
					imports = append(imports, Import{
						Name:   name,
						File:   filename,
						Line:   line,
						Source: module,
					})
				}
				if asName != "" {
					key = asName + "|" + module
					if !seen[key] {
						seen[key] = true
						imports = append(imports, Import{
							Name:   asName,
							File:   filename,
							Line:   line,
							Source: module,
						})
					}
				}
			}
		case *ast.Import:
			for _, alias := range n.Names {
				name := string(alias.Name)
				if name == "" || name == "_" {
					continue
				}
				asName := string(alias.AsName)

				line := n.GetLineno()
				key := name + "|"
				if !seen[key] {
					seen[key] = true
					imports = append(imports, Import{
						Name:   name,
						File:   filename,
						Line:   line,
						Source: name,
					})
				}
				if asName != "" {
					key = asName + "|"
					if !seen[key] {
						seen[key] = true
						imports = append(imports, Import{
							Name:   asName,
							File:   filename,
							Line:   line,
							Source: name,
						})
					}
				}
			}
		}
		return true
	})

	return imports
}

func findPythonDefinitionsForWorkspace(tree ast.Ast, filename string) []pyItem {
	var items []pyItem
	seen := make(map[string]bool)

	ast.Walk(tree, func(node ast.Ast) bool {
		switch n := node.(type) {
		case *ast.FunctionDef:
			name := string(n.Name)
			if name != "" && !seen[name] && !isPythonBuiltin(name) {
				seen[name] = true
				items = append(items, pyItem{
					name: name,
					line: n.GetLineno(),
					text: "function",
				})
			}
		case *ast.ClassDef:
			name := string(n.Name)
			if name != "" && !seen[name] && !isPythonBuiltin(name) {
				seen[name] = true
				items = append(items, pyItem{
					name: name,
					line: n.GetLineno(),
					text: "class",
				})
			}
		case *ast.Assign:
			for _, target := range n.Targets {
				if name, ok := target.(*ast.Name); ok {
					id := string(name.Id)
					if id != "" && !seen[id] && !isPythonBuiltin(id) && isValidPyIdent(id) {
						seen[id] = true
						items = append(items, pyItem{
							name: id,
							line: n.GetLineno(),
							text: "var",
						})
					}
				}
			}
		}
		return true
	})

	return items
}

func findPythonParametersForWorkspace(tree ast.Ast, filename string) []pyItem {
	var items []pyItem
	seen := make(map[string]bool)

	ast.Walk(tree, func(node ast.Ast) bool {
		switch n := node.(type) {
		case *ast.FunctionDef:
			if n.Args != nil {
				for _, arg := range n.Args.Args {
					name := string(arg.Arg)
					if name != "" && name != "self" && name != "cls" && !seen[name] {
						seen[name] = true
						items = append(items, pyItem{
							name: name,
							line: n.GetLineno(),
							text: "parameter " + name,
						})
					}
				}
			}
		}
		return true
	})

	return items
}

func findPythonParametersFromContent(content, filename string) []CodeIssue {
	// Use AST-based analysis instead of regex
	return []CodeIssue{}
}

type pyItem struct {
	name string
	line int
	text string
}

func findPythonImportsAST(tree ast.Ast) []pyItem {
	var items []pyItem
	seen := make(map[string]bool)

	ast.Walk(tree, func(node ast.Ast) bool {
		switch n := node.(type) {
		case *ast.ImportFrom:
			module := string(n.Module)
			for _, alias := range n.Names {
				name := string(alias.Name)
				if name == "" || name == "_" {
					continue
				}
				asName := string(alias.AsName)

				line := n.GetLineno()
				text := "from " + module + " import " + name
				if asName != "" {
					text += " as " + asName
				}

				if !seen[name] {
					seen[name] = true
					items = append(items, pyItem{name: name, line: line, text: text})
				}
				if asName != "" && !seen[asName] {
					seen[asName] = true
					items = append(items, pyItem{name: asName, line: line, text: text})
				}
			}
		case *ast.Import:
			for _, alias := range n.Names {
				name := string(alias.Name)
				if name == "" || name == "_" {
					continue
				}
				asName := string(alias.AsName)

				line := n.GetLineno()
				text := "import " + name
				if asName != "" {
					text += " as " + asName
				}

				if !seen[name] {
					seen[name] = true
					items = append(items, pyItem{name: name, line: line, text: text})
				}
				if asName != "" && !seen[asName] {
					seen[asName] = true
					items = append(items, pyItem{name: asName, line: line, text: text})
				}
			}
		}
		return true
	})

	return items
}

func findPythonVariablesAST(tree ast.Ast) []pyItem {
	var items []pyItem
	seen := make(map[string]bool)

	ast.Walk(tree, func(node ast.Ast) bool {
		switch n := node.(type) {
		case *ast.FunctionDef:
			name := string(n.Name)
			if name != "" && !seen[name] && !isPythonBuiltin(name) {
				seen[name] = true
				items = append(items, pyItem{
					name: name,
					line: n.GetLineno(),
					text: "def " + name,
				})
			}
		case *ast.ClassDef:
			name := string(n.Name)
			if name != "" && !seen[name] && !isPythonBuiltin(name) {
				seen[name] = true
				items = append(items, pyItem{
					name: name,
					line: n.GetLineno(),
					text: "class " + name,
				})
			}
		case *ast.Assign:
			for _, target := range n.Targets {
				if name, ok := target.(*ast.Name); ok {
					id := string(name.Id)
					if id != "" && !seen[id] && !isPythonBuiltin(id) && isValidPyIdent(id) {
						seen[id] = true
						items = append(items, pyItem{
							name: id,
							line: n.GetLineno(),
							text: id + " = ...",
						})
					}
				}
			}
		}
		return true
	})

	return items
}

func findPythonParametersAST(tree ast.Ast) []pyItem {
	var items []pyItem
	seen := make(map[string]bool)

	ast.Walk(tree, func(node ast.Ast) bool {
		switch n := node.(type) {
		case *ast.FunctionDef:
			if n.Args != nil {
				for _, arg := range n.Args.Args {
					name := string(arg.Arg)
					if name != "" && name != "self" && name != "cls" && !seen[name] {
						seen[name] = true
						items = append(items, pyItem{
							name: name,
							line: n.GetLineno(),
							text: "parameter " + name,
						})
					}
				}
			}
		}
		return true
	})

	return items
}

func isPythonBuiltin(name string) bool {
	builtins := map[string]bool{
		"True": true, "False": true, "None": true,
		"abs": true, "all": true, "any": true, "bin": true, "bool": true,
		"bytes": true, "callable": true, "chr": true, "classmethod": true,
		"compile": true, "complex": true, "delattr": true, "dict": true,
		"dir": true, "divmod": true, "enumerate": true, "eval": true,
		"exec": true, "filter": true, "float": true, "format": true,
		"frozenset": true, "getattr": true, "globals": true, "hasattr": true,
		"hash": true, "help": true, "hex": true, "id": true, "input": true,
		"int": true, "isinstance": true, "issubclass": true, "iter": true,
		"len": true, "list": true, "locals": true, "map": true, "max": true,
		"memoryview": true, "min": true, "next": true, "object": true,
		"oct": true, "open": true, "ord": true, "pow": true, "print": true,
		"property": true, "range": true, "repr": true, "reversed": true,
		"round": true, "set": true, "setattr": true, "slice": true,
		"sorted": true, "staticmethod": true, "str": true, "sum": true,
		"super": true, "tuple": true, "type": true, "vars": true, "zip": true,
		"__import__": true,
	}
	return builtins[name]
}

func isValidPyIdent(s string) bool {
	if s == "" {
		return false
	}

	// Check first character
	first := s[0]
	if !((first >= 'a' && first <= 'z') || (first >= 'A' && first <= 'Z') || first == '_') {
		return false
	}

	// Check remaining characters
	for i := 1; i < len(s); i++ {
		c := s[i]
		if !((c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || (c >= '0' && c <= '9') || c == '_') {
			return false
		}
	}

	return true
}

func findUsedNamesPython(content string, imports []pyItem, variables []pyItem) map[string]bool {
	var items []NamedItem
	for _, item := range imports {
		items = append(items, NamedItem{Name: item.name, Line: item.line})
	}
	for _, item := range variables {
		items = append(items, NamedItem{Name: item.name, Line: item.line})
	}
	return FindUsedNames(content, items)
}
