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

	for t.pos < len(t.content) {
		ch := t.peek()

		if inString {
			if ch == stringChar && t.pos > 0 && t.content[t.pos-1] != '\\' {
				t.next()
				inString = false
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
		if tokens[i].Type == RubyTokenRequire && i+1 < len(tokens) && tokens[i+1].Type == RubyTokenString {
			line := tokens[i].Line
			path := strings.Trim(tokens[i+1].Value, `"`)
			name := path
			if strings.Contains(path, "/") {
				parts := strings.Split(path, "/")
				name = parts[len(parts)-1]
			}
			name = strings.TrimSuffix(name, ".rb")

			imports = append(imports, RubyImportItem{
				name: name,
				path: path,
				line: line,
				text: "require " + tokens[i+1].Value,
			})
			i += 2
			continue
		}

		if tokens[i].Type == RubyTokenRequireRelative && i+1 < len(tokens) && tokens[i+1].Type == RubyTokenString {
			line := tokens[i].Line
			path := strings.Trim(tokens[i+1].Value, `"`)
			name := path
			if strings.Contains(path, "/") {
				parts := strings.Split(path, "/")
				name = parts[len(parts)-1]
			}
			name = strings.TrimSuffix(name, ".rb")

			imports = append(imports, RubyImportItem{
				name: name,
				path: path,
				line: line,
				text: "require_relative " + tokens[i+1].Value,
			})
			i += 2
			continue
		}

		i++
	}

	return imports
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
	counts := FindUsedRubyNames(content)

	var outImports []Import
	var unusedImports []CodeIssue

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

	return []Definition{}, outImports, unusedImports, []CodeIssue{}
}

func buildResultRuby(file AnalyzeFile, defs []Definition, imports []Import, usedNames map[string]bool, allFiles []AnalyzeFile) AnalysisResult {
	localImports := FindRubyImports(file.Content)

	var unusedImports []CodeIssue
	for _, imp := range localImports {
		if !usedNames[imp.name+"@"+file.Filename] && imp.name != "" {
			unusedImports = append(unusedImports, CodeIssue{
				ID:   generateUUID(),
				Line: imp.line,
				Text: imp.text,
				File: file.Filename,
			})
		}
	}

	return AnalysisResult{
		Imports:    unusedImports,
		Variables:  []CodeIssue{},
		Parameters: []CodeIssue{},
	}
}
