package main

import (
	"go/ast"
	"go/parser"
	"go/token"
	"strings"
)

func analyzeGo(content, filename string) AnalysisResult {
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, filename, content, parser.ParseComments)
	if err != nil {
		return AnalysisResult{}
	}

	imports := findGoImportsAST(f)
	variables := findGoVariablesAST(f)
	parameters := findGoParametersAST(f, fset)

	usedNames := findUsedGoNames(content, imports, variables)

	unusedImports := filterUnusedImports(imports, usedNames)
	unusedVars := filterUnusedVars(variables, usedNames)
	unusedParams := filterUnusedParams(parameters, usedNames)

	return AnalysisResult{
		Imports:    unusedImports,
		Variables:  unusedVars,
		Parameters: unusedParams,
	}
}

func analyzeGoForWorkspace(content, filename string) ([]Definition, []Import, []CodeIssue, []CodeIssue) {
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, filename, content, parser.ParseComments)
	if err != nil {
		return []Definition{}, []Import{}, []CodeIssue{}, []CodeIssue{}
	}

	imports := findGoImportsForWorkspace(f, fset, filename)
	variables := findGoDefinitionsForWorkspace(f, fset, filename)
	parameters := findGoParametersForWorkspace(f, fset, filename)

	usedNames := findUsedGoNames(content, toGoImportSlice(imports), toGoVarSlice(variables))

	unusedVars := filterUnusedVars(toGoVarSlice(variables), usedNames)
	unusedParams := filterUnusedParams(toGoParamSlice(parameters), usedNames)

	return toDefinitionSliceGo(variables, filename), imports, unusedVars, unusedParams
}

func findGoImportsForWorkspace(f *ast.File, fset *token.FileSet, filename string) []Import {
	var imports []Import
	seen := make(map[string]bool)

	ast.Inspect(f, func(n ast.Node) bool {
		if imp, ok := n.(*ast.ImportSpec); ok {
			name := ""
			isBlank := false

			if imp.Name != nil {
				if imp.Name.Name == "_" {
					isBlank = true
				} else {
					name = imp.Name.Name
				}
			}

			path := strings.Trim(imp.Path.Value, `"`)
			parts := strings.Split(path, "/")
			pkgName := parts[len(parts)-1]

			if name == "" {
				name = pkgName
			}

			key := name + "|" + path
			if !isBlank && !seen[key] {
				seen[key] = true
				imports = append(imports, Import{
					Name:   name,
					File:   filename,
					Line:   fset.Position(imp.Pos()).Line,
					Source: path,
				})
			}
		}
		return true
	})

	return imports
}

func findGoDefinitionsForWorkspace(f *ast.File, fset *token.FileSet, filename string) []goVar {
	var variables []goVar

	ast.Inspect(f, func(n ast.Node) bool {
		switch decl := n.(type) {
		case *ast.GenDecl:
			if decl.Tok == token.VAR || decl.Tok == token.CONST {
				for _, spec := range decl.Specs {
					if vspec, ok := spec.(*ast.ValueSpec); ok {
						for _, name := range vspec.Names {
							if name.Name != "_" {
								variables = append(variables, goVar{
									name: name.Name,
									line: fset.Position(name.Pos()).Line,
								})
							}
						}
					}
				}
			}
		case *ast.FuncDecl:
			if decl.Name != nil && decl.Name.Name != "" {
				isExported := len(decl.Name.Name) > 0 && decl.Name.Name[0] >= 'A' && decl.Name.Name[0] <= 'Z'
				if isExported {
					variables = append(variables, goVar{
						name: decl.Name.Name,
						line: fset.Position(decl.Name.Pos()).Line,
					})
				}
			}
		}
		return true
	})

	return variables
}

func findGoParametersForWorkspace(f *ast.File, fset *token.FileSet, filename string) []goParam {
	var params []goParam

	ast.Inspect(f, func(n ast.Node) bool {
		if fn, ok := n.(*ast.FuncDecl); ok {
			if fn.Type.Params != nil {
				for _, param := range fn.Type.Params.List {
					for _, name := range param.Names {
						if name.Name != "_" {
							params = append(params, goParam{
								name: name.Name,
								line: fset.Position(name.Pos()).Line,
							})
						}
					}
				}
			}
		}
		return true
	})

	return params
}

func toGoImportSlice(imports []Import) []goImport {
	var result []goImport
	for _, imp := range imports {
		result = append(result, goImport{
			name:  imp.Name,
			path:  imp.Source,
			line:  imp.Line,
			alias: imp.Name,
		})
	}
	return result
}

func toGoVarSlice(vars []goVar) []goVar {
	return vars
}

