package main

import (
	"encoding/json"
	"regexp"
	"strings"
	"sync"
	"syscall/js"
)

type CacheEntry struct {
	hash   string
	result AnalysisResult
}

type MultiLangAnalyzer struct {
	mu             sync.Mutex
	cache          map[string]CacheEntry
	allDefinitions map[string][]Definition
	allImports     map[string][]Import
}

type WorkspaceAnalysisResult struct {
	Results map[string]AnalysisResult
}

func NewMultiLangAnalyzer() *MultiLangAnalyzer {
	return &MultiLangAnalyzer{
		cache:          make(map[string]CacheEntry),
		allDefinitions: make(map[string][]Definition),
		allImports:     make(map[string][]Import),
	}
}

func (a *MultiLangAnalyzer) Analyze(req AnalyzeRequest) AnalysisResult {
	a.mu.Lock()
	defer a.mu.Unlock()

	hash := req.Hash
	if hash == "" {
		hash = "default"
	}

	if entry, ok := a.cache[req.Filename]; ok {
		if entry.hash == hash {
			return entry.result
		}
	}

	lang := DetectLanguage(req.Filename)

	var result AnalysisResult
	switch lang {
	case LangPython:
		result = analyzePython(req.Content, req.Filename)
	case LangJavaScript, LangTypeScript:
		result = analyzeJavaScript(req.Content, req.Filename)
	case LangGo:
		result = analyzeGo(req.Content, req.Filename)
	default:
		result = AnalysisResult{}
	}

	a.cache[req.Filename] = CacheEntry{hash: hash, result: result}
	return result
}

func (a *MultiLangAnalyzer) AnalyzeWorkspace(req WorkspaceAnalyzeRequest) WorkspaceAnalysisResult {
	a.mu.Lock()
	defer a.mu.Unlock()

	a.allDefinitions = make(map[string][]Definition)
	a.allImports = make(map[string][]Import)

	for _, file := range req.Files {
		lang := DetectLanguage(file.Filename)

		switch lang {
		case LangJavaScript, LangTypeScript:
			defs, imports, _, _ := analyzeJavaScriptForWorkspace(file.Content, file.Filename)
			a.allDefinitions[file.Filename] = defs
			a.allImports[file.Filename] = imports
		case LangPython:
			defs, imports, _, _ := analyzePythonForWorkspace(file.Content, file.Filename)
			a.allDefinitions[file.Filename] = defs
			a.allImports[file.Filename] = imports
		case LangGo:
			defs, imports, _, _ := analyzeGoForWorkspace(file.Content, file.Filename)
			a.allDefinitions[file.Filename] = defs
			a.allImports[file.Filename] = imports
		}
	}

	usedNames := make(map[string]bool)

	for filename, defs := range a.allDefinitions {
		content := getFileContent(req.Files, filename)

		var defItems []NamedItem
		for _, def := range defs {
			defItems = append(defItems, NamedItem{Name: def.Name, Line: def.Line})
		}
		defUsed := FindUsedNames(content, defItems)
		for name := range defUsed {
			usedNames[name+"@"+filename] = true
		}

		var impItems []NamedItem
		for _, imp := range a.allImports[filename] {
			impItems = append(impItems, NamedItem{Name: imp.Name, Line: imp.Line})
		}
		impUsed := FindUsedNames(content, impItems)
		for name := range impUsed {
			isUsedInOtherFile := isNameUsedInOtherFiles(req.Files, filename, name)
			if isUsedInOtherFile {
				usedNames[name+"@"+filename] = true
			}
		}
	}

	for _, file := range req.Files {
		for _, def := range a.allDefinitions[file.Filename] {
			if def.Exported {
				isUsed := isExportedUsedInOtherFiles(req.Files, file.Filename, def.Name)
				if isUsed {
					usedNames[def.Name+"@"+file.Filename] = true
				}
			}
		}
	}

	results := make(map[string]AnalysisResult)

	for _, file := range req.Files {
		lang := DetectLanguage(file.Filename)

		switch lang {
		case LangJavaScript, LangTypeScript:
			results[file.Filename] = buildResultJS(file, a.allDefinitions[file.Filename], a.allImports[file.Filename], usedNames)
			a.cache[file.Filename] = CacheEntry{hash: file.Hash, result: results[file.Filename]}
		case LangPython:
			results[file.Filename] = buildResultPython(file, a.allDefinitions[file.Filename], a.allImports[file.Filename], usedNames)
			a.cache[file.Filename] = CacheEntry{hash: file.Hash, result: results[file.Filename]}
		case LangGo:
			results[file.Filename] = buildResultGo(file, a.allDefinitions[file.Filename], a.allImports[file.Filename], usedNames)
			a.cache[file.Filename] = CacheEntry{hash: file.Hash, result: results[file.Filename]}
		default:
			results[file.Filename] = AnalysisResult{}
		}
	}

	return WorkspaceAnalysisResult{Results: results}
}

