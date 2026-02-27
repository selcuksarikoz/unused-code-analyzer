package main

import (
	"strings"
	"unicode"
)

// Astro file parsing (frontmatter between ---)
func extractAstroScript(content string) string {
	lines := strings.Split(content, "\n")
	var inFrontmatter bool
	var frontmatterLines []string

	for i, line := range lines {
		trimmed := strings.TrimSpace(line)

		// Start of frontmatter
		if trimmed == "---" && !inFrontmatter {
			// Check if this is the start (should be at beginning or have content before that's not frontmatter)
			if i == 0 || len(frontmatterLines) == 0 {
				inFrontmatter = true
				continue
			}
		}

		// End of frontmatter
		if trimmed == "---" && inFrontmatter {
			inFrontmatter = false
			break
		}

		if inFrontmatter {
			frontmatterLines = append(frontmatterLines, line)
		}
	}

	return strings.Join(frontmatterLines, "\n")
}

// Svelte/Vue file parsing (script tags)
func extractScriptContent(content string) string {
	var result strings.Builder
	var inScript bool
	var inScriptTag bool

	for i := 0; i < len(content); i++ {
		// Look for <script
		if !inScript && i+7 < len(content) {
			chunk := strings.ToLower(content[i : i+7])
			if strings.HasPrefix(chunk, "<script") {
				// Check if it's a script tag (not <script something else)
				nextChar := content[i+7]
				if nextChar == '>' || nextChar == ' ' || nextChar == '\t' || nextChar == '\n' {
					inScriptTag = true
					i += 6
					continue
				}
			}
		}

		// Find end of script tag
		if inScriptTag && !inScript {
			if content[i] == '>' {
				inScript = true
				inScriptTag = false
				continue
			}
		}

		// Look for </script>
		if inScript && i+8 < len(content) {
			chunk := strings.ToLower(content[i : i+9])
			if chunk == "</script>" {
				inScript = false
				i += 8
				continue
			}
		}

		if inScript {
			result.WriteByte(content[i])
		}
	}

	return result.String()
}

// Tokenizer-based JS/TS analysis for extracted script content
type tokenType int

const (
	tokEOF tokenType = iota
	tokIdentifier
	tokImport
	tokFrom
	tokString
	tokLeftBrace
	tokRightBrace
	tokStar
	tokComma
	tokAs
	tokConst
	tokLet
	tokVar
	tokFunction
	tokExport
	tokDefault
	tokType
	tokInterface
	tokOther
)

type token struct {
	typ  tokenType
	val  string
	line int
}

