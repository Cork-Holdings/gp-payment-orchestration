package middleware

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/Cork-Holdings/gp_payment_orchestration/internal/global"
	"github.com/Cork-Holdings/gp_payment_orchestration/internal/modules/merchantapis"
	"github.com/gin-gonic/gin"
)

type TokenVerifier interface {
	VerifyTokenAndIP(ctx context.Context, req *merchantapis.VerifyRequest) (*merchantapis.VerifyResponse, error)
}

// AuthMiddleware validates OAuth JWT token, Client ID, and whitelisted IP via TokenVerifier
func AuthMiddleware(app *global.App, verifier TokenVerifier) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		clientID := c.GetHeader("X-Client-ID")

		if authHeader == "" || clientID == "" {
			// Try to get client_id from query or body as fallback for debugging
			if clientID == "" {
				clientID = c.Query("client_id")
			}
			if clientID == "" {
				clientID = c.PostForm("client_id")
			}
		}

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

		// Extract IP from X-Forwarded-For if behind a proxy, fallback to client IP
		clientIP := c.GetHeader("X-Forwarded-For")
		if clientIP == "" {
			clientIP = c.ClientIP()
		} else {
			ips := strings.Split(clientIP, ",")
			clientIP = strings.TrimSpace(ips[0])
		}

		//test webhook

		res, err := verifier.VerifyTokenAndIP(c.Request.Context(), &merchantapis.VerifyRequest{
			Token:     token,
			ClientID:  clientID,
			IPAddress: clientIP,
		})

		if err != nil {
			fmt.Printf("[Middleware] Auth error: %v\n", err)
			c.JSON(http.StatusUnauthorized, gin.H{"error": "auth service validation failed: " + err.Error()})
			c.Abort()
			return
		}

		if !res.Valid {
			fmt.Printf("[Middleware] Auth failed: %s (ClientID: %s, IP: %s)\n", res.ErrorMessage, clientID, clientIP)
			c.JSON(http.StatusUnauthorized, gin.H{"error": res.ErrorMessage})
			c.Abort()
			return
		}

		c.Set("client_id", clientID)
		c.Set("tenant_id", res.TenantID)
		c.Set("merchant_id", res.MerchantID)
		c.Next()
	}
}

// IPRateLimiter limits requests per IP using Redis (capped at 100 req/min)
func IPRateLimiter(app *global.App) gin.HandlerFunc {
	return func(c *gin.Context) {
		if app.Cache == nil {
			c.Next()
			return
		}
		ip := c.ClientIP()
		ipKey := "rate:ip:" + ip

		val, err := app.Cache.Incr(c.Request.Context(), ipKey).Result()
		if err == nil {
			if val == 1 {
				app.Cache.Expire(c.Request.Context(), ipKey, time.Minute)
			}
			if val > 100 {
				c.JSON(http.StatusTooManyRequests, gin.H{"error": "IP rate limit exceeded"})
				c.Abort()
				return
			}
		}
		c.Next()
	}
}

// TenantRateLimiter limits requests per Merchant Client ID using Redis (capped at 200 req/min)
func TenantRateLimiter(app *global.App) gin.HandlerFunc {
	return func(c *gin.Context) {
		if app.Cache == nil {
			c.Next()
			return
		}
		clientID := c.GetString("client_id")
		if clientID == "" {
			c.Next()
			return
		}

		tenantKey := "rate:tenant:" + clientID
		val, err := app.Cache.Incr(c.Request.Context(), tenantKey).Result()
		if err == nil {
			if val == 1 {
				app.Cache.Expire(c.Request.Context(), tenantKey, time.Minute)
			}
			if val > 200 {
				c.JSON(http.StatusTooManyRequests, gin.H{"error": "Merchant rate limit exceeded"})
				c.Abort()
				return
			}
		}
		c.Next()
	}
}
