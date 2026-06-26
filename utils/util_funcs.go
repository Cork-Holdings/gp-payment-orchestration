package utils

import (
	"context"
	"crypto/aes"
	"crypto/cipher"
	cryptorand "crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"math/rand"
	"strings"
	"time"
)

func CapitalizeError(msg string) error {
	return errors.New(strings.ToUpper(msg[:1]) + msg[1:])
}

// GetNetworkProvider returns the network provider based on the phone number prefix
func GetNetworkProvider(phonenumber string) (string, error) {

	if strings.HasPrefix(phonenumber, "+26") {
		phonenumber = strings.TrimPrefix(phonenumber, "+26")
	} else if strings.HasPrefix(phonenumber, "26") {
		phonenumber = strings.TrimPrefix(phonenumber, "26")
	}
	// Ensure the number has at least 9 digits (after removing country code)
	if len(phonenumber) < 9 {
		return "", errors.New("invalid phone number format")
	}

	// Extract the prefix (first three digits)
	prefix := phonenumber[:3]

	// Map of Zambian mobile network prefixes
	providerMap := map[string]string{
		"096": "mtn",
		"076": "mtn",
		"056": "mtn",
		"097": "airtel",
		"077": "airtel",
		"057": "airtel",
		"095": "zamtel",
		"075": "zamtel",
		"055": "zamtel",
	}

	// Get the provider based on the prefix
	if provider, exists := providerMap[prefix]; exists {
		return provider, nil
	}

	return "", errors.New("unknown network provider")
}

func GenerateTenDigitCode() string {
	rand.Seed(time.Now().UnixNano())
	code := rand.Intn(9000000000) + 100000000
	return fmt.Sprintf("%d", code)
}

func GenerateSixDigitCode() string {
	rand.Seed(time.Now().UnixNano())
	code := rand.Intn(900000) + 100000
	return fmt.Sprintf("%d", code)
}

func GetIPAddress(ctx context.Context) string {
	ip, ok := ctx.Value("ip").(string)

	if !ok {
		return "unkwown"
	}

	return ip

}

// Encrypt encrypts plaintext using AES-GCM.
// key must be 16, 24, or 32 bytes long.
func Encrypt(plaintext string, key []byte) (string, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(cryptorand.Reader, nonce); err != nil {
		return "", err
	}

	ciphertext := gcm.Seal(nonce, nonce, []byte(plaintext), nil)

	return base64.StdEncoding.EncodeToString(ciphertext), nil
}

// Decrypt decrypts an AES-GCM encrypted string.
func Decrypt(encodedCiphertext string, key []byte) (string, error) {
	data, err := base64.StdEncoding.DecodeString(encodedCiphertext)
	if err != nil {
		return "", err
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	nonceSize := gcm.NonceSize()
	if len(data) < nonceSize {
		return "", fmt.Errorf("ciphertext too short")
	}

	nonce, ciphertext := data[:nonceSize], data[nonceSize:]

	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return "", err
	}

	return string(plaintext), nil
}