// Simple tokenizer for JS/TS script content
func tokenizeJS(content string) []token {
	var tokens []token
	lines := strings.Split(content, "\n")

	for lineNum, line := range lines {
		i := 0
		for i < len(line) {
			// Skip whitespace
			for i < len(line) && unicode.IsSpace(rune(line[i])) {
				i++
			}
			if i >= len(line) {
				break
			}

			// Skip comments
			if i+1 < len(line) && line[i] == '/' && line[i+1] == '/' {
				break // Skip rest of line
			}
			if i+1 < len(line) && line[i] == '/' && line[i+1] == '*' {
				// Skip multi-line comment
				for i < len(line) && !(line[i-1] == '*' && line[i] == '/') {
					i++
				}
				i++
				continue
			}

			// String literals
			if line[i] == '"' || line[i] == '\'' || line[i] == '`' {
				quote := line[i]
				start := i
				i++
				for i < len(line) && line[i] != quote {
					if line[i] == '\\' && i+1 < len(line) {
						i += 2
					} else {
						i++
					}
				}
				if i < len(line) {
					i++
				}
				tokens = append(tokens, token{typ: tokString, val: line[start:i], line: lineNum + 1})
				continue
			}

			// Identifiers and keywords
			if isJSIdentStart(rune(line[i])) {
				start := i
				for i < len(line) && isJSIdentPart(rune(line[i])) {
					i++
				}
				ident := line[start:i]
				tok := token{typ: tokIdentifier, val: ident, line: lineNum + 1}

				switch ident {
				case "import":
					tok.typ = tokImport
				case "from":
					tok.typ = tokFrom
				case "as":
					tok.typ = tokAs
				case "const":
					tok.typ = tokConst
				case "let":
					tok.typ = tokLet
				case "var":
					tok.typ = tokVar
				case "function":
					tok.typ = tokFunction
				case "export":
					tok.typ = tokExport
				case "default":
					tok.typ = tokDefault
				case "type":
					tok.typ = tokType
				case "interface":
					tok.typ = tokInterface
				}

				tokens = append(tokens, tok)
				continue
			}

			// Single character tokens
			switch line[i] {
			case '{':
				tokens = append(tokens, token{typ: tokLeftBrace, val: "{", line: lineNum + 1})
			case '}':
				tokens = append(tokens, token{typ: tokRightBrace, val: "}", line: lineNum + 1})
			case '*':
				tokens = append(tokens, token{typ: tokStar, val: "*", line: lineNum + 1})
			case ',':
				tokens = append(tokens, token{typ: tokComma, val: ",", line: lineNum + 1})
			case ';':
				// Statement separator
			default:
				tokens = append(tokens, token{typ: tokOther, val: string(line[i]), line: lineNum + 1})
			}
			i++
		}
	}

	tokens = append(tokens, token{typ: tokEOF, line: len(lines)})
	return tokens
}

func isJSIdentStart(r rune) bool {
	return unicode.IsLetter(r) || r == '_' || r == '$'
}

func isJSIdentPart(r rune) bool {
	return unicode.IsLetter(r) || unicode.IsDigit(r) || r == '_' || r == '$'
}

// Parse imports from tokens
func parseJSImports(tokens []token) []Import {
	var imports []Import
	i := 0

	for i < len(tokens) {
		if tokens[i].typ != tokImport {
			i++
			continue
		}

		importLine := tokens[i].line
		i++

		// Skip whitespace and comments in parsing
		for i < len(tokens) && (tokens[i].typ == tokOther || tokens[i].typ == tokEOF) {
			i++
		}

		if i >= len(tokens) {
			break
		}

		// import * as name from "..."
		if tokens[i].typ == tokStar {
			i++
			// Skip "as"
			if i < len(tokens) && tokens[i].typ == tokAs {
				i++
			}
			// Get namespace name
			if i < len(tokens) && tokens[i].typ == tokIdentifier {
				namespace := tokens[i].val
				imports = append(imports, Import{
					Name: namespace,
					Line: importLine,
				})
				i++
			}
			// Skip to end of import statement
			for i < len(tokens) && tokens[i].typ != tokString && tokens[i].typ != tokEOF {
				i++
			}
			if i < len(tokens) && tokens[i].typ == tokString {
				i++
			}
			continue
		}

		// import name from "..."
		if tokens[i].typ == tokIdentifier {
			defaultImport := tokens[i].val
			imports = append(imports, Import{
				Name: defaultImport,
				Line: importLine,
			})
			i++

			// Check for named imports: import name, { ... } from "..."
			if i < len(tokens) && tokens[i].typ == tokComma {
				i++
				// Continue to parse { ... }
			}
		}

		// import { a, b, c } from "..."
		if i < len(tokens) && tokens[i].typ == tokLeftBrace {
			i++
			var names []string
			for i < len(tokens) && tokens[i].typ != tokRightBrace && tokens[i].typ != tokEOF {
				if tokens[i].typ == tokIdentifier {
					names = append(names, tokens[i].val)
					i++
				} else if tokens[i].typ == tokComma {
					i++
				} else {
					i++
				}
			}
			if i < len(tokens) && tokens[i].typ == tokRightBrace {
				i++
			}
			for _, name := range names {
				imports = append(imports, Import{
					Name: name,
					Line: importLine,
				})
			}
		}

		// Skip "from" and source
		for i < len(tokens) && tokens[i].typ != tokString && tokens[i].typ != tokEOF {
			i++
		}
		if i < len(tokens) && tokens[i].typ == tokString {
			i++
		}
	}

	return imports
}

