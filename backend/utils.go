package main

import (
	"crypto/rand"
	"encoding/hex"
	"strings"
	"unicode"
)

func generateUUID() string {
	b := make([]byte, 16)
	rand.Read(b)
	return hex.EncodeToString(b)
}

type NamedItem struct {
	Name string
	Line int
}

// isWordBoundary checks if the character at position pos is a word boundary
func isWordBoundary(s string, pos int) bool {
	if pos <= 0 || pos >= len(s) {
		return true
	}
	return !unicode.IsLetter(rune(s[pos-1])) && !unicode.IsDigit(rune(s[pos-1])) && s[pos-1] != '_'
}

// isWordEnd checks if the position pos marks the end of a word
func isWordEnd(s string, pos int) bool {
	if pos >= len(s) {
		return true
	}
	return !unicode.IsLetter(rune(s[pos])) && !unicode.IsDigit(rune(s[pos])) && s[pos] != '_'
}

func FindUsedNames(content string, items []NamedItem) map[string]bool {
	used := make(map[string]bool)
	lines := strings.Split(content, "\n")

	for _, item := range items {
		found := false
		for i, line := range lines {
			if i+1 == item.Line {
				continue
			}

			// Search for the name as a whole word
			idx := 0
			for {
				pos := strings.Index(line[idx:], item.Name)
				if pos == -1 {
					break
				}
				pos += idx

				// Check if it's a whole word match
				if isWordBoundary(line, pos) && isWordEnd(line, pos+len(item.Name)) {
					found = true
					break
				}
				idx = pos + 1
			}

			if found {
				break
			}
		}
		if found {
			used[item.Name] = true
		}
	}

	return used
}

// containsWord checks if a word exists as a whole word in the content
func containsWord(content, word string) bool {
	idx := 0
	for {
		pos := strings.Index(content[idx:], word)
		if pos == -1 {
			break
		}
		pos += idx
		if isWordBoundary(content, pos) && isWordEnd(content, pos+len(word)) {
			return true
		}
		idx = pos + 1
	}
	return false
}

func findUsedParameterNames(content string, params []CodeIssue) map[string]bool {
	var items []NamedItem
	for _, p := range params {
		name := strings.TrimSpace(strings.TrimPrefix(p.Text, "parameter "))
		if name == "" {
			continue
		}
		items = append(items, NamedItem{Name: name, Line: p.Line})
	}
	return FindUsedNames(content, items)
}
