package main

import (
	"crypto/rand"
	"encoding/hex"
	"regexp"
	"strings"
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

func FindUsedNames(content string, items []NamedItem) map[string]bool {
	used := make(map[string]bool)
	lines := strings.Split(content, "\n")

	for _, item := range items {
		found := false
		for i, line := range lines {
			if i+1 == item.Line {
				continue
			}
			re := regexp.MustCompile(`\b` + regexp.QuoteMeta(item.Name) + `\b`)
			if re.MatchString(line) {
				found = true
				break
			}
		}
		if found {
			used[item.Name] = true
		}
	}

	return used
}
