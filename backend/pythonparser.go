package main

import (
	"strings"
	"unicode"
	"unicode/utf8"
)

type PyTokenType int

const (
	PyTokenImport PyTokenType = iota
	PyTokenFrom
	PyTokenDef
	PyTokenClass
	PyTokenIdentifier
	PyTokenString
	PyTokenLParen
	PyTokenRParen
	PyTokenComma
	PyTokenDot
	PyTokenNewline
	PyTokenColon
	PyTokenEquals
	PyTokenAs
	PyTokenEOF
	PyTokenUnknown
)

type PyToken struct {
	Type  PyTokenType
	Value string
	Line  int
}

type PyTokenizer struct {
	content string
	pos     int
	line    int
	tokens  []PyToken
}

func NewPyTokenizer(content string) *PyTokenizer {
	return &PyTokenizer{
		content: content,
		pos:     0,
		line:    1,
		tokens:  make([]PyToken, 0),
	}
}

func (t *PyTokenizer) peek() rune {
	if t.pos >= len(t.content) {
		return 0
	}
	r, _ := utf8.DecodeRuneInString(t.content[t.pos:])
	return r
}

func (t *PyTokenizer) next() rune {
	if t.pos >= len(t.content) {
		return 0
	}
	r, size := utf8.DecodeRuneInString(t.content[t.pos:])
	t.pos += size
	if r == '\n' {
		t.line++
	}
	return r
}

func (t *PyTokenizer) Tokenize() []PyToken {
	for t.pos < len(t.content) {
		ch := t.peek()

		if ch == 0 {
			break
		}

		if ch == '\n' {
			t.next()
			t.tokens = append(t.tokens, PyToken{Type: PyTokenNewline, Value: "\n", Line: t.line})
			continue
		}

		if unicode.IsSpace(ch) {
			t.next()
			continue
		}

		if ch == '#' {
			for t.peek() != '\n' && t.peek() != 0 {
				t.next()
			}
			continue
		}

		if ch == '"' || ch == '\'' {
			t.tokens = append(t.tokens, t.readString(ch))
			continue
		}

		if unicode.IsLetter(ch) || ch == '_' {
			t.tokens = append(t.tokens, t.readIdentifier())
			continue
		}

		if unicode.IsDigit(ch) {
			t.next()
			for unicode.IsDigit(t.peek()) {
				t.next()
			}
			continue
		}

		switch ch {
		case '(':
			t.next()
			t.tokens = append(t.tokens, PyToken{Type: PyTokenLParen, Value: "(", Line: t.line})
		case ')':
			t.next()
			t.tokens = append(t.tokens, PyToken{Type: PyTokenRParen, Value: ")", Line: t.line})
		case ',':
			t.next()
			t.tokens = append(t.tokens, PyToken{Type: PyTokenComma, Value: ",", Line: t.line})
		case '.':
			t.next()
			t.tokens = append(t.tokens, PyToken{Type: PyTokenDot, Value: ".", Line: t.line})
		case ':':
			t.next()
			t.tokens = append(t.tokens, PyToken{Type: PyTokenColon, Value: ":", Line: t.line})
		case '=':
			t.next()
			t.tokens = append(t.tokens, PyToken{Type: PyTokenEquals, Value: "=", Line: t.line})
		default:
			t.next()
		}
	}

	t.tokens = append(t.tokens, PyToken{Type: PyTokenEOF, Value: "", Line: t.line})
	return t.tokens
}

func (t *PyTokenizer) readString(quote rune) PyToken {
	start := t.pos
	line := t.line
	t.next()
	for t.peek() != quote && t.peek() != 0 {
		if t.peek() == '\\' {
			t.next()
			if t.peek() != 0 {
				t.next()
			}
		} else {
			t.next()
		}
	}
	if t.peek() == quote {
		t.next()
	}
	return PyToken{Type: PyTokenString, Value: t.content[start:t.pos], Line: line}
}

