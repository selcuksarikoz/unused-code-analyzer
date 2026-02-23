package main

import (
	"strings"
	"unicode"
	"unicode/utf8"
)

type PHPTokenType int

const (
	PHPTokenUse PHPTokenType = iota
	PHPTokenFunction
	PHPTokenClass
	PHPTokenInterface
	PHPTokenTrait
	PHPTokenExtends
	PHPTokenImplements
	PHPTokenIdentifier
	PHPTokenString
	PHPTokenLParen
	PHPTokenRParen
	PHPTokenSemi
	PHPTokenNamespace
	PHPTokenNewline
	PHPTokenEOF
	PHPTokenUnknown
)

type PHPToken struct {
	Type  PHPTokenType
	Value string
	Line  int
}

type PHPTokenizer struct {
	content string
	pos     int
	line    int
	tokens  []PHPToken
}

func NewPHPTokenizer(content string) *PHPTokenizer {
	return &PHPTokenizer{
		content: content,
		pos:     0,
		line:    1,
		tokens:  make([]PHPToken, 0),
	}
}

func (t *PHPTokenizer) peek() rune {
	if t.pos >= len(t.content) {
		return 0
	}
	r, _ := utf8.DecodeRuneInString(t.content[t.pos:])
	return r
}

func (t *PHPTokenizer) next() rune {
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

func (t *PHPTokenizer) Tokenize() []PHPToken {
	inString := false
	var stringChar rune
	var stringStart int
	var stringLine int

	for t.pos < len(t.content) {
		ch := t.peek()

		if ch == '<' && t.pos+1 < len(t.content) && t.content[t.pos:t.pos+2] == "<?" {
			t.next()
			t.next()
			continue
		}

		if inString {
			if ch == stringChar && t.pos > 0 && t.content[t.pos-1] != '\\' {
				t.next()
				inString = false
				t.tokens = append(t.tokens, PHPToken{Type: PHPTokenString, Value: t.content[stringStart:t.pos], Line: stringLine})
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
			t.tokens = append(t.tokens, PHPToken{Type: PHPTokenNewline, Value: "\n", Line: t.line})
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
			t.tokens = append(t.tokens, PHPToken{Type: PHPTokenUnknown, Value: "/", Line: t.line})
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

		if unicode.IsLetter(ch) || ch == '_' || ch == '\\' {
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
			t.tokens = append(t.tokens, PHPToken{Type: PHPTokenLParen, Value: "(", Line: t.line})
		case ')':
			t.next()
			t.tokens = append(t.tokens, PHPToken{Type: PHPTokenRParen, Value: ")", Line: t.line})
		case ';':
			t.next()
			t.tokens = append(t.tokens, PHPToken{Type: PHPTokenSemi, Value: ";", Line: t.line})
		default:
			t.next()
		}
	}

	t.tokens = append(t.tokens, PHPToken{Type: PHPTokenEOF, Value: "", Line: t.line})
	return t.tokens
}

func (t *PHPTokenizer) readIdentifier() PHPToken {
	start := t.pos
	line := t.line
	for unicode.IsLetter(t.peek()) || unicode.IsDigit(t.peek()) || t.peek() == '_' || t.peek() == '\\' {
		t.next()
	}
	value := t.content[start:t.pos]

	switch value {
	case "use":
		return PHPToken{Type: PHPTokenUse, Value: value, Line: line}
	case "function":
		return PHPToken{Type: PHPTokenFunction, Value: value, Line: line}
	case "class":
		return PHPToken{Type: PHPTokenClass, Value: value, Line: line}
	case "interface":
		return PHPToken{Type: PHPTokenInterface, Value: value, Line: line}
	case "trait":
		return PHPToken{Type: PHPTokenTrait, Value: value, Line: line}
	case "extends":
		return PHPToken{Type: PHPTokenExtends, Value: value, Line: line}
	case "implements":
		return PHPToken{Type: PHPTokenImplements, Value: value, Line: line}
	case "namespace":
		return PHPToken{Type: PHPTokenNamespace, Value: value, Line: line}
	default:
		return PHPToken{Type: PHPTokenIdentifier, Value: value, Line: line}
	}
}

type PHPImportItem struct {
	name     string
	fullPath string
	line     int
	text     string
}

func FindPHPImports(content string) []PHPImportItem {
	t := NewPHPTokenizer(content)
	tokens := t.Tokenize()

	imports := findPHPImportsFromTokens(tokens)
	return imports
}

func findPHPImportsFromTokens(tokens []PHPToken) []PHPImportItem {
	var imports []PHPImportItem

	var i int
	for i < len(tokens) {
		if tokens[i].Type == PHPTokenUse {
			line := tokens[i].Line
			useType := "class"
			useIdx := i + 1

			if useIdx < len(tokens) && (tokens[useIdx].Type == PHPTokenFunction || tokens[useIdx].Value == "function") {
				useType = "function"
				useIdx++
			} else if useIdx < len(tokens) && tokens[useIdx].Value == "const" {
				useType = "const"
				useIdx++
			}

			if useIdx >= len(tokens) || tokens[useIdx].Type != PHPTokenIdentifier {
				i++
				continue
			}

			var fullPath []string
			fullPath = append(fullPath, tokens[useIdx].Value)
			useIdx++

			for useIdx < len(tokens) && tokens[useIdx].Type != PHPTokenSemi {
				if tokens[useIdx].Type == PHPTokenIdentifier {
					fullPath = append(fullPath, tokens[useIdx].Value)
				}
				useIdx++
			}

			path := strings.Join(fullPath, "\\")
			name := fullPath[len(fullPath)-1]

			useText := "use "
			if useType == "function" {
				useText = "use function "
			} else if useType == "const" {
				useText = "use const "
			}
			useText += path + ";"

			imports = append(imports, PHPImportItem{
				name:     name,
				fullPath: path,
				line:     line,
				text:     useText,
			})
			i = useIdx
			continue
		}

		i++
	}

	return imports
}

type PHPDefinition struct {
	name      string
	defType   string
	line      int
	modifiers []string
}

func FindPHPDefinitions(content string) []PHPDefinition {
	t := NewPHPTokenizer(content)
	tokens := t.Tokenize()

	var defs []PHPDefinition

	for i := 0; i < len(tokens); i++ {
		if tokens[i].Type == PHPTokenFunction && i+1 < len(tokens) && tokens[i+1].Type == PHPTokenIdentifier {
			defs = append(defs, PHPDefinition{
				name:    tokens[i+1].Value,
				defType: "function",
				line:    tokens[i].Line,
			})
		}
		if tokens[i].Type == PHPTokenClass && i+1 < len(tokens) && tokens[i+1].Type == PHPTokenIdentifier {
			defs = append(defs, PHPDefinition{
				name:    tokens[i+1].Value,
				defType: "class",
				line:    tokens[i].Line,
			})
		}
		if tokens[i].Type == PHPTokenInterface && i+1 < len(tokens) && tokens[i+1].Type == PHPTokenIdentifier {
			defs = append(defs, PHPDefinition{
				name:    tokens[i+1].Value,
				defType: "interface",
				line:    tokens[i].Line,
			})
		}
		if tokens[i].Type == PHPTokenTrait && i+1 < len(tokens) && tokens[i+1].Type == PHPTokenIdentifier {
			defs = append(defs, PHPDefinition{
				name:    tokens[i+1].Value,
				defType: "trait",
				line:    tokens[i].Line,
			})
		}
		if tokens[i].Value == "const" && i+1 < len(tokens) && tokens[i+1].Type == PHPTokenIdentifier {
			defs = append(defs, PHPDefinition{
				name:    tokens[i+1].Value,
				defType: "const",
				line:    tokens[i].Line,
			})
		}
	}

	return defs
}

func FindPHPParameters(content, filename string) []CodeIssue {
	t := NewPHPTokenizer(content)
	tokens := t.Tokenize()

	var params []CodeIssue
	for i := 0; i < len(tokens); i++ {
		if tokens[i].Type == PHPTokenFunction && i+1 < len(tokens) {
			parenCount := 0
			paramStart := -1
			for j := i + 2; j < len(tokens); j++ {
				if tokens[j].Type == PHPTokenLParen {
					if parenCount == 0 {
						paramStart = j + 1
					}
					parenCount++
				}
				if tokens[j].Type == PHPTokenRParen {
					parenCount--
					if parenCount == 0 {
						break
					}
				}
			}

			if paramStart > 0 && paramStart < len(tokens) {
				for j := paramStart; j < len(tokens) && tokens[j].Type != PHPTokenRParen; j++ {
					if tokens[j].Type == PHPTokenIdentifier && j+1 < len(tokens) && tokens[j+1].Type != PHPTokenIdentifier {
						paramName := tokens[j].Value
						if paramName != "" && paramName != "string" && paramName != "int" && paramName != "bool" &&
							paramName != "float" && paramName != "array" && paramName != "void" && paramName != "mixed" &&
							paramName != "null" && paramName != "true" && paramName != "false" {
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

func FindUsedPHPNames(content string) map[string]int {
	t := NewPHPTokenizer(content)
	tokens := t.Tokenize()

	counts := make(map[string]int)

	for _, tok := range tokens {
		if tok.Type == PHPTokenIdentifier {
			counts[tok.Value]++
		}
	}

	return counts
}

func analyzePHP(content, filename string) AnalysisResult {
	imports := FindPHPImports(content)
	counts := FindUsedPHPNames(content)

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

func analyzePHPForWorkspace(content, filename string) ([]Definition, []Import, []CodeIssue, []CodeIssue) {
	imports := FindPHPImports(content)
	defs := FindPHPDefinitions(content)
	counts := FindUsedPHPNames(content)

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

func buildResultPHP(file AnalyzeFile, defs []Definition, imports []Import, usedNames map[string]bool, allFiles []AnalyzeFile) AnalysisResult {
	localImports := FindPHPImports(file.Content)
	localDefs := FindPHPDefinitions(file.Content)
	parameters := FindPHPParameters(file.Content, file.Filename)
	counts := FindUsedPHPNames(file.Content)

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
		isLocallyUsed := counts[d.name] > 1
		if !usedNames[key] && !isLocallyUsed {
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
		isLocallyUsed := counts[paramName] > 1
		if usedNames[key] || isLocallyUsed {
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
