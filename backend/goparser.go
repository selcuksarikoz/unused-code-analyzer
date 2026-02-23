package main

import (
	"strings"
	"unicode"
	"unicode/utf8"
)

type GoTokenType int

const (
	GoTokenImport GoTokenType = iota
	GoTokenPackage
	GoTokenFunc
	GoTokenTypeDef
	GoTokenStruct
	GoTokenInterface
	GoTokenVar
	GoTokenConst
	GoTokenIdentifier
	GoTokenString
	GoTokenLParen
	GoTokenRParen
	GoTokenLBrace
	GoTokenRBrace
	GoTokenComma
	GoTokenDot
	GoTokenColon
	GoTokenEqual
	GoTokenNewline
	GoTokenEOF
	GoTokenUnknown
)

type GoToken struct {
	Type  GoTokenType
	Value string
	Line  int
}

type GoTokenizer struct {
	content string
	pos     int
	line    int
	tokens  []GoToken
}

func NewGoTokenizer(content string) *GoTokenizer {
	return &GoTokenizer{
		content: content,
		pos:     0,
		line:    1,
		tokens:  make([]GoToken, 0),
	}
}

func (t *GoTokenizer) peek() rune {
	if t.pos >= len(t.content) {
		return 0
	}
	r, _ := utf8.DecodeRuneInString(t.content[t.pos:])
	return r
}

