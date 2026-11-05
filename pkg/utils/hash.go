package utils

import (
	"crypto/sha256"
	"encoding/hex"
)

// HashToken computes a SHA-256 checksum of a JWT token string
func HashToken(token string) string {
	hash := sha256.Sum256([]byte(token))
	return hex.EncodeToString(hash[:])
}
