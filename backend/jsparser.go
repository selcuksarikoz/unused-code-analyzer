package main

import (
	"strings"
	"unicode"
	"unicode/utf8"
)

type TokenType int

const (
	TokenImport TokenType = iota
	TokenExport
	TokenFrom
	TokenIdentifier
	TokenString
	TokenLBrace
	TokenRBrace
	TokenLParen
	TokenRParen
	TokenComma
	TokenDot
	TokenStar
	TokenAs
	TokenNewline
	TokenEOF
	TokenEqual
	TokenAsync
	TokenUnknown
)

type Token struct {
	Type     TokenType
	Value    string
	Line     int
	Position int
}

type ImportItem struct {
	names []string
	from  string
	line  int
	text  string
}

type Tokenizer struct {
	content  string
	pos      int
	line     int
	startPos int
	tokens   []Token
}

func NewTokenizer(content string) *Tokenizer {
	return &Tokenizer{
		content: content,
		pos:     0,
		line:    1,
		tokens:  make([]Token, 0),
	}
}

func (t *Tokenizer) isAlpha(ch rune) bool {
	return unicode.IsLetter(ch) || ch == '_' || ch == '$'
}

func (t *Tokenizer) isAlphaNumeric(ch rune) bool {
	return unicode.IsLetter(ch) || unicode.IsDigit(ch) || ch == '_' || ch == '$'
}

func (t *Tokenizer) peek() rune {
	if t.pos >= len(t.content) {
		return 0
	}
	r, _ := utf8.DecodeRuneInString(t.content[t.pos:])
	return r
}

