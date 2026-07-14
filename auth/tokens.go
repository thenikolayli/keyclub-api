package auth

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
)

// Generates a random 32 byte token
// for use in authentication
func MustGenerateToken() string {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		panic(err)
	}
	return base64.URLEncoding.WithPadding(base64.NoPadding).EncodeToString(b)
}

// Hashes a token using SHA-256
// for use in authentication
func MustHashToken(token string) string {
	hash := sha256.Sum256([]byte(token))
	return base64.URLEncoding.WithPadding(base64.NoPadding).EncodeToString(hash[:])
}
