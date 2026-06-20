package routes

import (
	"context"
	"encoding/json"

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
	verifyClient := &rabbitmqVerifyClient{app: app}

	protected := e.Group("/api/v1")
	protected.Use(middleware.IPRateLimiter(app), middleware.AuthMiddleware(app, verifyClient), middleware.TenantRateLimiter(app))
	{
		// A. Mobile Money Collections (Asynchronous, responds 202 Accepted)
		protected.POST("/mobile-money/collect", merchantapihandlers.HandleCollectionHandler)

		// B. Mobile Money Disbursements (Synchronous, X-Auth-Signature validated)
		protected.POST("/mobile-money/disburse", merchantapihandlers.HandleDisbursementHandler)

	}
}

type rabbitmqVerifyClient struct {
	app *global.App
}

func (v *rabbitmqVerifyClient) VerifyTokenAndIP(ctx context.Context, req *merchantapis.VerifyRequest) (*merchantapis.VerifyResponse, error) {
	respBytes, err := v.app.MQ.Request("auth.verify_token_ip", req)
	if err != nil {
		return nil, err
	}
	var resp merchantapis.VerifyResponse
	if err := json.Unmarshal(respBytes, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}