// Parse definitions (variables, functions, types, interfaces) from tokens
func parseJSDefinitions(tokens []token, filename string) []Definition {
	var defs []Definition
	i := 0

	for i < len(tokens) {
		switch tokens[i].typ {
		case tokConst, tokLet, tokVar:
			line := tokens[i].line
			i++
			// Handle destructuring: const { a, b } = ...
			if i < len(tokens) && tokens[i].typ == tokLeftBrace {
				i++
				for i < len(tokens) && tokens[i].typ != tokRightBrace && tokens[i].typ != tokEOF {
					if tokens[i].typ == tokIdentifier {
						defs = append(defs, Definition{
							Name: tokens[i].val,
							File: filename,
							Line: line,
							Type: "variable",
						})
					}
					i++
				}
				if i < len(tokens) && tokens[i].typ == tokRightBrace {
					i++
				}
			} else if i < len(tokens) && tokens[i].typ == tokIdentifier {
				// Simple declaration: const name = ...
				defs = append(defs, Definition{
					Name: tokens[i].val,
					File: filename,
					Line: line,
					Type: "variable",
				})
				i++
			}

		case tokFunction:
			line := tokens[i].line
			i++
			if i < len(tokens) && tokens[i].typ == tokIdentifier {
				defs = append(defs, Definition{
					Name: tokens[i].val,
					File: filename,
					Line: line,
					Type: "function",
				})
				i++
			}

		case tokType:
			line := tokens[i].line
			i++
			if i < len(tokens) && tokens[i].typ == tokIdentifier {
				defs = append(defs, Definition{
					Name: tokens[i].val,
					File: filename,
					Line: line,
					Type: "type",
				})
				i++
			}

		case tokInterface:
			line := tokens[i].line
			i++
			if i < len(tokens) && tokens[i].typ == tokIdentifier {
				defs = append(defs, Definition{
					Name: tokens[i].val,
					File: filename,
					Line: line,
					Type: "interface",
				})
				i++
			}

		default:
			i++
		}
	}

	return defs
}

// Find all used identifiers in script content (excluding imports and definitions)
func findJSUsedNames(scriptContent string, defs []Definition, imports []Import) map[string]bool {
	used := make(map[string]bool)

	// Create sets for quick lookup
	defNames := make(map[string]bool)
	for _, d := range defs {
		defNames[d.Name] = true
	}

	importNames := make(map[string]bool)
	for _, imp := range imports {
		importNames[imp.Name] = true
	}

	tokens := tokenizeJS(scriptContent)

	for _, tok := range tokens {
		if tok.typ == tokIdentifier {
			name := tok.val
			// Skip if it's a definition or import
			if !defNames[name] && !importNames[name] {
				used[name] = true
			}
		}
	}

	return used
}

// Check if Astro component is used in template
func isAstroComponentUsedInTemplate(content string, componentName string) bool {
	// Split at end of frontmatter
	parts := strings.Split(content, "---")
	if len(parts) < 3 {
		return false
	}

	// Template is after second ---
	template := parts[2]

	// Check for component usage: <ComponentName or ComponentName in template
	lowerComponent := strings.ToLower(componentName)
	lowerTemplate := strings.ToLower(template)

	// Direct usage: <ComponentName
	if strings.Contains(lowerTemplate, "<"+lowerComponent) {
		return true
	}

	// In expressions: {ComponentName}
	if strings.Contains(lowerTemplate, "{"+componentName) {
		return true
	}

	return false
}

// Check if Svelte variable is used in template
func isSvelteVariableUsedInTemplate(content string, varName string) bool {
	// Extract template (outside <script> tags)
	scriptContent := extractScriptContent(content)
	template := strings.Replace(content, scriptContent, "", 1)

	// Remove script tags from template
	template = removeScriptTags(template)

	// Check for variable usage in template
	return containsWord(template, varName)
}