func (t *Tokenizer) next() rune {
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

func (t *Tokenizer) Tokenize() []Token {
	for t.pos < len(t.content) {
		ch := t.peek()

		if ch == 0 {
			break
		}

		if ch == '\n' {
			t.next()
			t.tokens = append(t.tokens, Token{Type: TokenNewline, Value: "\n", Line: t.line})
			continue
		}

		if unicode.IsSpace(ch) {
			t.next()
			continue
		}

		if ch == '/' {
			t.next()
			nextCh := t.peek()
			if nextCh == '/' {
				for t.peek() != '\n' && t.peek() != 0 {
					t.next()
				}
				continue
			} else if nextCh == '*' {
				t.next()
				for {
					if t.peek() == 0 {
						break
					}
					if t.peek() == '*' {
						t.next()
						if t.peek() == '/' {
							t.next()
							break
						}
					} else {
						t.next()
					}
				}
				continue
			}
			t.tokens = append(t.tokens, Token{Type: TokenUnknown, Value: "/", Line: t.line})
			continue
		}

		if ch == '"' || ch == '\'' || ch == '`' {
			t.tokens = append(t.tokens, t.readString(ch))
			continue
		}

		if t.isAlpha(ch) {
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
		case '{':
			t.next()
			t.tokens = append(t.tokens, Token{Type: TokenLBrace, Value: "{", Line: t.line})
		case '}':
			t.next()
			t.tokens = append(t.tokens, Token{Type: TokenRBrace, Value: "}", Line: t.line})
		case '(':
			t.next()
			t.tokens = append(t.tokens, Token{Type: TokenLParen, Value: "(", Line: t.line})
		case ')':
			t.next()
			t.tokens = append(t.tokens, Token{Type: TokenRParen, Value: ")", Line: t.line})
		case ',':
			t.next()
			t.tokens = append(t.tokens, Token{Type: TokenComma, Value: ",", Line: t.line})
		case '.':
			t.next()
			t.tokens = append(t.tokens, Token{Type: TokenDot, Value: ".", Line: t.line})
		case '*':
			t.next()
			t.tokens = append(t.tokens, Token{Type: TokenStar, Value: "*", Line: t.line})
		case '=':
			t.next()
			t.tokens = append(t.tokens, Token{Type: TokenEqual, Value: "=", Line: t.line})
		default:
			t.next()
			t.tokens = append(t.tokens, Token{Type: TokenUnknown, Value: string(ch), Line: t.line})
		}
	}

	t.tokens = append(t.tokens, Token{Type: TokenEOF, Value: "", Line: t.line})
	return t.tokens
}

func (t *Tokenizer) readString(quote rune) Token {
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
	return Token{Type: TokenString, Value: t.content[start:t.pos], Line: line}
}

func (t *Tokenizer) readIdentifier() Token {
	start := t.pos
	line := t.line
	for t.isAlphaNumeric(t.peek()) {
		t.next()
	}
	value := t.content[start:t.pos]

	switch value {
	case "import":
		return Token{Type: TokenImport, Value: value, Line: line}
	case "export":
		return Token{Type: TokenExport, Value: value, Line: line}
	case "from":
		return Token{Type: TokenFrom, Value: value, Line: line}
	case "as":
		return Token{Type: TokenAs, Value: value, Line: line}
	case "async":
		return Token{Type: TokenAsync, Value: value, Line: line}
	default:
		return Token{Type: TokenIdentifier, Value: value, Line: line}
	}
}

func FindJSImports(content string) []ImportItem {
	t := NewTokenizer(content)
	tokens := t.Tokenize()

	var imports []ImportItem
	var i int

	for i < len(tokens) {
		if tokens[i].Type == TokenImport {
			line := tokens[i].Line
			var names []string
			var importText strings.Builder

			i++

			// Handle `import type`
			if i < len(tokens) && tokens[i].Type == TokenIdentifier && tokens[i].Value == "type" {
				// lookahead
				nextType := TokenUnknown
				if i+1 < len(tokens) {
					nextType = tokens[i+1].Type
				}
				if nextType == TokenLBrace || nextType == TokenIdentifier || nextType == TokenStar {
					importText.WriteString("type ")
					i++
				}
			}

			// Handle default import: `import name ...`
			if i < len(tokens) && tokens[i].Type == TokenIdentifier {
				names = append(names, tokens[i].Value)
				importText.WriteString(tokens[i].Value)
				i++
				if i < len(tokens) && tokens[i].Type == TokenComma {
					importText.WriteString(", ")
					i++
				}
			}

			// Handle named & star imports
			if i < len(tokens) {
				if tokens[i].Type == TokenLBrace {
					namedImports := t.readNamedImports(tokens, &i)
					names = append(names, namedImports...)
					importText.WriteString("{ ")
					importText.WriteString(strings.Join(namedImports, ", "))
					importText.WriteString(" }")
				} else if tokens[i].Type == TokenStar {
					i++
					if i < len(tokens) && tokens[i].Type == TokenAs {
						i++
						if i < len(tokens) && tokens[i].Type == TokenIdentifier {
							names = append(names, tokens[i].Value)
							importText.WriteString("* as " + tokens[i].Value)
							i++
						}
					}
				}
			}

			// skip everything until `from` or string literal
			for i < len(tokens) && tokens[i].Type != TokenFrom && tokens[i].Type != TokenString && tokens[i].Type != TokenNewline && tokens[i].Type != TokenEOF {
				i++
			}

			if i < len(tokens) && tokens[i].Type == TokenFrom {
				i++
				if i < len(tokens) && tokens[i].Type == TokenString {
					from := tokens[i].Value
					// Only save if we actually extracted names
					if len(names) > 0 {
						imports = append(imports, ImportItem{
							names: names,
							from:  from,
							line:  line,
							text:  "import " + importText.String() + " from " + from,
						})
					}
				}
			} else if i < len(tokens) && tokens[i].Type == TokenString {
				// Handle bare import `import "module"`
				// We don't usually report bare imports as unused, but we shouldn't fail parsing.
			}
		} else {
			i++
		}
	}

	return imports
}

func (t *Tokenizer) readNamedImports(tokens []Token, i *int) []string {
	var names []string
	*i++

	for *i < len(tokens) {
		if tokens[*i].Type == TokenRBrace {
			*i++
			break
		}

		if tokens[*i].Type == TokenIdentifier {
			name := tokens[*i].Value
			*i++

			// Handle `import { type Something }`
			if name == "type" && *i < len(tokens) && tokens[*i].Type == TokenIdentifier {
				name = tokens[*i].Value
				*i++
			}

			if *i < len(tokens) && tokens[*i].Type == TokenAs {
				*i++
				if *i < len(tokens) && tokens[*i].Type == TokenIdentifier {
					name = tokens[*i].Value
					*i++
				}
			}
			names = append(names, name)
		} else if tokens[*i].Type == TokenComma {
			*i++
		} else {
			*i++
		}
	}

	return names
}

func FindUsedJSNames(content string) map[string]int {
	t := NewTokenizer(content)
	tokens := t.Tokenize()

	counts := make(map[string]int)
	reserved := map[string]bool{
		"import": true, "export": true, "from": true, "as": true,
		"const": true, "let": true, "var": true, "function": true,
		"class": true, "interface": true, "type": true, "enum": true,
		"return": true, "if": true, "else": true, "for": true, "while": true,
		"switch": true, "case": true, "break": true, "continue": true,
		"try": true, "catch": true, "finally": true, "throw": true,
		"new": true, "this": true, "super": true, "extends": true,
		"implements": true, "public": true, "private": true, "protected": true,
		"static": true, "readonly": true, "async": true, "await": true,
		"default": true, "typeof": true, "instanceof": true, "in": true,
		"of": true, "null": true, "undefined": true, "true": true, "false": true,
	}

	inImport := false

	for i, tok := range tokens {
		if tok.Type == TokenImport {
			inImport = true
			continue
		}
		if inImport && (tok.Type == TokenNewline || tok.Type == TokenUnknown && tok.Value == ";") {
			inImport = false
			continue
		}

		if tok.Type == TokenIdentifier && !reserved[tok.Value] {
			if inImport {
				continue // don't count the imported names themselves
			}

			// Skip property access `.prop`
			if i > 0 && tokens[i-1].Type == TokenDot {
				continue
			}

			// Skip method declarations in classes: `name(` where previous token is `{`, `}`, `public`, `private`, `protected`, `static`, `async`
			if i+1 < len(tokens) && tokens[i+1].Type == TokenLParen {
				if i > 0 {
					prev := tokens[i-1]
					if prev.Type == TokenLBrace || prev.Type == TokenRBrace || (prev.Type == TokenIdentifier && (prev.Value == "public" || prev.Value == "private" || prev.Value == "protected" || prev.Value == "static" || prev.Value == "async" || prev.Value == "get" || prev.Value == "set")) {
						continue
					}
				}
			}

			// Skip interface/type property declarations `prop: type;`
			if i+1 < len(tokens) && tokens[i+1].Value == ":" && tokens[i+1].Type == TokenUnknown {
				if i > 0 {
					prev := tokens[i-1]
					if prev.Type == TokenLBrace || prev.Type == TokenRBrace || prev.Type == TokenNewline || (prev.Type == TokenUnknown && prev.Value == ";") {
						continue
					}
				}
			}

			counts[tok.Value]++
		}
	}

	return counts
}

type JSVariable struct {
	Name string
	Line int
	Type string
}

type JSParameter struct {
	Name string
	Line int
}

func FindJSVariables(content string) []JSVariable {
	t := NewTokenizer(content)
	tokens := t.Tokenize()

	var vars []JSVariable
	reserved := map[string]bool{
		"import": true, "export": true, "from": true, "as": true,
		"const": true, "let": true, "var": true, "function": true,
		"class": true, "interface": true, "type": true, "enum": true,
		"return": true, "if": true, "else": true, "for": true, "while": true,
		"switch": true, "case": true, "break": true, "continue": true,
		"try": true, "catch": true, "finally": true, "throw": true,
		"new": true, "this": true, "super": true, "extends": true,
		"implements": true, "public": true, "private": true, "protected": true,
		"static": true, "readonly": true, "async": true, "await": true,
		"default": true, "typeof": true, "instanceof": true, "in": true,
		"of": true, "null": true, "undefined": true, "true": true, "false": true,
	}

	for i := 0; i < len(tokens); i++ {
		tok := tokens[i]
		if tok.Type == TokenIdentifier && (tok.Value == "const" || tok.Value == "let" || tok.Value == "var") {
			if i+1 < len(tokens) && tokens[i+1].Type == TokenIdentifier {
				name := tokens[i+1].Value
				if !reserved[name] {
					vars = append(vars, JSVariable{
						Name: name,
						Line: tokens[i+1].Line,
						Type: tok.Value,
					})
				}
			}
		}
		if tok.Type == TokenIdentifier && tok.Value == "function" {
			if i+1 < len(tokens) && tokens[i+1].Type == TokenIdentifier {
				name := tokens[i+1].Value
				if !reserved[name] {
					vars = append(vars, JSVariable{
						Name: name,
						Line: tokens[i+1].Line,
						Type: "function",
					})
				}
			}
		}
		if tok.Type == TokenIdentifier && tok.Value == "class" {
			if i+1 < len(tokens) && tokens[i+1].Type == TokenIdentifier {
				name := tokens[i+1].Value
				if !reserved[name] {
					vars = append(vars, JSVariable{
						Name: name,
						Line: tokens[i+1].Line,
						Type: "class",
					})
				}
			}
		}
	}

	return vars
}

func extractParams(tokens []Token, startIdx int) []JSParameter {
	var params []JSParameter
	expectingParam := true
	for i := startIdx; i < len(tokens) && tokens[i].Type != TokenRParen && tokens[i].Type != TokenEOF; i++ {
		if tokens[i].Type == TokenIdentifier && expectingParam {
			reserved := map[string]bool{"true": true, "false": true, "null": true, "undefined": true, "this": true}
			if !reserved[tokens[i].Value] {
				params = append(params, JSParameter{Name: tokens[i].Value, Line: tokens[i].Line})
			}
			expectingParam = false
		} else if tokens[i].Type == TokenComma {
			expectingParam = true
		} else if tokens[i].Type == TokenLBrace || tokens[i].Type == TokenRBrace || tokens[i].Type == TokenEqual {
			// skip handling destructing for parameters for simple tokenizer, just reset expectingParam carefully
		}
	}
	return params
}

func FindJSParameters(content string) []JSParameter {
	t := NewTokenizer(content)
	tokens := t.Tokenize()

	var params []JSParameter

	for i := 0; i < len(tokens); i++ {
		tok := tokens[i]
		if tok.Type == TokenIdentifier && tok.Value == "function" {
			i++
			for i < len(tokens) && tokens[i].Type != TokenLParen {
				i++
			}
			if i < len(tokens) && tokens[i].Type == TokenLParen {
				i++
				funcParams := extractParams(tokens, i)
				params = append(params, funcParams...)
			}
		}
		if tok.Type == TokenIdentifier && (tok.Value == "arrow" || (i > 0 && tokens[i-1].Type == TokenEqual && tokens[i].Type == TokenIdentifier)) {
			// very naive arrow check
			arrowFunc := false
			for j := i; j < len(tokens) && j < i+5; j++ {
				if tokens[j].Type == TokenLParen {
					arrowFunc = true
					break
				}
				if tokens[j].Type == TokenIdentifier || tokens[j].Type == TokenNewline {
					break
				}
			}
			if arrowFunc && i > 0 {
				startIdx := i
				if tokens[startIdx].Type == TokenIdentifier {
					for startIdx < len(tokens) && tokens[startIdx].Type != TokenLParen {
						startIdx++
					}
					if startIdx < len(tokens) && tokens[startIdx].Type == TokenLParen {
						startIdx++
						funcParams := extractParams(tokens, startIdx)
						params = append(params, funcParams...)
					}
				}
			}
		}
		if tok.Type == TokenAsync {
			for j := i + 1; j < len(tokens) && j < i+10; j++ {
				if tokens[j].Type == TokenLParen && j > i && tokens[j-1].Type == TokenIdentifier {
					paramStart := j + 1
					funcParams := extractParams(tokens, paramStart)
					params = append(params, funcParams...)
					break
				}
			}
		}
	}

	return params
}

func analyzeJSTS(content, filename string) AnalysisResult {
	imports := FindJSImports(content)
	counts := FindUsedJSNames(content)

	var unusedImports []CodeIssue
	for _, imp := range imports {
		allUnused := true
		for _, name := range imp.names {
			if counts[name] > 0 {
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

	return AnalysisResult{
		Imports:    unusedImports,
		Variables:  []CodeIssue{},
		Parameters: []CodeIssue{},
	}
}

func analyzeJSTSForWorkspace(content, filename string) ([]Definition, []Import, []CodeIssue, []CodeIssue) {
	imports := FindJSImports(content)
	variables := FindJSVariables(content)

	var allImportNames []string
	for _, imp := range imports {
		allImportNames = append(allImportNames, imp.names...)
	}

	counts := FindUsedJSNames(content)

	var outDefs []Definition
	var outImports []Import

	for _, v := range variables {
		outDefs = append(outDefs, Definition{
			Name: v.Name,
			File: filename,
			Line: v.Line,
			Type: v.Type,
		})
	}

	var unusedImports []CodeIssue
	for _, imp := range imports {
		outImports = append(outImports, Import{
			Name: strings.Join(imp.names, ", "),
			File: filename,
			Line: imp.line,
		})

		allUnused := true
		for _, name := range imp.names {
			if counts[name] > 0 {
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

	var unusedVars []CodeIssue
	for _, v := range variables {
		unusedVars = append(unusedVars, CodeIssue{
			ID:   generateUUID(),
			Line: v.Line,
			Text: v.Name + " (variable)",
			File: filename,
		})
	}

	return outDefs, outImports, unusedImports, unusedVars
}

func buildResultJSTS(file AnalyzeFile, defs []Definition, imports []Import, usedNames map[string]bool, allFiles []AnalyzeFile) AnalysisResult {
	localImports := FindJSImports(file.Content)
	localVars := FindJSVariables(file.Content)
	localParams := FindJSParameters(file.Content)

	counts := FindUsedJSNames(file.Content)

	var unusedImports []CodeIssue
	for _, imp := range localImports {
		allUnused := true
		for _, name := range imp.names {
			if usedNames[name+"@"+file.Filename] {
				allUnused = false
				break
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
	for _, v := range localVars {
		key := v.Name + "@" + file.Filename
		isLocallyUsed := counts[v.Name] > 1
		isCrossFileUsed := usedNames[key]
		if !isLocallyUsed && !isCrossFileUsed {
			unusedVars = append(unusedVars, CodeIssue{
				ID:   generateUUID(),
				Line: v.Line,
				Text: v.Name + " (variable)",
				File: file.Filename,
			})
		}
	}

	var unusedParams []CodeIssue
	for _, p := range localParams {
		isLocallyUsed := counts[p.Name] > 1
		if !isLocallyUsed {
			unusedParams = append(unusedParams, CodeIssue{
				ID:   generateUUID(),
				Line: p.Line,
				Text: p.Name + " (parameter)",
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
