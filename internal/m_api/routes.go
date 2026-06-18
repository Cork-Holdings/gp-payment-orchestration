package m_api

import (
	"context"
	"encoding/json"
	"net/http"
	"regexp"

	"github.com/Cork-Holdings/gp_payment_orchestration/internal/global"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

var phoneValidationRegex = regexp.MustCompile(`^\d{12}$`)

func RegisterMerchantRoutes(e *gin.Engine, app *global.App) {
	// 1. OAuth Token Generation (Public, urlencoded, rate limited by IP)
	e.POST("/oauth/token", IPRateLimiter(app), func(c *gin.Context) {
		clientID := c.PostForm("client_id")
		clientSecret := c.PostForm("client_secret")

		if clientID == "" || clientSecret == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "missing client_id or client_secret"})
			return
		}

		reqPayload := map[string]string{
			"client_id":     clientID,
			"client_secret": clientSecret,
		}

		respBytes, err := app.MQ.Request("auth.generate_token", reqPayload)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "authentication request failed: " + err.Error()})
			return
		}

		var resp map[string]interface{}
		if err := json.Unmarshal(respBytes, &resp); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to parse auth response"})
			return
		}

		if errMsg, ok := resp["error"].(string); ok && errMsg != "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": errMsg})
			return
		}

		c.JSON(http.StatusOK, resp)
	})

	// 2. Protected Merchant Endpoints
	verifyClient := &rabbitmqVerifyClient{app: app}

	protected := e.Group("/api/v1")
	protected.Use(IPRateLimiter(app), AuthMiddleware(app, verifyClient), TenantRateLimiter(app))
	{
		// A. Mobile Money Collections (Asynchronous, responds 202 Accepted)
		protected.POST("/mobile-money/collect", func(c *gin.Context) {
			var req struct {
				PhoneNumber string  `json:"phone_number" binding:"required"`
				Amount      float64 `json:"amount" binding:"required,gt=0"`
				Currency    string  `json:"currency" binding:"required"`
			}

			if err := c.ShouldBindJSON(&req); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
				return
			}

			// Validate phone_number is exactly 12 digits
			if !phoneValidationRegex.MatchString(req.PhoneNumber) {
				c.JSON(http.StatusBadRequest, gin.H{"error": "phone_number must be exactly 12 digits long"})
				return
			}

			clientID := c.GetString("client_id")
			trackingRef := "col_" + uuid.New().String()

			// Emit async event
			eventPayload := map[string]interface{}{
				"client_id":    clientID,
				"phone_number": req.PhoneNumber,
				"amount":       req.Amount,
				"currency":     req.Currency,
				"tracking_ref": trackingRef,
			}

			err := app.MQ.Emit("collection.collect", eventPayload)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to initiate collection: " + err.Error()})
				return
			}

			c.JSON(http.StatusAccepted, gin.H{
				"tracking_ref": trackingRef,
				"status":       "PENDING",
			})
		})

		// B. Mobile Money Disbursements (Synchronous, X-Auth-Signature validated)
		protected.POST("/mobile-money/disburse", func(c *gin.Context) {
			signature := c.GetHeader("X-Auth-Signature")
			if signature == "" {
				c.JSON(http.StatusBadRequest, gin.H{"error": "missing X-Auth-Signature header"})
				return
			}

			rawBody, err := c.GetRawData()
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "failed to read request body"})
				return
			}

			var req struct {
				PhoneNumber string  `json:"phone_number" binding:"required"`
				Amount      float64 `json:"amount" binding:"required,gt=0"`
				Currency    string  `json:"currency" binding:"required"`
			}

			if err := json.Unmarshal(rawBody, &req); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "invalid JSON body: " + err.Error()})
				return
			}

			clientID := c.GetString("client_id")
			trackingRef := "dis_" + uuid.New().String()

			reqPayload := map[string]interface{}{
				"client_id":    clientID,
				"phone_number": req.PhoneNumber,
				"amount":       req.Amount,
				"currency":     req.Currency,
				"signature":    signature,
				"raw_body":     string(rawBody),
				"tracking_ref": trackingRef,
			}

			respBytes, err := app.MQ.Request("disbursement.disburse", reqPayload)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "disbursement request failed: " + err.Error()})
				return
			}

			var resp map[string]interface{}
			if err := json.Unmarshal(respBytes, &resp); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to parse disbursement response"})
				return
			}

			if statusVal, ok := resp["status"].(string); ok && statusVal == "FAILED" {
				errCode, _ := resp["error_code"].(string)
				c.JSON(http.StatusBadRequest, gin.H{
					"status":     "FAILED",
					"error_code": errCode,
				})
				return
			}

			c.JSON(http.StatusOK, resp)
		})
	}
}

type rabbitmqVerifyClient struct {
	app *global.App
}

func (v *rabbitmqVerifyClient) VerifyTokenAndIP(ctx context.Context, req *VerifyRequest) (*VerifyResponse, error) {
	respBytes, err := v.app.MQ.Request("auth.verify_token_ip", req)
	if err != nil {
		return nil, err
	}
	var resp VerifyResponse
	if err := json.Unmarshal(respBytes, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}