func (t *GoTokenizer) next() rune {
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

func (t *GoTokenizer) Tokenize() []GoToken {
	inString := false
	var stringChar rune

	for t.pos < len(t.content) {
		ch := t.peek()

		if inString {
			if ch == stringChar {
				t.next()
				inString = false
			} else if ch == '\\' {
				t.next()
				t.next()
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
			t.tokens = append(t.tokens, GoToken{Type: GoTokenNewline, Value: "\n", Line: t.line})
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
			t.tokens = append(t.tokens, GoToken{Type: GoTokenUnknown, Value: "/", Line: t.line})
			continue
		}

		if ch == '"' || ch == '\'' || ch == '`' {
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
			for unicode.IsDigit(t.peek()) || t.peek() == '.' {
				t.next()
			}
			continue
		}

		switch ch {
		case '(':
			t.next()
			t.tokens = append(t.tokens, GoToken{Type: GoTokenLParen, Value: "(", Line: t.line})
		case ')':
			t.next()
			t.tokens = append(t.tokens, GoToken{Type: GoTokenRParen, Value: ")", Line: t.line})
		case '{':
			t.next()
			t.tokens = append(t.tokens, GoToken{Type: GoTokenLBrace, Value: "{", Line: t.line})
		case '}':
			t.next()
			t.tokens = append(t.tokens, GoToken{Type: GoTokenRBrace, Value: "}", Line: t.line})
		case ',':
			t.next()
			t.tokens = append(t.tokens, GoToken{Type: GoTokenComma, Value: ",", Line: t.line})
		case '.':
			t.next()
			t.tokens = append(t.tokens, GoToken{Type: GoTokenDot, Value: ".", Line: t.line})
		case ':':
			t.next()
			t.tokens = append(t.tokens, GoToken{Type: GoTokenColon, Value: ":", Line: t.line})
		case '=':
			t.next()
			t.tokens = append(t.tokens, GoToken{Type: GoTokenEqual, Value: "=", Line: t.line})
		default:
			t.next()
		}
	}

	t.tokens = append(t.tokens, GoToken{Type: GoTokenEOF, Value: "", Line: t.line})
	return t.tokens
}

func (t *GoTokenizer) readIdentifier() GoToken {
	start := t.pos
	line := t.line
	for unicode.IsLetter(t.peek()) || unicode.IsDigit(t.peek()) || t.peek() == '_' {
		t.next()
	}
	value := t.content[start:t.pos]

	switch value {
	case "import":
		return GoToken{Type: GoTokenImport, Value: value, Line: line}
	case "package":
		return GoToken{Type: GoTokenPackage, Value: value, Line: line}
	case "func":
		return GoToken{Type: GoTokenFunc, Value: value, Line: line}
	case "type":
		return GoToken{Type: GoTokenTypeDef, Value: value, Line: line}
	case "struct":
		return GoToken{Type: GoTokenStruct, Value: value, Line: line}
	case "interface":
		return GoToken{Type: GoTokenInterface, Value: value, Line: line}
	case "var":
		return GoToken{Type: GoTokenVar, Value: value, Line: line}
	case "const":
		return GoToken{Type: GoTokenConst, Value: value, Line: line}
	default:
		return GoToken{Type: GoTokenIdentifier, Value: value, Line: line}
	}
}

type GoImportItem struct {
	name  string
	alias string
	path  string
	line  int
	text  string
}

type GoVariable struct {
	Name string
	Line int
	Type string
}

type GoParameter struct {
	Name string
	Line int
}

func FindGoVariables(content string) []GoVariable {
	t := NewGoTokenizer(content)
	tokens := t.Tokenize()

	var vars []GoVariable

	for i := 0; i < len(tokens); i++ {
		tok := tokens[i]

		if tok.Type == GoTokenVar || tok.Type == GoTokenConst {
			if i+1 < len(tokens) && tokens[i+1].Type == GoTokenIdentifier {
				name := tokens[i+1].Value
				vars = append(vars, GoVariable{
					Name: name,
					Line: tokens[i+1].Line,
					Type: "variable",
				})
			}
		}

		if tok.Type == GoTokenFunc {
			if i+1 < len(tokens) && tokens[i+1].Type == GoTokenIdentifier {
				name := tokens[i+1].Value
				vars = append(vars, GoVariable{
					Name: name,
					Line: tokens[i+1].Line,
					Type: "function",
				})
			}
		}

		if tok.Type == GoTokenTypeDef {
			if i+1 < len(tokens) && tokens[i+1].Type == GoTokenIdentifier {
				name := tokens[i+1].Value
				vars = append(vars, GoVariable{
					Name: name,
					Line: tokens[i+1].Line,
					Type: "type",
				})
			}
		}

		if tok.Type == GoTokenIdentifier && i+2 < len(tokens) {
			if tokens[i+1].Type == GoTokenColon && tokens[i+2].Type == GoTokenEqual {
				name := tok.Value
				vars = append(vars, GoVariable{
					Name: name,
					Line: tok.Line,
					Type: "variable",
				})
			}
		}
	}

	return vars
}

func FindGoParameters(content string) []GoParameter {
	t := NewGoTokenizer(content)
	tokens := t.Tokenize()

	var params []GoParameter

	for i := 0; i < len(tokens); i++ {
		tok := tokens[i]
		if tok.Type == GoTokenFunc {
			i++
			for i < len(tokens) && tokens[i].Type != GoTokenLParen {
				i++
			}
			if i < len(tokens) && tokens[i].Type == GoTokenLParen {
				i++
				depth := 1
				for i < len(tokens) && depth > 0 {
					if tokens[i].Type == GoTokenLParen {
						depth++
					} else if tokens[i].Type == GoTokenRParen {
						depth--
					} else if tokens[i].Type == GoTokenIdentifier && depth == 1 {
						params = append(params, GoParameter{
							Name: tokens[i].Value,
							Line: tokens[i].Line,
						})
					}
					i++
				}
			}
		}
	}

	return params
}

func FindGoImports(content string) []GoImportItem {
	t := NewGoTokenizer(content)
	tokens := t.Tokenize()

	var imports []GoImportItem

	var i int
	for i < len(tokens) {
		if tokens[i].Type == GoTokenImport && i+1 < len(tokens) && tokens[i+1].Type == GoTokenLParen {
			importLine := tokens[i].Line
			i += 2

			for i < len(tokens) && tokens[i].Type != GoTokenRParen {
				if tokens[i].Type == GoTokenString {
					path := strings.Trim(tokens[i].Value, `"`)
					name := path
					if strings.Contains(path, "/") {
						parts := strings.Split(path, "/")
						name = parts[len(parts)-1]
					}
					imports = append(imports, GoImportItem{
						name: name,
						path: path,
						line: importLine,
						text: "import " + tokens[i].Value,
					})
				}
				i++
			}
		}

		if tokens[i].Type == GoTokenImport && i+1 < len(tokens) && tokens[i+1].Type == GoTokenString {
			line := tokens[i].Line
			path := strings.Trim(tokens[i+1].Value, `"`)
			name := path
			if strings.Contains(path, "/") {
				parts := strings.Split(path, "/")
				name = parts[len(parts)-1]
			}
			imports = append(imports, GoImportItem{
				name: name,
				path: path,
				line: line,
				text: "import " + tokens[i+1].Value,
			})
			i += 2
		}

		i++
	}

	return imports
}

func FindUsedGoNames(content string) map[string]int {
	t := NewGoTokenizer(content)
	tokens := t.Tokenize()

	counts := make(map[string]int)
	reserved := map[string]bool{
		"import": true, "package": true, "func": true, "type": true,
		"struct": true, "interface": true, "var": true, "const": true,
		"return": true, "if": true, "else": true, "for": true, "range": true,
		"switch": true, "case": true, "default": true, "break": true,
		"continue": true, "go": true, "chan": true, "select": true,
		"defer": true, "map": true, "make": true, "new": true,
		"append": true, "copy": true, "delete": true, "len": true,
		"cap": true, "panic": true, "recover": true, "close": true,
		"nil": true, "true": true, "false": true, "iota": true,
	}

	for _, tok := range tokens {
		if tok.Type == GoTokenIdentifier && !reserved[tok.Value] {
			counts[tok.Value]++
		}
	}

	return counts
}

func analyzeGo(content, filename string) AnalysisResult {
	imports := FindGoImports(content)
	counts := FindUsedGoNames(content)
	vars := FindGoVariables(content)
	params := FindGoParameters(content)

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

	var unusedVars []CodeIssue
	for _, v := range vars {
		if counts[v.Name] <= 1 {
			unusedVars = append(unusedVars, CodeIssue{
				ID:   generateUUID(),
				Line: v.Line,
				Text: v.Name + " (" + v.Type + ")",
				File: filename,
			})
		}
	}

	var unusedParams []CodeIssue
	for _, p := range params {
		if counts[p.Name] <= 1 {
			unusedParams = append(unusedParams, CodeIssue{
				ID:   generateUUID(),
				Line: p.Line,
				Text: p.Name + " (parameter)",
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

func analyzeGoForWorkspace(content, filename string) ([]Definition, []Import, []CodeIssue, []CodeIssue) {
	imports := FindGoImports(content)
	counts := FindUsedGoNames(content)
	vars := FindGoVariables(content)

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

	var defs []Definition
	for _, v := range vars {
		defs = append(defs, Definition{
			Name: v.Name,
			File: filename,
			Line: v.Line,
		})
	}

	return defs, outImports, unusedImports, []CodeIssue{}
}

func buildResultGo(file AnalyzeFile, defs []Definition, imports []Import, usedNames map[string]bool, allFiles []AnalyzeFile) AnalysisResult {
	localImports := FindGoImports(file.Content)
	counts := FindUsedGoNames(file.Content)
	vars := FindGoVariables(file.Content)
	params := FindGoParameters(file.Content)

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
	for _, v := range vars {
		isCrossFileUsed := usedNames[v.Name+"@"+file.Filename]
		isLocallyUsed := counts[v.Name] > 1
		if !isCrossFileUsed && !isLocallyUsed {
			unusedVars = append(unusedVars, CodeIssue{
				ID:   generateUUID(),
				Line: v.Line,
				Text: v.Name + " (" + v.Type + ")",
				File: file.Filename,
			})
		}
	}

	var unusedParams []CodeIssue
	for _, p := range params {
		isCrossFileUsed := usedNames[p.Name+"@"+file.Filename]
		isLocallyUsed := counts[p.Name] > 1
		if !isCrossFileUsed && !isLocallyUsed {
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
