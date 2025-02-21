package validation

import (
	"crypto/hmac"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
	"fmt"
	"strconv"
	"strings"
	"time"
)

var hmacSecretKey = []byte("your-secret-hmac-key") // Replace with your actual secret key

func ValidateSecureUploadToken(token string, fileID string) error {
	parts := strings.SplitN(token, "_", 2)
	if len(parts) != 2 {
		return fmt.Errorf("invalid token format: missing timestamp or hash")
	}

	expirationTimestampStr := parts[0]
	hashBase64 := parts[1]

	expirationTimestampUnix, err := strconv.ParseInt(expirationTimestampStr, 10, 64)
	if err != nil {
		return fmt.Errorf("invalid token format: invalid timestamp: %w", err)
	}
	expirationTime := time.Unix(expirationTimestampUnix, 0)

	if time.Now().After(expirationTime) {
		return fmt.Errorf("upload token expired")
	}

	decodedHmacBytes, err := base64.URLEncoding.DecodeString(hashBase64)
	if err != nil {
		return fmt.Errorf("invalid token format: invalid base64 hash: %w", err)
	}
	if len(decodedHmacBytes) != sha256.Size {
		return fmt.Errorf("invalid token format: hash has incorrect length")
	}
	var decodedHmacSignature [sha256.Size]byte
	copy(decodedHmacSignature[:], decodedHmacBytes)

	message := fmt.Sprintf("%s_%d", fileID, expirationTimestampUnix)

	hmacHasher := hmac.New(sha256.New, hmacSecretKey)
	hmacHasher.Write([]byte(message))
	recomputedHmacSlice := hmacHasher.Sum(nil) // recomputedHmacSlice is []byte

	// **Convert recomputedHmacSlice (slice) to recomputedHmacArray ([32]byte array):**
	var recomputedHmacArray [sha256.Size]byte
	copy(recomputedHmacArray[:], recomputedHmacSlice)

	// Compare the re-hashed token data with the decoded hash from the token
	if !CompareHashes(recomputedHmacArray, decodedHmacSignature) { // Use recomputedHmacArray now
		return fmt.Errorf("upload token HMAC signature mismatch: token is invalid or tampered with")
	}

	return nil // Token is valid
}

// Helper function to securely compare hashes to prevent timing attacks
func CompareHashes(hash1 [sha256.Size]byte, hash2 [sha256.Size]byte) bool {
	return subtle.ConstantTimeCompare(hash1[:], hash2[:]) == 1
}

// Optional: Token validation function
func IsUploadTokenValid(token string, fileID string) bool {
	if token == "" {
		return false
	}

	if fileID == "" {
		return false
	}

	parts := strings.Split(token, "_")
	if len(parts) != 2 {
		return false
	}

	// Check token expiration
	expirationTime, err := strconv.ParseInt(parts[0], 10, 64)
	if err != nil {
		return false
	}

	if time.Now().Unix() > expirationTime {
		return false // Token expired
	}

	// Optionally, you could add additional validation logic here
	return true
}