func (t *PyTokenizer) readIdentifier() PyToken {
	start := t.pos
	line := t.line
	for unicode.IsLetter(t.peek()) || unicode.IsDigit(t.peek()) || t.peek() == '_' {
		t.next()
	}
	value := t.content[start:t.pos]

	switch value {
	case "import":
		return PyToken{Type: PyTokenImport, Value: value, Line: line}
	case "from":
		return PyToken{Type: PyTokenFrom, Value: value, Line: line}
	case "def":
		return PyToken{Type: PyTokenDef, Value: value, Line: line}
	case "class":
		return PyToken{Type: PyTokenClass, Value: value, Line: line}
	default:
		return PyToken{Type: PyTokenIdentifier, Value: value, Line: line}
	}
}

type PyImportItem struct {
	module string
	names  []string
	alias  map[string]string
	line   int
	text   string
}

func FindPythonImports(content string) []PyImportItem {
	t := NewPyTokenizer(content)
	tokens := t.Tokenize()

	var imports []PyImportItem

	var i int
	for i < len(tokens) {
		if tokens[i].Type == PyTokenImport && i+1 < len(tokens) && tokens[i+1].Type == PyTokenIdentifier {
			line := tokens[i].Line
			var module string
			alias := make(map[string]string)

			i++
			module = tokens[i].Value

			for i+1 < len(tokens) && tokens[i+1].Type == PyTokenDot {
				i += 2
				if i < len(tokens) && tokens[i].Type == PyTokenIdentifier {
					module += "." + tokens[i].Value
				}
			}

			if i+1 < len(tokens) && tokens[i+1].Type == PyTokenIdentifier && (i+2 >= len(tokens) || tokens[i+2].Type != PyTokenEquals) {
				i++
				var names []string
				for i < len(tokens) && tokens[i].Type == PyTokenIdentifier {
					names = append(names, tokens[i].Value)
					i++
					if i < len(tokens) && tokens[i].Type == PyTokenAs {
						i++
						if i < len(tokens) && tokens[i].Type == PyTokenIdentifier {
							alias[names[len(names)-1]] = tokens[i].Value
							i++
						}
					}
					if i < len(tokens) && tokens[i].Type == PyTokenComma {
						i++
					}
				}

				text := "import " + module
				if len(names) > 0 {
					text += " import " + strings.Join(names, ", ")
				}

				imports = append(imports, PyImportItem{
					module: module,
					names:  names,
					alias:  alias,
					line:   line,
					text:   text,
				})
			}
		}

		if tokens[i].Type == PyTokenFrom && i+1 < len(tokens) && tokens[i+1].Type == PyTokenIdentifier {
			line := tokens[i].Line
			var module string

			i += 2
			module = tokens[i-1].Value

			for i+1 < len(tokens) && tokens[i].Type == PyTokenDot {
				i += 2
				if i < len(tokens) && tokens[i].Type == PyTokenIdentifier {
					module += "." + tokens[i].Value
				}
			}

			if i+1 < len(tokens) && tokens[i].Type == PyTokenImport {
				i++
				var names []string
				alias := make(map[string]string)

				for i < len(tokens) && tokens[i].Type == PyTokenIdentifier {
					names = append(names, tokens[i].Value)
					i++
					if i < len(tokens) && tokens[i].Type == PyTokenAs {
						i++
						if i < len(tokens) && tokens[i].Type == PyTokenIdentifier {
							alias[names[len(names)-1]] = tokens[i].Value
							i++
						}
					}
					if i < len(tokens) && tokens[i].Type == PyTokenComma {
						i++
					}
				}

				text := "from " + module + " import " + strings.Join(names, ", ")

				imports = append(imports, PyImportItem{
					module: module,
					names:  names,
					alias:  alias,
					line:   line,
					text:   text,
				})
			}
		}

		i++
	}

	return imports
}

type PyDefinition struct {
	name    string
	defType string
	line    int
}

