package utils

import (
	"crypto/rand"
	"encoding/base64"
)

func GenerateRandomKey(length int) (string, error) {
	// Generate random bytes
	b := make([]byte, length)
	_, err := rand.Read(b)
	// Note that err == nil only if we read len(b) bytes.
	if err != nil {
		return "", err
	}

	// Encode to base64
	return base64.URLEncoding.EncodeToString(b), nil
}
