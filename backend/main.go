package main

import (
	"encoding/json"
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

const analyzerCacheVersion = "2026-02-20-ruby-php-v3"

type MultiLangAnalyzer struct {
	mu             sync.Mutex
	cache          map[string]CacheEntry
	allDefinitions map[string][]Definition
	allImports     map[string][]Import
	allParameters  map[string][]CodeIssue
	parsedFiles    map[string]ParsedWorkspaceEntry
	workspaceSig   string
	workspaceRes   map[string]AnalysisResult
}

type WorkspaceAnalysisResult struct {
	Results map[string]AnalysisResult
}

type ParsedWorkspaceEntry struct {
	Version     string
	Hash        string
	Lang        Language
	Definitions []Definition
	Imports     []Import
	Parameters  []CodeIssue
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

	var result AnalysisResult
	switch lang {
	case LangPython:
		result = analyzePython(req.Content, req.Filename)
	case LangGo:
		result = analyzeGo(req.Content, req.Filename)
	case LangRuby:
		result = analyzeRuby(req.Content, req.Filename)
	case LangPHP:
		result = analyzePHP(req.Content, req.Filename)
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
	a.allParameters = make(map[string][]CodeIssue)

	for _, file := range req.Files {
		lang := DetectLanguage(file.Filename)

		defs, imports, params := a.getParsedWorkspaceData(file, lang)
		a.allDefinitions[file.Filename] = defs
		a.allImports[file.Filename] = imports
		a.allParameters[file.Filename] = params
	}

	usedNames := make(map[string]bool)

	isNameUsedInOtherFiles := func(currentFilename, name string) bool {
		if name == "" {
			return false
		}
		for _, otherFile := range req.Files {
			if otherFile.Filename == currentFilename {
				continue
			}
			otherContent := removeImportLines(otherFile.Content)
			if containsWord(otherContent, name) {
				return true
			}
		}
		return false
	}

	for _, file := range req.Files {
		for _, def := range a.allDefinitions[file.Filename] {
			if def.Name == "" {
				continue
			}
			usedNames[def.Name+"@"+file.Filename] = isNameUsedInOtherFiles(file.Filename, def.Name)
		}

		for _, imp := range a.allImports[file.Filename] {
			if imp.Name == "" {
				continue
			}

			impNames := strings.Split(imp.Name, ", ")
			for _, impName := range impNames {
				if impName == "" {
					continue
				}
				usedNames[impName+"@"+file.Filename] = isNameUsedInOtherFiles(file.Filename, impName)
			}
		}

		for _, param := range a.allParameters[file.Filename] {
			paramName := strings.TrimSpace(strings.TrimSuffix(param.Text, " (parameter)"))
			if paramName == "" {
				continue
			}
			usedNames[paramName+"@"+file.Filename] = isNameUsedInOtherFiles(file.Filename, paramName)
		}
	}

	results := make(map[string]AnalysisResult)

	for _, file := range req.Files {
		lang := DetectLanguage(file.Filename)

		switch lang {
		case LangPython:
			results[file.Filename] = buildResultPython(file, a.allDefinitions[file.Filename], a.allImports[file.Filename], usedNames, req.Files)
			a.cache[file.Filename] = CacheEntry{hash: file.Hash, result: results[file.Filename]}
		case LangGo:
			results[file.Filename] = buildResultGo(file, a.allDefinitions[file.Filename], a.allImports[file.Filename], usedNames, req.Files)
			a.cache[file.Filename] = CacheEntry{hash: file.Hash, result: results[file.Filename]}
		case LangRuby:
			results[file.Filename] = buildResultRuby(file, a.allDefinitions[file.Filename], a.allImports[file.Filename], usedNames, req.Files)
			a.cache[file.Filename] = CacheEntry{hash: file.Hash, result: results[file.Filename]}
		case LangPHP:
			results[file.Filename] = buildResultPHP(file, a.allDefinitions[file.Filename], a.allImports[file.Filename], usedNames, req.Files)
			a.cache[file.Filename] = CacheEntry{hash: file.Hash, result: results[file.Filename]}
		default:
			results[file.Filename] = AnalysisResult{}
		}
	}

	return WorkspaceAnalysisResult{Results: results}
}

func (a *MultiLangAnalyzer) getParsedWorkspaceData(file AnalyzeFile, lang Language) ([]Definition, []Import, []CodeIssue) {
	if cached, ok := a.parsedFiles[file.Filename]; ok {
		if cached.Version == analyzerCacheVersion && cached.Hash == file.Hash && cached.Lang == lang {
			return cached.Definitions, cached.Imports, cached.Parameters
		}
	}

	var defs []Definition
	var imports []Import
	var params []CodeIssue

	switch lang {
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
		params = []CodeIssue{}
	}

	a.parsedFiles[file.Filename] = ParsedWorkspaceEntry{
		Version:     analyzerCacheVersion,
		Hash:        file.Hash,
		Lang:        lang,
		Definitions: defs,
		Imports:     imports,
		Parameters:  params,
	}
	return defs, imports, params
}

func workspaceSignature(files []AnalyzeFile) string {
	rows := make([]string, 0, len(files))
	for _, f := range files {
		rows = append(rows, f.Filename+"|"+f.Hash)
	}
	sort.Strings(rows)

	h := fnv.New64a()
	h.Write([]byte(analyzerCacheVersion))
	h.Write([]byte{'\n'})
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

func removeImportLines(content string) string {
	lines := strings.Split(content, "\n")
	var result []string
	inGoImportBlock := false

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			if !inGoImportBlock {
				result = append(result, line)
			}
			continue
		}

		if inGoImportBlock {
			if trimmed == ")" {
				inGoImportBlock = false
			}
			continue
		}

		if strings.HasPrefix(trimmed, "import (") {
			inGoImportBlock = true
			continue
		}
		if strings.HasPrefix(trimmed, "import ") {
			continue
		}
		if strings.HasPrefix(trimmed, "from ") && strings.Contains(trimmed, " import ") {
			continue
		}
		if strings.HasPrefix(trimmed, "use ") {
			continue
		}
		if strings.HasPrefix(trimmed, "require ") || strings.HasPrefix(trimmed, "require_relative ") {
			continue
		}
		if strings.HasPrefix(trimmed, "export ") && strings.Contains(trimmed, "from") {
			continue
		}
		result = append(result, line)
	}

	return strings.Join(result, "\n")
}

func main() {
	globalAnalyzer = NewMultiLangAnalyzer()

	js.Global().Set("analyzeCode", js.FuncOf(analyzeCodeWrapper))
	js.Global().Set("analyzeWorkspace", js.FuncOf(analyzeWorkspaceWrapper))
	js.Global().Set("detectLanguage", js.FuncOf(detectLanguageWrapper))

	select {}
}
