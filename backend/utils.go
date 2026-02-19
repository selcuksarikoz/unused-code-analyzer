package main

import (
	"crypto/rand"
	"encoding/hex"
)

func generateUUID() string {
	b := make([]byte, 16)
	rand.Read(b)
	return hex.EncodeToString(b)
}
