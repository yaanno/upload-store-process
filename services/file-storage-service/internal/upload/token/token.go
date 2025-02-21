package token

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"time"
)

var hmacSecretKey = []byte("your-secret-hmac-key") // Replace with your actual secret key

// generateSecureFileID creates a cryptographically secure, unique file identifier
func GenerateSecureFileID() (string, error) {
	// Generate 32 bytes of cryptographically secure random data
	randomBytes := make([]byte, 32)
	_, err := rand.Read(randomBytes)
	if err != nil {
		return "", fmt.Errorf("failed to generate secure random bytes: %w", err)
	}

	// Create a hash to add an extra layer of unpredictability
	hash := sha256.Sum256(append(randomBytes, []byte(time.Now().String())...))

	// Use URL-safe base64 encoding to ensure safe use in URLs and file systems
	return base64.URLEncoding.EncodeToString(hash[:]), nil
}

// generateSecureUploadToken creates a time-limited, secure upload token
func GenerateSecureUploadToken(fileID string) (string, error) {
	expirationTimestamp := time.Now().Add(time.Hour).Unix()

	message := fmt.Sprintf("%s_%d", fileID, expirationTimestamp)

	hmacHasher := hmac.New(sha256.New, hmacSecretKey)
	hmacHasher.Write([]byte(message))
	hmacBytes := hmacHasher.Sum(nil)

	hmacBase64 := base64.URLEncoding.EncodeToString(hmacBytes)
	token := fmt.Sprintf("%d_%s", expirationTimestamp, hmacBase64)
	return token, nil
}
