package main

import (
	"strings"
	"unicode"
	"unicode/utf8"
)

type RubyTokenType int

const (
	RubyTokenRequire RubyTokenType = iota
	RubyTokenRequireRelative
	RubyTokenDef
	RubyTokenClass
	RubyTokenModule
	RubyTokenIdentifier
	RubyTokenString
	RubyTokenLParen
	RubyTokenRParen
	RubyTokenComma
	RubyTokenDot
	RubyTokenNewline
	RubyTokenEOF
	RubyTokenUnknown
)

type RubyToken struct {
	Type  RubyTokenType
	Value string
	Line  int
}

type RubyTokenizer struct {
	content string
	pos     int
	line    int
	tokens  []RubyToken
}

func NewRubyTokenizer(content string) *RubyTokenizer {
	return &RubyTokenizer{
		content: content,
		pos:     0,
		line:    1,
		tokens:  make([]RubyToken, 0),
	}
}

func (t *RubyTokenizer) peek() rune {
	if t.pos >= len(t.content) {
		return 0
	}
	r, _ := utf8.DecodeRuneInString(t.content[t.pos:])
	return r
}

func (t *RubyTokenizer) next() rune {
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

func (t *RubyTokenizer) Tokenize() []RubyToken {
	inString := false
	var stringChar rune
	var stringStart int
	var stringLine int

	for t.pos < len(t.content) {
		ch := t.peek()

		if inString {
			if ch == stringChar && t.pos > 0 && t.content[t.pos-1] != '\\' {
				t.next()
				inString = false
				t.tokens = append(t.tokens, RubyToken{Type: RubyTokenString, Value: t.content[stringStart:t.pos], Line: stringLine})
			} else {
				t.next()
			}
			continue
		}

		if ch == 0 {
			break
		}

		if ch == '\n' {
			t.next()
			t.tokens = append(t.tokens, RubyToken{Type: RubyTokenNewline, Value: "\n", Line: t.line})
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
			inString = true
			stringChar = ch
			stringStart = t.pos
			stringLine = t.line
			t.next()
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
			t.tokens = append(t.tokens, RubyToken{Type: RubyTokenLParen, Value: "(", Line: t.line})
		case ')':
			t.next()
			t.tokens = append(t.tokens, RubyToken{Type: RubyTokenRParen, Value: ")", Line: t.line})
		case ',':
			t.next()
			t.tokens = append(t.tokens, RubyToken{Type: RubyTokenComma, Value: ",", Line: t.line})
		case '.':
			t.next()
			t.tokens = append(t.tokens, RubyToken{Type: RubyTokenDot, Value: ".", Line: t.line})
		default:
			t.next()
		}
	}

	t.tokens = append(t.tokens, RubyToken{Type: RubyTokenEOF, Value: "", Line: t.line})
	return t.tokens
}

func (t *RubyTokenizer) readIdentifier() RubyToken {
	start := t.pos
	line := t.line
	for unicode.IsLetter(t.peek()) || unicode.IsDigit(t.peek()) || t.peek() == '_' {
		t.next()
	}
	value := t.content[start:t.pos]

	switch value {
	case "require":
		return RubyToken{Type: RubyTokenRequire, Value: value, Line: line}
	case "require_relative":
		return RubyToken{Type: RubyTokenRequireRelative, Value: value, Line: line}
	case "def":
		return RubyToken{Type: RubyTokenDef, Value: value, Line: line}
	case "class":
		return RubyToken{Type: RubyTokenClass, Value: value, Line: line}
	case "module":
		return RubyToken{Type: RubyTokenModule, Value: value, Line: line}
	default:
		return RubyToken{Type: RubyTokenIdentifier, Value: value, Line: line}
	}
}

type RubyImportItem struct {
	name string
	path string
	line int
	text string
}

func FindRubyImports(content string) []RubyImportItem {
	t := NewRubyTokenizer(content)
	tokens := t.Tokenize()

	var imports []RubyImportItem

	var i int
	for i < len(tokens) {
		if (tokens[i].Type == RubyTokenRequire || tokens[i].Type == RubyTokenRequireRelative) && i+1 < len(tokens) {
			line := tokens[i].Line

			var path string
			if tokens[i+1].Type == RubyTokenString {
				path = strings.Trim(tokens[i+1].Value, `"'`)
			} else if tokens[i+1].Type == RubyTokenIdentifier {
				path = tokens[i+1].Value
			}

			if path == "" {
				i++
				continue
			}

			name := path
			if strings.Contains(path, "/") {
				parts := strings.Split(path, "/")
				name = parts[len(parts)-1]
			}
			name = strings.TrimSuffix(name, ".rb")

			text := "require"
			if tokens[i].Type == RubyTokenRequireRelative {
				text = "require_relative"
			}
			if tokens[i+1].Type == RubyTokenString {
				text += " " + tokens[i+1].Value
			} else {
				text += " " + path
			}

			imports = append(imports, RubyImportItem{
				name: name,
				path: path,
				line: line,
				text: text,
			})
			i += 2
			continue
		}

		i++
	}

	return imports
}

type RubyDefinition struct {
	name    string
	defType string
	line    int
}

func FindRubyDefinitions(content string) []RubyDefinition {
	t := NewRubyTokenizer(content)
	tokens := t.Tokenize()

	var defs []RubyDefinition

	for i := 0; i < len(tokens); i++ {
		if tokens[i].Type == RubyTokenDef && i+1 < len(tokens) {
			name := ""
			if tokens[i+1].Type == RubyTokenIdentifier {
				if strings.Contains(tokens[i+1].Value, ".") {
					parts := strings.Split(tokens[i+1].Value, ".")
					name = parts[len(parts)-1]
				} else {
					name = tokens[i+1].Value
				}
				defs = append(defs, RubyDefinition{
					name:    name,
					defType: "method",
					line:    tokens[i].Line,
				})
			}
		}
		if tokens[i].Type == RubyTokenClass && i+1 < len(tokens) && tokens[i+1].Type == RubyTokenIdentifier {
			defs = append(defs, RubyDefinition{
				name:    tokens[i+1].Value,
				defType: "class",
				line:    tokens[i].Line,
			})
		}
		if tokens[i].Type == RubyTokenModule && i+1 < len(tokens) && tokens[i+1].Type == RubyTokenIdentifier {
			defs = append(defs, RubyDefinition{
				name:    tokens[i+1].Value,
				defType: "module",
				line:    tokens[i].Line,
			})
		}
	}

	return defs
}

func FindRubyParameters(content, filename string) []CodeIssue {
	t := NewRubyTokenizer(content)
	tokens := t.Tokenize()

	var params []CodeIssue
	for i := 0; i < len(tokens); i++ {
		if tokens[i].Type == RubyTokenDef && i+1 < len(tokens) && tokens[i+1].Type == RubyTokenIdentifier {
			parenCount := 0
			paramStart := -1
			for j := i + 2; j < len(tokens); j++ {
				if tokens[j].Type == RubyTokenLParen {
					if parenCount == 0 {
						paramStart = j + 1
					}
					parenCount++
				}
				if tokens[j].Type == RubyTokenRParen {
					parenCount--
					if parenCount == 0 {
						break
					}
				}
			}

			if paramStart > 0 && paramStart < len(tokens) {
				for j := paramStart; j < len(tokens) && tokens[j].Type != RubyTokenRParen; j++ {
					if tokens[j].Type == RubyTokenIdentifier && j+1 < len(tokens) && (tokens[j+1].Type == RubyTokenComma || tokens[j+1].Type == RubyTokenRParen) {
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

func FindUsedRubyNames(content string) map[string]int {
	t := NewRubyTokenizer(content)
	tokens := t.Tokenize()

	counts := make(map[string]int)
	reserved := map[string]bool{
		"require": true, "require_relative": true, "def": true, "class": true,
		"module": true, "end": true, "if": true, "elsif": true, "else": true,
		"unless": true, "case": true, "when": true, "while": true, "until": true,
		"for": true, "do": true, "begin": true, "rescue": true, "ensure": true,
		"raise": true, "return": true, "yield": true, "break": true, "next": true,
		"redo": true, "retry": true, "self": true, "super": true, "true": true,
		"false": true, "nil": true, "and": true, "or": true, "not": true,
		"in": true, "then": true, "alias": true, "defined": true, "lambda": true,
	}

	for _, tok := range tokens {
		if tok.Type == RubyTokenIdentifier && !reserved[tok.Value] {
			counts[tok.Value]++
		}
	}

	return counts
}

func analyzeRuby(content, filename string) AnalysisResult {
	imports := FindRubyImports(content)
	counts := FindUsedRubyNames(content)

	var unusedImports []CodeIssue
	for _, imp := range imports {
		if counts[imp.name] <= 1 && imp.name != "" {
			unusedImports = append(unusedImports, CodeIssue{
				ID:   generateUUID(),
				Line: imp.line,
				Text: imp.text,
				File: filename,
			})
		}
	}

	return AnalysisResult{
		Imports:    unusedImports,
		Variables:  []CodeIssue{},
		Parameters: []CodeIssue{},
	}
}

func analyzeRubyForWorkspace(content, filename string) ([]Definition, []Import, []CodeIssue, []CodeIssue) {
	imports := FindRubyImports(content)
	defs := FindRubyDefinitions(content)
	counts := FindUsedRubyNames(content)

	var outImports []Import
	var unusedImports []CodeIssue
	var outDefs []Definition

	for _, imp := range imports {
		outImports = append(outImports, Import{
			Name: imp.name,
			File: filename,
			Line: imp.line,
		})

		if counts[imp.name] <= 1 && imp.name != "" {
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

func buildResultRuby(file AnalyzeFile, defs []Definition, imports []Import, usedNames map[string]bool, allFiles []AnalyzeFile) AnalysisResult {
	localImports := FindRubyImports(file.Content)
	localDefs := FindRubyDefinitions(file.Content)
	parameters := FindRubyParameters(file.Content, file.Filename)
	counts := FindUsedRubyNames(file.Content)

	var unusedImports []CodeIssue
	for _, imp := range localImports {
		isCrossFileUsed := usedNames[imp.name+"@"+file.Filename]
		isLocallyUsed := counts[imp.name] > 1
		if !isCrossFileUsed && !isLocallyUsed && imp.name != "" {
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
		key := d.name + "@" + file.Filename
		if !usedNames[key] {
			unusedVars = append(unusedVars, CodeIssue{
				ID:   generateUUID(),
				Line: d.line,
				Text: d.defType + " " + d.name,
				File: file.Filename,
			})
		}
	}

	paramUsed := make(map[string]bool)
	for _, p := range parameters {
		paramName := strings.TrimSpace(strings.TrimSuffix(p.Text, " (parameter)"))
		key := paramName + "@" + file.Filename
		if usedNames[key] {
			paramUsed[paramName] = true
		}
	}

	var unusedParams []CodeIssue
	for _, p := range parameters {
		paramName := strings.TrimSpace(strings.TrimSuffix(p.Text, " (parameter)"))
		if !paramUsed[paramName] {
			unusedParams = append(unusedParams, p)
		}
	}

	return AnalysisResult{
		Imports:    unusedImports,
		Variables:  unusedVars,
		Parameters: unusedParams,
	}
}