func getFileContent(files []AnalyzeFile, filename string) string {
	for _, f := range files {
		if f.Filename == filename {
			return f.Content
		}
	}
	return ""
}

func isNameUsedInOtherFiles(files []AnalyzeFile, excludeFilename, name string) bool {
	re := regexp.MustCompile(`\b` + regexp.QuoteMeta(name) + `\b`)
	for _, f := range files {
		if f.Filename != excludeFilename && re.MatchString(f.Content) {
			return true
		}
	}
	return false
}

func isExportedUsedInOtherFiles(files []AnalyzeFile, excludeFilename, name string) bool {
	for _, f := range files {
		if f.Filename == excludeFilename {
			continue
		}

		for _, imp := range a.allImports[f.Filename] {
			if imp.Name == name {
				return true
			}
		}
	}
	return false
}

var a *MultiLangAnalyzer

func buildResultJS(file AnalyzeFile, defs []Definition, imports []Import, usedNames map[string]bool) AnalysisResult {
	parameters := findJSParameters(file.Content, file.Filename)

	var unusedImports, unusedVars, unusedParams []CodeIssue

	for _, imp := range imports {
		key := imp.Name + "@" + file.Filename
		if !usedNames[key] {
			unusedImports = append(unusedImports, CodeIssue{
				ID:   generateUUID(),
				Line: imp.Line,
				Text: "import " + imp.Name,
				File: file.Filename,
			})
		}
	}

	for _, v := range defs {
		if v.Exported {
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

	for _, p := range parameters {
		paramName := strings.TrimPrefix(p.Text, "parameter ")
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

func buildResultPython(file AnalyzeFile, defs []Definition, imports []Import, usedNames map[string]bool) AnalysisResult {
	parameters := findPythonParametersFromContent(file.Content, file.Filename)

	var unusedImports, unusedVars, unusedParams []CodeIssue

	for _, imp := range imports {
		key := imp.Name + "@" + file.Filename
		if !usedNames[key] {
			unusedImports = append(unusedImports, CodeIssue{
				ID:   generateUUID(),
				Line: imp.Line,
				Text: "import " + imp.Name,
				File: file.Filename,
			})
		}
	}

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

	for _, p := range parameters {
		paramName := strings.TrimPrefix(p.Text, "parameter ")
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

func buildResultGo(file AnalyzeFile, defs []Definition, imports []Import, usedNames map[string]bool) AnalysisResult {
	parameters := findGoParametersFromContent(file.Content, file.Filename)

	var unusedImports, unusedVars, unusedParams []CodeIssue

	for _, imp := range imports {
		key := imp.Name + "@" + file.Filename
		if !usedNames[key] {
			unusedImports = append(unusedImports, CodeIssue{
				ID:   generateUUID(),
				Line: imp.Line,
				Text: "import " + imp.Source,
				File: file.Filename,
			})
		}
	}

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

	for _, p := range parameters {
		paramName := strings.TrimPrefix(p.Text, "parameter ")
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

var globalAnalyzer *MultiLangAnalyzer

func analyzeCodeWrapper(this js.Value, args []js.Value) interface{} {
	if len(args) < 1 {
		return js.ValueOf(nil)
	}

	var req AnalyzeRequest
	json.Unmarshal([]byte(args[0].String()), &req)

	result := globalAnalyzer.Analyze(req)

	jsonBytes, err := json.Marshal(result)
	if err != nil {
		return js.ValueOf(nil)
	}

	return js.ValueOf(string(jsonBytes))
}

func analyzeWorkspaceWrapper(this js.Value, args []js.Value) interface{} {
	if len(args) < 1 {
		return js.ValueOf(nil)
	}

	var req WorkspaceAnalyzeRequest
	json.Unmarshal([]byte(args[0].String()), &req)

	result := globalAnalyzer.AnalyzeWorkspace(req)

	jsonBytes, err := json.Marshal(result)
	if err != nil {
		return js.ValueOf(nil)
	}

	return js.ValueOf(string(jsonBytes))
}

func detectLanguageWrapper(this js.Value, args []js.Value) interface{} {
	if len(args) < 1 {
		return js.ValueOf(string(LangUnknown))
	}

	filename := args[0].String()
	lang := DetectLanguage(filename)

	return js.ValueOf(string(lang))
}

func main() {
	globalAnalyzer = NewMultiLangAnalyzer()

	js.Global().Set("analyzeCode", js.FuncOf(analyzeCodeWrapper))
	js.Global().Set("analyzeWorkspace", js.FuncOf(analyzeWorkspaceWrapper))
	js.Global().Set("detectLanguage", js.FuncOf(detectLanguageWrapper))

	select {}
}
