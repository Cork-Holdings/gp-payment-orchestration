package routes

import (
	"context"

	merchantapihandlers "github.com/Cork-Holdings/gp_payment_orchestration/internal/api/handlers/merchant_api_handlers"
	"github.com/Cork-Holdings/gp_payment_orchestration/internal/modules/merchantapis"

	"github.com/Cork-Holdings/gp_payment_orchestration/internal/global"
	"github.com/Cork-Holdings/gp_payment_orchestration/internal/middleware"
	"github.com/gin-gonic/gin"
)

func RegisterMerchantRoutes(e *gin.Engine, app *global.App) {
	// 1. OAuth Token Generation (Public, urlencoded, rate limited by IP)
	e.POST("/oauth/token", middleware.IPRateLimiter(app), merchantapihandlers.HandleGenerateTokenHandler)

	// 2. Protected Merchant Endpoints
	verifyClient := &directVerifyClient{app: app}

	protected := e.Group("/api/v1")
	protected.Use(middleware.IPRateLimiter(app), middleware.AuthMiddleware(app, verifyClient), middleware.TenantRateLimiter(app))
	{
		// A. Mobile Money Collections (Asynchronous, responds 202 Accepted)
		protected.POST("/mobile-money/collect", merchantapihandlers.HandleCollectionHandler)

		// B. Mobile Money Disbursements (Synchronous, X-Auth-Signature validated)
		protected.POST("/mobile-money/disburse", merchantapihandlers.HandleDisbursementHandler)

	}
}

// directVerifyClient verifies tokens directly without going through RabbitMQ
type directVerifyClient struct {
	app *global.App
}

func (v *directVerifyClient) VerifyTokenAndIP(ctx context.Context, req *merchantapis.VerifyRequest) (*merchantapis.VerifyResponse, error) {
	// Direct call to the verification function - no RabbitMQ needed
	result := merchantapis.VerifyTokenAndIPDirect(v.app, req)
	return result, nil
}