func FindPythonDefinitions(content string) []PyDefinition {
	t := NewPyTokenizer(content)
	tokens := t.Tokenize()

	var defs []PyDefinition

	for i := 0; i < len(tokens); i++ {
		if tokens[i].Type == PyTokenDef && i+1 < len(tokens) && tokens[i+1].Type == PyTokenIdentifier {
			defs = append(defs, PyDefinition{
				name:    tokens[i+1].Value,
				defType: "function",
				line:    tokens[i].Line,
			})
		}
		if tokens[i].Type == PyTokenClass && i+1 < len(tokens) && tokens[i+1].Type == PyTokenIdentifier {
			defs = append(defs, PyDefinition{
				name:    tokens[i+1].Value,
				defType: "class",
				line:    tokens[i].Line,
			})
		}
	}

	return defs
}

func FindPythonParameters(content, filename string) []CodeIssue {
	t := NewPyTokenizer(content)
	tokens := t.Tokenize()

	var params []CodeIssue
	for i := 0; i < len(tokens); i++ {
		if tokens[i].Type == PyTokenDef && i+1 < len(tokens) && tokens[i+1].Type == PyTokenIdentifier {
			parenCount := 0
			paramStart := -1
			for j := i + 2; j < len(tokens); j++ {
				if tokens[j].Type == PyTokenLParen {
					if parenCount == 0 {
						paramStart = j + 1
					}
					parenCount++
				}
				if tokens[j].Type == PyTokenRParen {
					parenCount--
					if parenCount == 0 {
						break
					}
				}
			}

			if paramStart > 0 && paramStart < len(tokens) {
				for j := paramStart; j < len(tokens) && tokens[j].Type != PyTokenRParen; j++ {
					if tokens[j].Type == PyTokenIdentifier && j+1 < len(tokens) && (tokens[j+1].Type == PyTokenComma || tokens[j+1].Type == PyTokenRParen) {
						paramName := tokens[j].Value
						if paramName != "" && paramName != "self" && paramName != "cls" {
							params = append(params, CodeIssue{
								ID:   generateUUID(),
								Line: tokens[j].Line,
								Text: paramName + " (parameter)",
								File: filename,
							})
						}
					}
				}
			}
		}
	}

	return params
}

func FindUsedPythonNames(content string) map[string]int {
	t := NewPyTokenizer(content)
	tokens := t.Tokenize()

	counts := make(map[string]int)
	reserved := map[string]bool{
		"import": true, "from": true, "def": true, "class": true,
		"return": true, "if": true, "elif": true, "else": true,
		"for": true, "while": true, "try": true, "except": true,
		"finally": true, "with": true, "as": true, "pass": true,
		"break": true, "continue": true, "True": true, "False": true,
		"None": true, "and": true, "or": true, "not": true,
		"in": true, "is": true, "lambda": true, "yield": true,
		"global": true, "nonlocal": true, "assert": true, "raise": true,
	}

	inImport := false

	for _, tok := range tokens {
		if tok.Type == PyTokenImport || tok.Type == PyTokenFrom {
			inImport = true
			continue
		}
		if inImport && tok.Type == PyTokenNewline {
			inImport = false
			continue
		}

		if tok.Type == PyTokenIdentifier && !reserved[tok.Value] {
			if inImport {
				continue
			}
			counts[tok.Value]++
		}
	}

	return counts
}

func analyzePython(content, filename string) AnalysisResult {
	imports := FindPythonImports(content)
	params := FindPythonParameters(content, filename)
	counts := FindUsedPythonNames(content)

	var unusedImports []CodeIssue
	for _, imp := range imports {
		allUnused := true
		for _, name := range imp.names {
			if counts[name] > 1 {
				allUnused = false
				break
			}
			if alias, ok := imp.alias[name]; ok && counts[alias] > 1 {
				allUnused = false
				break
			}
		}
		if allUnused && len(imp.names) > 0 {
			unusedImports = append(unusedImports, CodeIssue{
				ID:   generateUUID(),
				Line: imp.line,
				Text: imp.text,
				File: filename,
			})
		}
	}

	var unusedParams []CodeIssue = []CodeIssue{}
	for _, p := range params {
		paramName := strings.TrimSpace(strings.TrimSuffix(p.Text, " (parameter)"))
		isLocallyUsed := counts[paramName] > 1
		if !isLocallyUsed {
			unusedParams = append(unusedParams, CodeIssue{
				ID:   generateUUID(),
				Line: p.Line,
				Text: "parameter " + paramName,
				File: filename,
			})
		}
	}

	return AnalysisResult{
		Imports:    unusedImports,
		Variables:  []CodeIssue{},
		Parameters: unusedParams,
	}
}

