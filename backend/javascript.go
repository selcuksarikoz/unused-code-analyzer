package main

import (
	"regexp"
	"strconv"
	"strings"
)

type jsImportItem struct {
	name string
	line int
	text string
}

type jsVarItem struct {
	name     string
	line     int
	kind     string
	exported bool
}

type jsParamItem struct {
	name string
	line int
	text string
}

var (
	reJSImportFrom       = regexp.MustCompile(`(?m)^\s*import\s+([^;\n]+?)\s+from\s+['"][^'"]+['"]`)
	reJSImportRequire    = regexp.MustCompile(`(?m)^\s*(?:const|let|var)\s+([A-Za-z_$][\w$]*)\s*=\s*require\(['"][^'"]+['"]\)`)
	reJSImportReqDestr   = regexp.MustCompile(`(?m)^\s*(?:const|let|var)\s*\{([^}]+)\}\s*=\s*require\(['"][^'"]+['"]\)`)
	reJSExportNamedFrom  = regexp.MustCompile(`(?m)^\s*export\s*\{([^}]+)\}\s*from\s*['"][^'"]+['"]`)
	reJSExportNsFrom     = regexp.MustCompile(`(?m)^\s*export\s+\*\s+as\s+([A-Za-z_$][\w$]*)\s+from\s+['"][^'"]+['"]`)
	reJSFunctionDecl     = regexp.MustCompile(`(?m)^\s*(export\s+)?(?:default\s+)?(?:async\s+)?function\s+([A-Za-z_$][\w$]*)\s*\(([^)]*)\)`)
	reJSClassDecl        = regexp.MustCompile(`(?m)^\s*(export\s+)?(?:default\s+)?class\s+([A-Za-z_$][\w$]*)\b`)
	reJSVarDeclLine      = regexp.MustCompile(`(?m)^\s*(export\s+)?(?:const|let|var)\s+([^;\n]+)`)
	reJSArrowAssign      = regexp.MustCompile(`(?m)^\s*(export\s+)?(?:const|let|var)\s+([A-Za-z_$][\w$]*)\s*=\s*(?:async\s*)?(?:\(([^)]*)\)|([A-Za-z_$][\w$]*))\s*=>`)
	reJSFuncExprAssign   = regexp.MustCompile(`(?m)^\s*(export\s+)?(?:const|let|var)\s+([A-Za-z_$][\w$]*)\s*=\s*(?:async\s*)?function(?:\s+[A-Za-z_$][\w$]*)?\s*\(([^)]*)\)`)
	reJSIdentifier       = regexp.MustCompile(`^[A-Za-z_$][\w$]*`)
	reJSDefaultImportPre = regexp.MustCompile(`^\s*([A-Za-z_$][\w$]*)\s*(?:,|$)`)
)