func toGoParamSlice(params []goParam) []goParam {
	return params
}

func toDefinitionSliceGo(vars []goVar, filename string) []Definition {
	var defs []Definition
	for _, v := range vars {
		defs = append(defs, Definition{
			Name: v.name,
			File: filename,
			Line: v.line,
			Type: "var",
		})
	}
	return defs
}

type goImport struct {
	name    string
	path    string
	line    int
	alias   string
	isBlank bool
}

func findGoImportsAST(f *ast.File) []goImport {
	var imports []goImport
	fset := token.NewFileSet()

	ast.Inspect(f, func(n ast.Node) bool {
		if imp, ok := n.(*ast.ImportSpec); ok {
			name := ""
			isBlank := false

			if imp.Name != nil {
				if imp.Name.Name == "_" {
					isBlank = true
				} else {
					name = imp.Name.Name
				}
			}

			path := strings.Trim(imp.Path.Value, `"`)
			parts := strings.Split(path, "/")
			pkgName := parts[len(parts)-1]

			if name == "" {
				name = pkgName
			}

			imports = append(imports, goImport{
				name:    name,
				path:    path,
				line:    fset.Position(imp.Pos()).Line,
				alias:   name,
				isBlank: isBlank,
			})
		}
		return true
	})

	return imports
}

type goVar struct {
	name string
	line int
}

func findGoVariablesAST(f *ast.File) []goVar {
	var variables []goVar
	fset := token.NewFileSet()

	ast.Inspect(f, func(n ast.Node) bool {
		switch decl := n.(type) {
		case *ast.GenDecl:
			if decl.Tok == token.VAR || decl.Tok == token.CONST {
				for _, spec := range decl.Specs {
					if vspec, ok := spec.(*ast.ValueSpec); ok {
						for _, name := range vspec.Names {
							if name.Name != "_" {
								variables = append(variables, goVar{
									name: name.Name,
									line: fset.Position(name.Pos()).Line,
								})
							}
						}
					}
				}
			}
		case *ast.FuncDecl:
			if decl.Name != nil && decl.Name.Name != "" {
				variables = append(variables, goVar{
					name: decl.Name.Name,
					line: fset.Position(decl.Name.Pos()).Line,
				})
			}
		}
		return true
	})

	return variables
}

type goParam struct {
	name string
	line int
}

func findGoParametersAST(f *ast.File, fset *token.FileSet) []goParam {
	var params []goParam

	ast.Inspect(f, func(n ast.Node) bool {
		if fn, ok := n.(*ast.FuncDecl); ok {
			if fn.Type.Params != nil {
				for _, param := range fn.Type.Params.List {
					for _, name := range param.Names {
						if name.Name != "_" {
							params = append(params, goParam{
								name: name.Name,
								line: fset.Position(name.Pos()).Line,
							})
						}
					}
				}
			}
		}
		return true
	})

	return params
}

func findUsedGoNames(content string, imports []goImport, variables []goVar) map[string]bool {
	used := make(map[string]bool)

	var items []NamedItem
	for _, imp := range imports {
		if !imp.isBlank {
			items = append(items, NamedItem{Name: imp.name, Line: imp.line})
		}
	}
	for _, v := range variables {
		items = append(items, NamedItem{Name: v.name, Line: v.line})
	}

	used = FindUsedNames(content, items)

	return used
}

func filterUnusedImports(imports []goImport, used map[string]bool) []CodeIssue {
	var issues []CodeIssue

	for _, imp := range imports {
		if imp.isBlank {
			continue
		}

		if !used[imp.name] && !used[imp.alias] {
			issues = append(issues, CodeIssue{
				ID:   generateUUID(),
				Line: imp.line,
				Text: "import " + imp.path,
				File: "",
			})
		}
	}

	return issues
}

func filterUnusedVars(variables []goVar, used map[string]bool) []CodeIssue {
	var issues []CodeIssue

	for _, v := range variables {
		if !used[v.name] {
			issues = append(issues, CodeIssue{
				ID:   generateUUID(),
				Line: v.line,
				Text: "var/const " + v.name,
				File: "",
			})
		}
	}

	return issues
}

func filterUnusedParams(params []goParam, used map[string]bool) []CodeIssue {
	var issues []CodeIssue

	for _, p := range params {
		if !used[p.name] {
			issues = append(issues, CodeIssue{
				ID:   generateUUID(),
				Line: p.line,
				Text: "parameter " + p.name,
				File: "",
			})
		}
	}

	return issues
}

func findGoParametersFromContent(content, filename string) []CodeIssue {
	// Use AST-based analysis instead of regex
	return []CodeIssue{}
}
