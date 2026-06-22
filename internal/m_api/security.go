package m_api

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/hmac"
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"io"
	"log"
	"os"
	"strings"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
)

const (
	secretEncPrefix  = "enc:v1:"
	minJWTSecretSize = 32
)

var jwtSecret = loadJWTSecret()

func loadJWTSecret() []byte {
	secret := strings.TrimSpace(os.Getenv("MERCHANT_JWT_SECRET"))
	if secret == "" {
		if strings.EqualFold(os.Getenv("APP_ENV"), "production") {
			log.Fatal("MERCHANT_JWT_SECRET must be set when APP_ENV=production")
		}
		log.Println("WARNING: MERCHANT_JWT_SECRET not set; using insecure development default")
		return []byte("merchant_secret_jwt_key_999_dev_only")
	}
	if len(secret) < minJWTSecretSize {
		log.Fatalf("MERCHANT_JWT_SECRET must be at least %d characters", minJWTSecretSize)
	}
	return []byte(secret)
}

func encryptionKey() []byte {
	raw := strings.TrimSpace(os.Getenv("MERCHANT_SECRET_ENCRYPTION_KEY"))
	if raw == "" {
		return nil
	}
	key, err := hex.DecodeString(raw)
	if err != nil || len(key) != 32 {
		log.Fatal("MERCHANT_SECRET_ENCRYPTION_KEY must be a 64-character hex-encoded 32-byte key")
	}
	return key
}

// ClientIP returns the client IP using Gin's trusted-proxy configuration.
func ClientIP(c *gin.Context) string {
	return c.ClientIP()
}

func protectSecret(plain string) (string, error) {
	key := encryptionKey()
	if key == nil {
		if strings.EqualFold(os.Getenv("APP_ENV"), "production") {
			return "", errors.New("MERCHANT_SECRET_ENCRYPTION_KEY must be set when APP_ENV=production")
		}
		return plain, nil
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", err
	}

	ciphertext := gcm.Seal(nonce, nonce, []byte(plain), nil)
	return secretEncPrefix + base64.StdEncoding.EncodeToString(ciphertext), nil
}

func unprotectSecret(stored string) (string, error) {
	if strings.HasPrefix(stored, secretEncPrefix) {
		key := encryptionKey()
		if key == nil {
			return "", errors.New("encrypted secret found but MERCHANT_SECRET_ENCRYPTION_KEY is not configured")
		}

		raw, err := base64.StdEncoding.DecodeString(strings.TrimPrefix(stored, secretEncPrefix))
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
		if len(raw) < nonceSize {
			return "", errors.New("invalid encrypted secret payload")
		}

		plain, err := gcm.Open(nil, raw[:nonceSize], raw[nonceSize:], nil)
		if err != nil {
			return "", err
		}
		return string(plain), nil
	}

	return stored, nil
}

func verifyClientSecret(stored, provided string) bool {
	if strings.HasPrefix(stored, secretEncPrefix) {
		plain, err := unprotectSecret(stored)
		if err != nil {
			return false
		}
		return subtle.ConstantTimeCompare([]byte(plain), []byte(provided)) == 1
	}

	if isBcryptHash(stored) {
		return bcrypt.CompareHashAndPassword([]byte(stored), []byte(provided)) == nil
	}

	return subtle.ConstantTimeCompare([]byte(stored), []byte(provided)) == 1
}

func secretForHMAC(stored string) (string, error) {
	if isBcryptHash(stored) {
		return "", errors.New("client secret is hashed; HMAC signing requires MERCHANT_SECRET_ENCRYPTION_KEY")
	}
	return unprotectSecret(stored)
}

func isBcryptHash(value string) bool {
	return strings.HasPrefix(value, "$2a$") ||
		strings.HasPrefix(value, "$2b$") ||
		strings.HasPrefix(value, "$2y$")
}

func verifyHMAC(message, secret, signature string) bool {
	expectedMAC, err := hex.DecodeString(computeHMAC(message, secret))
	if err != nil {
		return false
	}

	providedMAC, err := hex.DecodeString(strings.TrimSpace(signature))
	if err != nil {
		return false
	}

	return hmac.Equal(expectedMAC, providedMAC)
}
