package middleware

import (
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/Cork-Holdings/gp_payment_orchestration/internal/common"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

// SessionAuthMiddleware validates the platform's shared ECDSA-signed session
// token — the same one gp_auth issues and gp_gateway validates before
// proxying here. It protects the merchant/admin-facing config routes
// (subscriptions, merchant-subscriptions, merchant-api-keys, etc.) that
// previously had no auth at all.
//
// This only authenticates the caller (rejects missing/invalid/expired
// tokens); it does not yet derive merchant identity from the claims for
// authorization — handlers still take merchant_id as a request parameter.
func SessionAuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"message": "missing authorization header",
				"status":  "failed",
			})
			return
		}

		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"message": "malformed authorization header",
				"status":  "failed",
			})
			return
		}

		claims := &common.Claims{}
		token, err := jwt.ParseWithClaims(parts[1], claims, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodECDSA); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
			}
			block, _ := pem.Decode([]byte(normalizePEM(os.Getenv("PUBLIC_KEY"))))
			if block == nil {
				return nil, fmt.Errorf("failed to decode PEM public key")
			}
			pub, err := x509.ParsePKIXPublicKey(block.Bytes)
			if err != nil {
				return nil, fmt.Errorf("failed to parse public key: %w", err)
			}
			return pub, nil
		})

		if err != nil || !token.Valid {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"message": "invalid or expired token",
				"status":  "failed",
			})
			return
		}

		c.Set("user", claims)
		c.Next()
	}
}

// normalizePEM ensures a PEM key has proper line breaks even when stored as a
// single-line string (common when set via environment variables). Mirrors
// gp_gateway's internal/api/middleware/auth.go so the same PUBLIC_KEY value
// works unmodified in both services.
func normalizePEM(key string) string {
	const header = "-----BEGIN PUBLIC KEY-----"
	const footer = "-----END PUBLIC KEY-----"
	key = strings.TrimSpace(key)
	if strings.Contains(key, "\n") {
		return key
	}
	body := key
	body = strings.TrimPrefix(body, header)
	body = strings.TrimSuffix(body, footer)
	body = strings.TrimSpace(body)
	var lines []string
	for len(body) > 64 {
		lines = append(lines, body[:64])
		body = body[64:]
	}
	if len(body) > 0 {
		lines = append(lines, body)
	}
	return header + "\n" + strings.Join(lines, "\n") + "\n" + footer
}
