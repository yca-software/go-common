package token

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
)

const (
	TokenLength = 32
)

// GenerateToken generates a secure random token using cryptographically secure random number generation.
// Returns a base64 URL-encoded string representing 32 random bytes.
func GenerateToken() (string, error) {
	tokenBytes := make([]byte, TokenLength)
	if _, err := rand.Read(tokenBytes); err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(tokenBytes), nil
}

// HashToken hashes a token using SHA256 and returns the hex-encoded hash.
// This is useful for storing tokens securely in databases.
func HashToken(token string) string {
	hash := sha256.Sum256([]byte(token))
	return hex.EncodeToString(hash[:])
}
