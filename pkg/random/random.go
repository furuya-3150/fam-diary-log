package random

import (
	"crypto/rand"
	"encoding/base64"
)

// GenerateRandomBase64String generates a cryptographically secure random string
// encoded in Base64 URL format with the specified number of bytes of randomness
func GenerateRandomBase64String(length int) (string, error) {
	b := make([]byte, length)
	_, err := rand.Read(b)
	if err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(b), nil
}