func analyzeJSTS(content, filename string) AnalysisResult {
	imports := findJSTSImports(content)
	defs := findJSTSDefinitions(content)
	params := findJSTSParameters(content)

	var named []NamedItem
	for _, imp := range imports {
		named = append(named, NamedItem{Name: imp.name, Line: imp.line})
	}
	for _, d := range defs {
		named = append(named, NamedItem{Name: d.name, Line: d.line})
	}
	for _, p := range params {
		named = append(named, NamedItem{Name: p.name, Line: p.line})
	}

	used := FindUsedNames(content, named)
	importUsed := findUsedJSImportNames(content, imports)

	var unusedImports []CodeIssue
	for _, imp := range imports {
		if !used[imp.name] && !importUsed[imp.name] {
			unusedImports = append(unusedImports, CodeIssue{
				ID:   generateUUID(),
				Line: imp.line,
				Text: imp.text,
				File: filename,
			})
		}
	}

	var unusedVars []CodeIssue
	for _, d := range defs {
		if d.exported && isFrameworkSpecialExport(d.name, filename) {
			continue
		}
		if !used[d.name] {
			unusedVars = append(unusedVars, CodeIssue{
				ID:   generateUUID(),
				Line: d.line,
				Text: d.kind + " " + d.name,
				File: filename,
			})
		}
	}

	var unusedParams []CodeIssue
	for _, p := range params {
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

func analyzeJSTSForWorkspace(content, filename string) ([]Definition, []Import, []CodeIssue, []CodeIssue) {
	imports := findJSTSImports(content)
	defs := findJSTSDefinitions(content)
	params := findJSTSParameters(content)

	var outDefs []Definition
	for _, d := range defs {
		outDefs = append(outDefs, Definition{
			Name:     d.name,
			File:     filename,
			Line:     d.line,
			Type:     d.kind,
			Exported: d.exported,
		})
	}

	var outImports []Import
	for _, imp := range imports {
		outImports = append(outImports, Import{
			Name: imp.name,
			File: filename,
			Line: imp.line,
		})
	}

	var named []NamedItem
	for _, d := range defs {
		named = append(named, NamedItem{Name: d.name, Line: d.line})
	}
	for _, p := range params {
		named = append(named, NamedItem{Name: p.name, Line: p.line})
	}
	used := FindUsedNames(content, named)

	var unusedVars []CodeIssue
	for _, d := range defs {
		if d.exported && isFrameworkSpecialExport(d.name, filename) {
			continue
		}
		if !used[d.name] {
			unusedVars = append(unusedVars, CodeIssue{
				ID:   generateUUID(),
				Line: d.line,
				Text: d.kind + " " + d.name,
				File: filename,
			})
		}
	}

	var unusedParams []CodeIssue
	for _, p := range params {
		if !used[p.name] {
			unusedParams = append(unusedParams, CodeIssue{
				ID:   generateUUID(),
				Line: p.line,
				Text: p.text,
				File: filename,
			})
		}
	}

	return outDefs, outImports, []CodeIssue{}, unusedParams
}

func buildResultJSTS(file AnalyzeFile, defs []Definition, imports []Import, usedNames map[string]bool) AnalysisResult {
	localImports := findJSTSImports(file.Content)
	localImportUsed := findUsedJSImportNames(file.Content, localImports)

	var unusedImports []CodeIssue
	for _, imp := range imports {
		key := imp.Name + "@" + file.Filename
		if !usedNames[key] && !localImportUsed[imp.Name] {
			unusedImports = append(unusedImports, CodeIssue{
				ID:   generateUUID(),
				Line: imp.Line,
				Text: "import " + imp.Name,
				File: file.Filename,
			})
		}
	}

	var unusedVars []CodeIssue
	for _, v := range defs {
		if v.Exported && isFrameworkSpecialExport(v.Name, file.Filename) {
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

	params := findJSTSParameters(file.Content)
	var pItems []NamedItem
	for _, p := range params {
		pItems = append(pItems, NamedItem{Name: p.name, Line: p.line})
	}
	usedInFile := FindUsedNames(file.Content, pItems)

	var unusedParams []CodeIssue
	for _, p := range params {
		if !usedInFile[p.name] {
			unusedParams = append(unusedParams, CodeIssue{
				ID:   generateUUID(),
				Line: p.line,
				Text: p.text,
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

func findJSTSImports(content string) []jsImportItem {
	var imports []jsImportItem
	seen := make(map[string]bool)

	for _, m := range reJSImportFrom.FindAllStringSubmatchIndex(content, -1) {
		clause := content[m[2]:m[3]]
		line := lineFromPos(content, m[0])
		collectJSImportNames(clause, line, seen, &imports)
	}

	for _, m := range reJSImportRequire.FindAllStringSubmatchIndex(content, -1) {
		name := content[m[2]:m[3]]
		addJSImport(name, lineFromPos(content, m[0]), seen, &imports)
	}

	for _, m := range reJSImportReqDestr.FindAllStringSubmatchIndex(content, -1) {
		clause := content[m[2]:m[3]]
		line := lineFromPos(content, m[0])
		for _, name := range extractDestructuredNames(clause) {
			addJSImport(name, line, seen, &imports)
		}
	}

	// Re-exported symbols should count as usage of source file exports in workspace mode.
	for _, m := range reJSExportNamedFrom.FindAllStringSubmatchIndex(content, -1) {
		clause := content[m[2]:m[3]]
		line := lineFromPos(content, m[0])
		collectJSImportNames("{"+clause+"}", line, seen, &imports)
	}
	for _, m := range reJSExportNsFrom.FindAllStringSubmatchIndex(content, -1) {
		name := content[m[2]:m[3]]
		addJSImport(name, lineFromPos(content, m[0]), seen, &imports)
	}

	return imports
}

func collectJSImportNames(clause string, line int, seen map[string]bool, imports *[]jsImportItem) {
	c := strings.TrimSpace(clause)
	if c == "" {
		return
	}

	// default import
	if m := reJSDefaultImportPre.FindStringSubmatch(c); len(m) > 1 && !strings.HasPrefix(c, "{") && !strings.HasPrefix(c, "*") {
		addJSImport(strings.TrimSpace(m[1]), line, seen, imports)
	}

	// named imports
	if strings.Contains(c, "{") && strings.Contains(c, "}") {
		start := strings.Index(c, "{")
		end := strings.LastIndex(c, "}")
		if start >= 0 && end > start {
			inner := c[start+1 : end]
			for _, part := range strings.Split(inner, ",") {
				part = strings.TrimSpace(strings.TrimPrefix(part, "type "))
				if part == "" {
					continue
				}
				name := part
				if strings.Contains(part, " as ") {
					t := strings.Split(part, " as ")
					name = strings.TrimSpace(t[len(t)-1])
				} else if strings.Contains(part, ":") {
					t := strings.Split(part, ":")
					name = strings.TrimSpace(t[len(t)-1])
				}
				addJSImport(name, line, seen, imports)
			}
		}
	}

	// namespace import
	if strings.Contains(c, "* as ") {
		parts := strings.Split(c, "* as ")
		if len(parts) > 1 {
			ns := strings.Fields(parts[1])
			if len(ns) > 0 {
				addJSImport(strings.TrimSpace(strings.TrimSuffix(ns[0], ",")), line, seen, imports)
			}
		}
	}
}

func addJSImport(name string, line int, seen map[string]bool, imports *[]jsImportItem) {
	name = strings.TrimSpace(name)
	if !isValidJSIdent(name) || seen[name] {
		return
	}
	seen[name] = true
	*imports = append(*imports, jsImportItem{
		name: name,
		line: line,
		text: "import " + name,
	})
}

func findJSTSDefinitions(content string) []jsVarItem {
	var defs []jsVarItem
	seen := make(map[string]bool)

	for _, m := range reJSFunctionDecl.FindAllStringSubmatchIndex(content, -1) {
		exported := m[2] != -1
		name := content[m[4]:m[5]]
		addJSDef(name, lineFromPos(content, m[0]), "function", exported, seen, &defs)
	}

	for _, m := range reJSClassDecl.FindAllStringSubmatchIndex(content, -1) {
		exported := m[2] != -1
		name := content[m[4]:m[5]]
		addJSDef(name, lineFromPos(content, m[0]), "class", exported, seen, &defs)
	}

	for _, m := range reJSArrowAssign.FindAllStringSubmatchIndex(content, -1) {
		exported := m[2] != -1
		name := content[m[4]:m[5]]
		addJSDef(name, lineFromPos(content, m[0]), "var", exported, seen, &defs)
	}

	for _, m := range reJSFuncExprAssign.FindAllStringSubmatchIndex(content, -1) {
		exported := m[2] != -1
		name := content[m[4]:m[5]]
		addJSDef(name, lineFromPos(content, m[0]), "var", exported, seen, &defs)
	}

	for _, m := range reJSVarDeclLine.FindAllStringSubmatchIndex(content, -1) {
		exported := m[2] != -1
		clause := strings.TrimSpace(content[m[4]:m[5]])
		line := lineFromPos(content, m[0])

		// Destructuring declaration: const {a, b: c} = obj / const [a, b] = arr
		if strings.HasPrefix(clause, "{") || strings.HasPrefix(clause, "[") {
			left := clause
			if strings.Contains(left, "=") {
				left = strings.TrimSpace(strings.Split(left, "=")[0])
			}
			for _, name := range extractDestructuredNames(left) {
				addJSDef(name, line, "var", exported, seen, &defs)
			}
			continue
		}

		for _, part := range strings.Split(clause, ",") {
			part = strings.TrimSpace(part)
			if part == "" {
				continue
			}
			if strings.Contains(part, "=") {
				part = strings.TrimSpace(strings.Split(part, "=")[0])
			}
			id := reJSIdentifier.FindString(part)
			addJSDef(id, line, "var", exported, seen, &defs)
		}
	}

	return defs
}

func extractDestructuredNames(pattern string) []string {
	var names []string
	seen := make(map[string]bool)

	p := strings.TrimSpace(pattern)
	if strings.HasPrefix(p, "{") && strings.HasSuffix(p, "}") {
		inner := strings.TrimSpace(p[1 : len(p)-1])
		for _, part := range strings.Split(inner, ",") {
			part = strings.TrimSpace(part)
			if part == "" {
				continue
			}
			part = strings.TrimPrefix(part, "...")
			if strings.Contains(part, "=") {
				part = strings.TrimSpace(strings.Split(part, "=")[0])
			}
			if strings.Contains(part, ":") {
				part = strings.TrimSpace(strings.Split(part, ":")[1])
			}
			if strings.HasPrefix(part, "{") || strings.HasPrefix(part, "[") {
				for _, nested := range extractDestructuredNames(part) {
					if !seen[nested] {
						seen[nested] = true
						names = append(names, nested)
					}
				}
				continue
			}
			id := reJSIdentifier.FindString(part)
			if isValidJSIdent(id) && !seen[id] {
				seen[id] = true
				names = append(names, id)
			}
		}
		return names
	}

	if strings.HasPrefix(p, "[") && strings.HasSuffix(p, "]") {
		inner := strings.TrimSpace(p[1 : len(p)-1])
		for _, part := range strings.Split(inner, ",") {
			part = strings.TrimSpace(part)
			if part == "" {
				continue
			}
			part = strings.TrimPrefix(part, "...")
			if strings.Contains(part, "=") {
				part = strings.TrimSpace(strings.Split(part, "=")[0])
			}
			if strings.HasPrefix(part, "{") || strings.HasPrefix(part, "[") {
				for _, nested := range extractDestructuredNames(part) {
					if !seen[nested] {
						seen[nested] = true
						names = append(names, nested)
					}
				}
				continue
			}
			id := reJSIdentifier.FindString(part)
			if isValidJSIdent(id) && !seen[id] {
				seen[id] = true
				names = append(names, id)
			}
		}
	}

	return names
}

func addJSDef(name string, line int, kind string, exported bool, seen map[string]bool, defs *[]jsVarItem) {
	name = strings.TrimSpace(name)
	if !isValidJSIdent(name) || seen[name] || name == "_" {
		return
	}
	seen[name] = true
	*defs = append(*defs, jsVarItem{
		name:     name,
		line:     line,
		kind:     kind,
		exported: exported,
	})
}

func findJSTSParameters(content string) []jsParamItem {
	var params []jsParamItem
	seen := make(map[string]bool)

	for _, m := range reJSFunctionDecl.FindAllStringSubmatchIndex(content, -1) {
		line := lineFromPos(content, m[0])
		list := content[m[6]:m[7]]
		collectJSParams(list, line, seen, &params)
	}

	for _, m := range reJSArrowAssign.FindAllStringSubmatchIndex(content, -1) {
		line := lineFromPos(content, m[0])
		var list string
		if m[6] != -1 {
			list = content[m[6]:m[7]]
		} else if m[8] != -1 {
			list = content[m[8]:m[9]]
		}
		collectJSParams(list, line, seen, &params)
	}

	for _, m := range reJSFuncExprAssign.FindAllStringSubmatchIndex(content, -1) {
		line := lineFromPos(content, m[0])
		list := content[m[6]:m[7]]
		collectJSParams(list, line, seen, &params)
	}

	return params
}

func collectJSParams(list string, line int, seen map[string]bool, params *[]jsParamItem) {
	for _, p := range strings.Split(list, ",") {
		p = strings.TrimSpace(p)
		if p == "" {
			continue
		}
		if strings.HasPrefix(p, "{") || strings.HasPrefix(p, "[") {
			continue
		}
		p = strings.TrimPrefix(p, "...")
		if strings.Contains(p, "=") {
			p = strings.TrimSpace(strings.Split(p, "=")[0])
		}
		if strings.Contains(p, ":") {
			p = strings.TrimSpace(strings.Split(p, ":")[0])
		}
		name := reJSIdentifier.FindString(p)
		key := name + "@" + intToStr(line)
		if !isValidJSIdent(name) || name == "_" || seen[key] {
			continue
		}
		seen[key] = true
		*params = append(*params, jsParamItem{
			name: name,
			line: line,
			text: "parameter " + name,
		})
	}
}

func isFrameworkSpecialExport(name, filename string) bool {
	if name == "" {
		return false
	}

	lowerName := strings.ToLower(name)
	lowerFile := strings.ToLower(strings.ReplaceAll(filename, "\\", "/"))

	if strings.Contains(lowerFile, "/route.") {
		switch name {
		case "GET", "POST", "PUT", "PATCH", "DELETE", "HEAD", "OPTIONS":
			return true
		}
	}

	if strings.Contains(lowerFile, "middleware.") {
		switch lowerName {
		case "config", "matcher", "runtime", "preferredregion":
			return true
		}
	}

	frameworkNames := map[string]bool{
		"generatestaticparams": true,
		"generatemetadata":     true,
		"metadata":             true,
		"viewport":             true,
		"revalidate":           true,
		"dynamic":              true,
		"fetchcache":           true,
		"runtime":              true,
		"preferredregion":      true,
		"maxduration":          true,
		"loader":               true,
		"action":               true,
		"meta":                 true,
		"links":                true,
		"headers":              true,
		"handle":               true,
		"load":                 true,
		"actions":              true,
		"prerender":            true,
		"ssr":                  true,
		"csr":                  true,
		"trailingslash":        true,
		"entries":              true,
	}
	if frameworkNames[lowerName] {
		return true
	}

	return false
}

func findUsedJSImportNames(content string, imports []jsImportItem) map[string]bool {
	used := make(map[string]bool)
	lines := strings.Split(content, "\n")

	for _, imp := range imports {
		if imp.name == "" {
			continue
		}

		openTag := "<" + imp.name
		closeTag := "</" + imp.name

		for i, line := range lines {
			if i+1 == imp.line {
				continue
			}

			if strings.Contains(line, openTag) || strings.Contains(line, closeTag) || containsWord(line, imp.name) {
				used[imp.name] = true
				break
			}
		}
	}

	return used
}

func isValidJSIdent(name string) bool {
	return name != "" && reJSIdentifier.MatchString(name)
}

func lineFromPos(content string, pos int) int {
	if pos <= 0 {
		return 1
	}
	return strings.Count(content[:pos], "\n") + 1
}

func intToStr(v int) string {
	return strconv.Itoa(v)
}
