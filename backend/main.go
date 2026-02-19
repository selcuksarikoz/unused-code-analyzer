package main

import (
	"encoding/json"
	"fmt"
	"hash/fnv"
	"sort"
	"strconv"
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
	parsedFiles    map[string]ParsedWorkspaceEntry
	workspaceSig   string
	workspaceRes   map[string]AnalysisResult
}

type WorkspaceAnalysisResult struct {
	Results map[string]AnalysisResult
}

type ParsedWorkspaceEntry struct {
	Hash        string
	Lang        Language
	Definitions []Definition
	Imports     []Import
}

func NewMultiLangAnalyzer() *MultiLangAnalyzer {
	return &MultiLangAnalyzer{
		cache:          make(map[string]CacheEntry),
		allDefinitions: make(map[string][]Definition),
		allImports:     make(map[string][]Import),
		parsedFiles:    make(map[string]ParsedWorkspaceEntry),
		workspaceRes:   make(map[string]AnalysisResult),
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
	fmt.Printf("[Analyzer] Analyzing file: %s, language: %s\n", req.Filename, lang)

	var result AnalysisResult
	switch lang {
	case LangJavaScript, LangTypeScript:
		result = analyzeJSTS(req.Content, req.Filename)
	case LangPython:
		result = analyzePython(req.Content, req.Filename)
	case LangGo:
		result = analyzeGo(req.Content, req.Filename)
	case LangRuby:
		result = analyzeRuby(req.Content, req.Filename)
	case LangPHP:
		result = analyzePHP(req.Content, req.Filename)
	default:
		fmt.Printf("[Analyzer] Unknown language for: %s\n", req.Filename)
		result = AnalysisResult{}
	}

	a.cache[req.Filename] = CacheEntry{hash: hash, result: result}
	return result
}

func (a *MultiLangAnalyzer) AnalyzeWorkspace(req WorkspaceAnalyzeRequest) WorkspaceAnalysisResult {
	a.mu.Lock()
	defer a.mu.Unlock()

	fmt.Printf("[Analyzer] AnalyzeWorkspace called with %d files\n", len(req.Files))
	sig := workspaceSignature(req.Files)
	if sig == a.workspaceSig && len(a.workspaceRes) > 0 {
		fmt.Printf("[Analyzer] Workspace cache hit for signature %s\n", sig)
		return WorkspaceAnalysisResult{Results: cloneAnalysisResults(a.workspaceRes)}
	}

	a.allDefinitions = make(map[string][]Definition)
	a.allImports = make(map[string][]Import)

	for _, file := range req.Files {
		lang := DetectLanguage(file.Filename)
		fmt.Printf("[Analyzer] Workspace file: %s, lang: %s\n", file.Filename, lang)

		defs, imports := a.getParsedWorkspaceData(file, lang)
		a.allDefinitions[file.Filename] = defs
		a.allImports[file.Filename] = imports
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
				isUsed := a.isExportedUsedInOtherFiles(req.Files, file.Filename, def.Name)
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
			results[file.Filename] = buildResultJSTS(file, a.allDefinitions[file.Filename], a.allImports[file.Filename], usedNames)
			a.cache[file.Filename] = CacheEntry{hash: file.Hash, result: results[file.Filename]}
		case LangPython:
			results[file.Filename] = buildResultPython(file, a.allDefinitions[file.Filename], a.allImports[file.Filename], usedNames)
			a.cache[file.Filename] = CacheEntry{hash: file.Hash, result: results[file.Filename]}
		case LangGo:
			results[file.Filename] = buildResultGo(file, a.allDefinitions[file.Filename], a.allImports[file.Filename], usedNames)
			a.cache[file.Filename] = CacheEntry{hash: file.Hash, result: results[file.Filename]}
		case LangRuby:
			results[file.Filename] = buildResultRuby(file, a.allDefinitions[file.Filename], a.allImports[file.Filename], usedNames)
			a.cache[file.Filename] = CacheEntry{hash: file.Hash, result: results[file.Filename]}
		case LangPHP:
			results[file.Filename] = buildResultPHP(file, a.allDefinitions[file.Filename], a.allImports[file.Filename], usedNames)
			a.cache[file.Filename] = CacheEntry{hash: file.Hash, result: results[file.Filename]}
		default:
			results[file.Filename] = AnalysisResult{}
		}
	}

	a.workspaceSig = sig
	a.workspaceRes = cloneAnalysisResults(results)
	return WorkspaceAnalysisResult{Results: results}
}

func (a *MultiLangAnalyzer) getParsedWorkspaceData(file AnalyzeFile, lang Language) ([]Definition, []Import) {
	if cached, ok := a.parsedFiles[file.Filename]; ok {
		if cached.Hash == file.Hash && cached.Lang == lang {
			return cached.Definitions, cached.Imports
		}
	}

	var defs []Definition
	var imports []Import

	switch lang {
	case LangJavaScript, LangTypeScript:
		defs, imports, _, _ = analyzeJSTSForWorkspace(file.Content, file.Filename)
	case LangPython:
		defs, imports, _, _ = analyzePythonForWorkspace(file.Content, file.Filename)
	case LangGo:
		defs, imports, _, _ = analyzeGoForWorkspace(file.Content, file.Filename)
	case LangRuby:
		defs, imports, _, _ = analyzeRubyForWorkspace(file.Content, file.Filename)
	case LangPHP:
		defs, imports, _, _ = analyzePHPForWorkspace(file.Content, file.Filename)
	default:
		defs = []Definition{}
		imports = []Import{}
	}

	a.parsedFiles[file.Filename] = ParsedWorkspaceEntry{
		Hash:        file.Hash,
		Lang:        lang,
		Definitions: defs,
		Imports:     imports,
	}
	return defs, imports
}

func workspaceSignature(files []AnalyzeFile) string {
	rows := make([]string, 0, len(files))
	for _, f := range files {
		rows = append(rows, f.Filename+"|"+f.Hash)
	}
	sort.Strings(rows)

	h := fnv.New64a()
	for _, r := range rows {
		h.Write([]byte(r))
		h.Write([]byte{'\n'})
	}
	return strconv.FormatUint(h.Sum64(), 16)
}

func cloneAnalysisResults(in map[string]AnalysisResult) map[string]AnalysisResult {
	out := make(map[string]AnalysisResult, len(in))
	for k, v := range in {
		out[k] = AnalysisResult{
			Imports:    append([]CodeIssue(nil), v.Imports...),
			Variables:  append([]CodeIssue(nil), v.Variables...),
			Parameters: append([]CodeIssue(nil), v.Parameters...),
		}
	}
	return out
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
	for _, f := range files {
		if f.Filename != excludeFilename && containsWord(f.Content, name) {
			return true
		}
	}
	return false
}

func (a *MultiLangAnalyzer) isExportedUsedInOtherFiles(files []AnalyzeFile, excludeFilename, name string) bool {
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
