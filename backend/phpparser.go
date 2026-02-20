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

	var imports []PHPImportItem

	var i int
	for i < len(tokens) {
		if tokens[i].Type == PHPTokenUse && i+1 < len(tokens) && tokens[i+1].Type == PHPTokenIdentifier {
			line := tokens[i].Line
			var fullPath strings.Builder
			fullPath.WriteString(tokens[i+1].Value)
			i += 2

			for i < len(tokens) && tokens[i].Type == PHPTokenSemi == false {
				if tokens[i].Type == PHPTokenIdentifier {
					fullPath.WriteString("\\" + tokens[i].Value)
				}
				i++
			}

			path := fullPath.String()
			name := path
			if strings.Contains(path, "\\") {
				parts := strings.Split(path, "\\")
				name = parts[len(parts)-1]
			}

			imports = append(imports, PHPImportItem{
				name:     name,
				fullPath: path,
				line:     line,
				text:     "use " + path + ";",
			})
		}

		i++
	}

	return imports
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
	counts := FindUsedPHPNames(content)

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

func buildResultPHP(file AnalyzeFile, defs []Definition, imports []Import, usedNames map[string]bool, allFiles []AnalyzeFile) AnalysisResult {
	localImports := FindPHPImports(file.Content)

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