func removeScriptTags(content string) string {
	var result strings.Builder
	var inScript bool
	var inScriptTag bool
	i := 0

	for i < len(content) {
		if !inScript && i+7 < len(content) {
			chunk := strings.ToLower(content[i : i+7])
			if strings.HasPrefix(chunk, "<script") {
				nextChar := content[i+7]
				if nextChar == '>' || nextChar == ' ' || nextChar == '\t' || nextChar == '\n' {
					inScriptTag = true
					i += 6
					continue
				}
			}
		}

		if inScriptTag && !inScript {
			if content[i] == '>' {
				inScript = true
				inScriptTag = false
				i++
				continue
			}
		}

		if inScript && i+8 < len(content) {
			chunk := strings.ToLower(content[i : i+9])
			if chunk == "</script>" {
				inScript = false
				i += 9
				continue
			}
		}

		if !inScript && !inScriptTag {
			result.WriteByte(content[i])
		}
		i++
	}

	return result.String()
}

// Analyze Astro file
func analyzeAstro(content string, filename string) AnalysisResult {
	scriptContent := extractAstroScript(content)
	if scriptContent == "" {
		return AnalysisResult{Imports: []CodeIssue{}, Variables: []CodeIssue{}, Parameters: []CodeIssue{}}
	}

	tokens := tokenizeJS(scriptContent)
	imports := parseJSImports(tokens)
	defs := parseJSDefinitions(tokens, filename)
	usedNames := findJSUsedNames(scriptContent, defs, imports)

	var unusedImports []CodeIssue
	var unusedVars []CodeIssue

	// Check imports - consider Astro template usage
	for _, imp := range imports {
		isUsed := usedNames[imp.Name]

		// For Astro components, also check template usage
		if !isUsed && isAstroComponentUsedInTemplate(content, imp.Name) {
			isUsed = true
		}

		if !isUsed {
			unusedImports = append(unusedImports, CodeIssue{
				ID:   generateUUID(),
				Line: imp.Line,
				Text: "import " + imp.Name,
				File: filename,
			})
		}
	}

	// Check variable/definitions
	for _, def := range defs {
		if !usedNames[def.Name] {
			unusedVars = append(unusedVars, CodeIssue{
				ID:   generateUUID(),
				Line: def.Line,
				Text: def.Type + " " + def.Name,
				File: filename,
			})
		}
	}

	// Find unused parameters
	params := findJSUnusedParameters(scriptContent, filename)

	return AnalysisResult{
		Imports:    unusedImports,
		Variables:  unusedVars,
		Parameters: params,
	}
}

// Analyze Svelte file
func analyzeSvelte(content string, filename string) AnalysisResult {
	scriptContent := extractScriptContent(content)
	if scriptContent == "" {
		return AnalysisResult{Imports: []CodeIssue{}, Variables: []CodeIssue{}, Parameters: []CodeIssue{}}
	}

	tokens := tokenizeJS(scriptContent)
	imports := parseJSImports(tokens)
	defs := parseJSDefinitions(tokens, filename)
	usedNames := findJSUsedNames(scriptContent, defs, imports)

	var unusedImports []CodeIssue
	var unusedVars []CodeIssue

	// Check imports
	for _, imp := range imports {
		if !usedNames[imp.Name] {
			unusedImports = append(unusedImports, CodeIssue{
				ID:   generateUUID(),
				Line: imp.Line,
				Text: "import " + imp.Name,
				File: filename,
			})
		}
	}

	// Check variables - also consider Svelte template usage
	for _, def := range defs {
		isUsed := usedNames[def.Name]

		// For Svelte, also check template usage
		if !isUsed && isSvelteVariableUsedInTemplate(content, def.Name) {
			isUsed = true
		}

		if !isUsed {
			unusedVars = append(unusedVars, CodeIssue{
				ID:   generateUUID(),
				Line: def.Line,
				Text: def.Type + " " + def.Name,
				File: filename,
			})
		}
	}

	// Find unused parameters
	params := findJSUnusedParameters(scriptContent, filename)

	return AnalysisResult{
		Imports:    unusedImports,
		Variables:  unusedVars,
		Parameters: params,
	}
}

