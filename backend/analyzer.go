package main

import "strings"

type ImportChecker func(content string, imports []Import) map[string]bool
type ParameterFinder func(content, filename string) []CodeIssue
type FrameworkExportChecker func(name, filename string) bool

type AnalyzerConfig struct {
	ImportTextPrefix     string
	CheckImport          ImportChecker
	FindParameters       ParameterFinder
	CheckFrameworkExport FrameworkExportChecker
}

func BuildAnalysisResult(file AnalyzeFile, defs []Definition, imports []Import, usedNames map[string]bool, config AnalyzerConfig) AnalysisResult {
	parameters := config.FindParameters(file.Content, file.Filename)
	localImportUsed := config.ImportTextPrefix != "" && config.CheckImport != nil

	var importUsed map[string]bool
	if localImportUsed {
		importUsed = config.CheckImport(file.Content, imports)
	}

	var unusedImports []CodeIssue
	for _, imp := range imports {
		isUsed := false
		if localImportUsed && importUsed != nil {
			isUsed = importUsed[imp.Name]
		}
		if !isUsed {
			importText := config.ImportTextPrefix + " " + imp.Name
			if imp.Source != "" && config.ImportTextPrefix == "use" {
				importText = config.ImportTextPrefix + " " + imp.Source
			}
			unusedImports = append(unusedImports, CodeIssue{
				ID:   generateUUID(),
				Line: imp.Line,
				Text: importText,
				File: file.Filename,
			})
		}
	}

	var unusedVars []CodeIssue
	for _, v := range defs {
		if config.CheckFrameworkExport != nil && config.CheckFrameworkExport(v.Name, file.Filename) {
			continue
		}
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

	var unusedParams []CodeIssue
	for _, p := range parameters {
		paramName := strings.TrimSpace(strings.TrimPrefix(p.Text, "parameter "))
		key := paramName + "@" + file.Filename
		if !usedNames[key] {
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

func BuildAnalysisResultWithLocalParams(file AnalyzeFile, defs []Definition, imports []Import, usedNames map[string]bool, config AnalyzerConfig, localParamChecker func(content string, params []CodeIssue) map[string]bool) AnalysisResult {
	parameters := config.FindParameters(file.Content, file.Filename)
	localImportUsed := config.ImportTextPrefix != "" && config.CheckImport != nil

	var importUsed map[string]bool
	if localImportUsed {
		importUsed = config.CheckImport(file.Content, imports)
	}

	var unusedImports []CodeIssue
	for _, imp := range imports {
		isUsed := false
		if localImportUsed && importUsed != nil {
			isUsed = importUsed[imp.Name]
		}
		if !isUsed {
			importText := config.ImportTextPrefix + " " + imp.Name
			if imp.Source != "" && config.ImportTextPrefix == "use" {
				importText = config.ImportTextPrefix + " " + imp.Source
			}
			unusedImports = append(unusedImports, CodeIssue{
				ID:   generateUUID(),
				Line: imp.Line,
				Text: importText,
				File: file.Filename,
			})
		}
	}

	var unusedVars []CodeIssue
	for _, v := range defs {
		if config.CheckFrameworkExport != nil && config.CheckFrameworkExport(v.Name, file.Filename) {
			continue
		}
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

	paramUsed := localParamChecker(file.Content, parameters)
	var unusedParams []CodeIssue
	for _, p := range parameters {
		paramName := strings.TrimSpace(strings.TrimPrefix(p.Text, "parameter "))
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