func analyzePythonForWorkspace(content, filename string) ([]Definition, []Import, []CodeIssue, []CodeIssue) {
	imports := FindPythonImports(content)
	defs := FindPythonDefinitions(content)
	counts := FindUsedPythonNames(content)

	var outImports []Import
	var unusedImports []CodeIssue
	var outDefs []Definition

	for _, imp := range imports {
		outImports = append(outImports, Import{
			Name: strings.Join(imp.names, ", "),
			File: filename,
			Line: imp.line,
		})

		allUnused := true
		for _, name := range imp.names {
			if counts[name] > 1 {
				allUnused = false
				break
			}
		}
		if allUnused && len(imp.names) > 0 {
			unusedImports = append(unusedImports, CodeIssue{
				ID:   generateUUID(),
				Line: imp.line,
				Text: imp.text,
				File: filename,
			})
		}
	}

	for _, d := range defs {
		outDefs = append(outDefs, Definition{
			Name: d.name,
			Type: d.defType,
			Line: d.line,
			File: filename,
		})
	}

	return outDefs, outImports, unusedImports, []CodeIssue{}
}

func buildResultPython(file AnalyzeFile, defs []Definition, imports []Import, usedNames map[string]bool, allFiles []AnalyzeFile) AnalysisResult {
	localImports := FindPythonImports(file.Content)
	localDefs := FindPythonDefinitions(file.Content)
	localParams := FindPythonParameters(file.Content, file.Filename)
	counts := FindUsedPythonNames(file.Content)

	var unusedImports []CodeIssue
	for _, imp := range localImports {
		allUnused := true
		for _, name := range imp.names {
			isCrossFileUsed := usedNames[name+"@"+file.Filename]
			isLocallyUsed := counts[name] > 0
			if isCrossFileUsed || isLocallyUsed {
				allUnused = false
				break
			}
			if alias, ok := imp.alias[name]; ok {
				isAliasCrossFileUsed := usedNames[alias+"@"+file.Filename]
				isAliasLocallyUsed := counts[alias] > 0
				if isAliasCrossFileUsed || isAliasLocallyUsed {
					allUnused = false
					break
				}
			}
		}
		if allUnused && len(imp.names) > 0 {
			unusedImports = append(unusedImports, CodeIssue{
				ID:   generateUUID(),
				Line: imp.line,
				Text: imp.text,
				File: file.Filename,
			})
		}
	}

	var unusedVars []CodeIssue
	for _, d := range localDefs {
		isCrossFileUsed := usedNames[d.name+"@"+file.Filename]
		isLocallyUsed := counts[d.name] > 1
		if !isCrossFileUsed && !isLocallyUsed {
			unusedVars = append(unusedVars, CodeIssue{
				ID:   generateUUID(),
				Line: d.line,
				Text: d.defType + " " + d.name,
				File: file.Filename,
			})
		}
	}

	var unusedParams []CodeIssue = []CodeIssue{}
	for _, p := range localParams {
		paramName := strings.TrimSpace(strings.TrimSuffix(p.Text, " (parameter)"))
		isLocallyUsed := counts[paramName] > 1
		if !isLocallyUsed {
			unusedParams = append(unusedParams, CodeIssue{
				ID:   generateUUID(),
				Line: p.Line,
				Text: "parameter " + paramName,
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
