package m_api

import (
	"context"
	"net/http"
	"strings"
	"time"

	"github.com/Cork-Holdings/gp_payment_orchestration/internal/global"
	"github.com/gin-gonic/gin"
)

type TokenVerifier interface {
	VerifyTokenAndIP(ctx context.Context, req *VerifyRequest) (*VerifyResponse, error)
}

// AuthMiddleware validates OAuth JWT token, Client ID, and whitelisted IP via TokenVerifier
func AuthMiddleware(app *global.App, verifier TokenVerifier) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		clientID := c.GetHeader("X-Client-ID")

		if authHeader == "" || clientID == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "missing Authorization or X-Client-ID header"})
			c.Abort()
			return
		}

		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid Authorization header format"})
			c.Abort()
			return
		}
		token := parts[1]

		clientIP := ClientIP(c)

		res, err := verifier.VerifyTokenAndIP(c.Request.Context(), &VerifyRequest{
			Token:     token,
			ClientID:  clientID,
			IPAddress: clientIP,
		})

		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "authentication validation failed"})
			c.Abort()
			return
		}

		if !res.Valid {
			c.JSON(http.StatusUnauthorized, gin.H{"error": res.ErrorMessage})
			c.Abort()
			return
		}

		c.Set("client_id", clientID)
		c.Set("tenant_id", res.TenantID)
		c.Next()
	}
}

// IPRateLimiter limits requests per IP using Redis (capped at 100 req/min)
func IPRateLimiter(app *global.App) gin.HandlerFunc {
	return func(c *gin.Context) {
		if app.Cache == nil {
			c.JSON(http.StatusServiceUnavailable, gin.H{"error": "rate limiting unavailable"})
			c.Abort()
			return
		}
		ip := ClientIP(c)
		ipKey := "rate:ip:" + ip

		val, err := app.Cache.Incr(c.Request.Context(), ipKey).Result()
		if err != nil {
			c.JSON(http.StatusServiceUnavailable, gin.H{"error": "rate limiting unavailable"})
			c.Abort()
			return
		}
		if val == 1 {
			app.Cache.Expire(c.Request.Context(), ipKey, time.Minute)
		}
		if val > 100 {
			c.JSON(http.StatusTooManyRequests, gin.H{"error": "IP rate limit exceeded"})
			c.Abort()
			return
		}
		c.Next()
	}
}

// TenantRateLimiter limits requests per Merchant Client ID using Redis (capped at 200 req/min)
func TenantRateLimiter(app *global.App) gin.HandlerFunc {
	return func(c *gin.Context) {
		if app.Cache == nil {
			c.JSON(http.StatusServiceUnavailable, gin.H{"error": "rate limiting unavailable"})
			c.Abort()
			return
		}
		clientID := c.GetString("client_id")
		if clientID == "" {
			c.Next()
			return
		}

		tenantKey := "rate:tenant:" + clientID
		val, err := app.Cache.Incr(c.Request.Context(), tenantKey).Result()
		if err != nil {
			c.JSON(http.StatusServiceUnavailable, gin.H{"error": "rate limiting unavailable"})
			c.Abort()
			return
		}
		if val == 1 {
			app.Cache.Expire(c.Request.Context(), tenantKey, time.Minute)
		}
		if val > 200 {
			c.JSON(http.StatusTooManyRequests, gin.H{"error": "Merchant rate limit exceeded"})
			c.Abort()
			return
		}
		c.Next()
	}
}