// Find unused parameters in JS/TS code
func findJSUnusedParameters(content string, filename string) []CodeIssue {
	lines := strings.Split(content, "\n")
	var issues []CodeIssue

	for lineNum, line := range lines {
		// Look for function declarations
		trimmed := strings.TrimSpace(line)

		// Match: function name(...)
		if strings.HasPrefix(trimmed, "function ") {
			// Extract parameters from function declaration
			start := strings.Index(trimmed, "(")
			end := strings.Index(trimmed, ")")
			if start != -1 && end != -1 && end > start {
				paramsStr := trimmed[start+1 : end]
				params := parseJSParams(paramsStr)

				// Check each parameter usage in function body
				for _, param := range params {
					if !isParamUsedInFunction(lines, lineNum, param) {
						issues = append(issues, CodeIssue{
							ID:   generateUUID(),
							Line: lineNum + 1,
							Text: "parameter " + param,
							File: filename,
						})
					}
				}
			}
		}

		// Match: const fn = (...) =>
		// Match: const fn = function(...)
		if strings.Contains(trimmed, "=") {
			if strings.Contains(trimmed, "=>") || strings.Contains(trimmed, "function(") {
				start := strings.Index(trimmed, "(")
				end := strings.Index(trimmed, ")")
				if start != -1 && end != -1 && end > start {
					paramsStr := trimmed[start+1 : end]
					params := parseJSParams(paramsStr)

					for _, param := range params {
						if !isParamUsedInFunction(lines, lineNum, param) {
							issues = append(issues, CodeIssue{
								ID:   generateUUID(),
								Line: lineNum + 1,
								Text: "parameter " + param,
								File: filename,
							})
						}
					}
				}
			}
		}
	}

	return issues
}

// Parse JS parameters from string like "a, b, c" or "a: Type, b: Type"
func parseJSParams(paramsStr string) []string {
	var params []string
	// Remove type annotations
	paramsStr = removeJSTypes(paramsStr)

	parts := strings.Split(paramsStr, ",")
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}

		// Handle destructuring by skipping
		if strings.HasPrefix(part, "{") || strings.HasPrefix(part, "[") {
			continue
		}

		// Get parameter name (before any type annotation)
		if idx := strings.Index(part, ":"); idx != -1 {
			part = strings.TrimSpace(part[:idx])
		}

		if part != "" {
			params = append(params, part)
		}
	}

	return params
}

// Remove TypeScript type annotations from a string
func removeJSTypes(s string) string {
	var result strings.Builder
	depth := 0

	for i := 0; i < len(s); i++ {
		if s[i] == '<' {
			depth++
			continue
		}
		if s[i] == '>' {
			depth--
			continue
		}
		if depth == 0 {
			result.WriteByte(s[i])
		}
	}

	return result.String()
}

// Check if a parameter is used in function body (simple version)
func isParamUsedInFunction(lines []string, funcStartLine int, param string) bool {
	// Find function body
	braceDepth := 0
	inFunction := false

	for i := funcStartLine; i < len(lines); i++ {
		line := lines[i]

		if !inFunction {
			// Look for opening brace
			if strings.Contains(line, "{") {
				inFunction = true
				braceDepth = 1
			}
			continue
		}

		// Count braces
		for _, ch := range line {
			if ch == '{' {
				braceDepth++
			} else if ch == '}' {
				braceDepth--
				if braceDepth == 0 {
					return false
				}
			}
		}

		// Check for parameter usage (not on the definition line)
		if i > funcStartLine && containsWord(line, param) {
			return true
		}
	}

	return false
}
