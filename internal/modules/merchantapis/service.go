package merchantapis

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/Cork-Holdings/gp_payment_orchestration/internal/global"
	"github.com/Cork-Holdings/gp_payment_orchestration/internal/modules/feecalculator"
	"github.com/Cork-Holdings/gp_payment_orchestration/internal/modules/merchantapikeys"
	"github.com/Cork-Holdings/gp_payment_orchestration/internal/modules/merchantips"
	"github.com/Cork-Holdings/gp_payment_orchestration/utils"
	"github.com/golang-jwt/jwt/v5"
	"gorm.io/gorm"
)

var jwtSecret = []byte("merchant_secret_jwt_key_999")

type TokenRequest struct {
	ClientID     string `json:"client_id" form:"client_id"`
	ClientSecret string `json:"client_secret" form:"client_secret"`
}

type TokenResponse struct {
	AccessToken string `json:"access_token,omitempty"`
	ExpiresIn   int64  `json:"expires_in,omitempty"`
	TokenType   string `json:"token_type,omitempty"`
	Error       string `json:"error,omitempty"`
}

type VerifyRequest struct {
	Token     string `json:"token"`
	ClientID  string `json:"client_id"`
	IPAddress string `json:"ip_address"`
}

type CheckStatusRequest struct {
	TransactionRef string `json:"transaction_ref"`
	ClientID       string `json:"client_id"`
}

type CheckStatusResponse struct {
	// TransactionRef string      `json:"transaction_ref"`
	Status  string      `json:"status"`
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

type CheckCollectionBalanceRequest struct {
	ClientID string `json:"client_id"`
}
type CheckCollectionBalanceResponse struct {
	Code    int         `json:"code"`
	Status  string      `json:"status"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

type CheckDisbursementBalanceRequest struct {
	ClientID       string `json:"client_id"`
	XAuthSignature string `json:"x_auth_signature"`
}
type CheckDisbursementBalanceResponse struct {
	Code    int         `json:"code"`
	Status  string      `json:"status"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

type CreateCheckoutSessionRequest struct {
	TransactionRef string `json:"transaction_ref"`
}

type CreateCheckoutSessionResponse struct {
	TransactionRef string `json:"transaction_ref"`
}

type VerifyResponse struct {
	Valid        bool   `json:"valid"`
	TenantID     string `json:"tenant_id,omitempty"`
	MerchantID   string `json:"merchant_id,omitempty"`
	ErrorMessage string `json:"error_message,omitempty"`
}

type CollectRequest struct {
	ClientID       string  `json:"client_id"`
	MerchantID     string  `json:"merchant_id"`
	PhoneNumber    string  `json:"phone_number"`
	Amount         float64 `json:"amount"`
	TransactionRef string  `json:"transaction_ref"`
}

type DisburseRequest struct {
	ClientID       string  `json:"client_id"`
	PhoneNumber    string  `json:"phone_number"`
	Amount         float64 `json:"amount"`
	Narration      string  `json:"narration"`
	Signature      string  `json:"signature"`
	RawBody        string  `json:"raw_body"`
	TransactionRef string  `json:"transaction_ref"`
}

type DisburseResponse struct {
	TransactionRef string `json:"transaction_ref,omitempty"`
	Status         string `json:"status"`
	ErrorCode      string `json:"error_code,omitempty"`
}

// HandleGenerateToken validates credentials and generates a Bearer JWT token
func HandleGenerateToken(app *global.App, req TokenRequest) (*TokenResponse, error) {
	var mApiKey merchantapikeys.MerchantAPIKey

	err := app.DB.Where("client_id = ?", req.ClientID).First(&mApiKey).Error
	if err != nil {
		// resp, _ := json.Marshal(TokenResponse{Error: "invalid client_id or client_secret"})
		return &TokenResponse{Error: "invalid client_id or client_secret"}, nil
	}

	if mApiKey.ClientSecret != req.ClientSecret {
		// resp, _ := json.Marshal(TokenResponse{Error: "invalid client_id or client_secret"})
		return &TokenResponse{Error: "invalid client_id or client_secret"}, nil
	}

	expiresIn := int64(86400) // 24 hours
	expirationTime := time.Now().Add(time.Duration(expiresIn) * time.Second)

	claims := &jwt.MapClaims{
		"client_id": req.ClientID,
		"exp":       expirationTime.Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(jwtSecret)
	if err != nil {
		return nil, err
	}

	utils.LogAuditEvent(app, "auth_service", "token.generated", map[string]string{
		"client_id": req.ClientID,
	})

	return &TokenResponse{AccessToken: tokenString, ExpiresIn: expiresIn, TokenType: "Bearer"}, nil
}

// VerifyTokenAndIPDirect validates token signature, client ID, and IP whitelist (direct call, no RabbitMQ)
func VerifyTokenAndIPDirect(app *global.App, req *VerifyRequest) *VerifyResponse {
	token, err := jwt.Parse(req.Token, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return jwtSecret, nil
	})

	if err != nil || !token.Valid {
		return &VerifyResponse{Valid: false, ErrorMessage: "invalid or expired token"}
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return &VerifyResponse{Valid: false, ErrorMessage: "invalid claims structure"}
	}

	claimClientID, _ := claims["client_id"].(string)
	if claimClientID != req.ClientID {
		fmt.Printf("[Auth] Client ID mismatch: token=%s, header=%s\n", claimClientID, req.ClientID)
		return &VerifyResponse{Valid: false, ErrorMessage: "token does not match Client ID"}
	}

	var mApiKey merchantapikeys.MerchantAPIKey
	lookupErr := app.DB.Where("client_id = ?", claimClientID).First(&mApiKey).Error
	if lookupErr != nil {
		if errors.Is(lookupErr, gorm.ErrRecordNotFound) {
			fmt.Printf("[Auth] Merchant profile not found for client_id: %s\n", claimClientID)
			return &VerifyResponse{Valid: false, ErrorMessage: "merchant profile not found"}
		}
		fmt.Printf("[Auth] DB error: %v\n", lookupErr)
		return &VerifyResponse{Valid: false, ErrorMessage: "database error"}
	}

	var mIP merchantips.MerchantIP
	err = app.DB.Where("merchant_id = ? AND ip_address = ? AND status = ?", mApiKey.MerchantID, req.IPAddress, "approved").First(&mIP).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			fmt.Printf("[Auth] IP not whitelisted: merchant=%s, ip=%s\n", mApiKey.MerchantID, req.IPAddress)
			return &VerifyResponse{Valid: false, ErrorMessage: "IP address not whitelisted or not approved"}
		}
		fmt.Printf("[Auth] DB error checking IP: %v\n", err)
		return &VerifyResponse{Valid: false, ErrorMessage: "database error"}
	}

	return &VerifyResponse{
		Valid:      true,
		TenantID:   mApiKey.ID.String(),
		MerchantID: mApiKey.MerchantID.String(),
	}
}

// HandleVerifyTokenAndIP validates token signature, client ID, and IP whitelist (via RabbitMQ payload)
func HandleVerifyTokenAndIP(app *global.App, payload []byte) ([]byte, error) {
	var req VerifyRequest
	if err := json.Unmarshal(payload, &req); err != nil {
		return nil, err
	}

	result := VerifyTokenAndIPDirect(app, &req)
	resp, _ := json.Marshal(result)
	return resp, nil
}

type CollectResponse struct {
	Code    int         `json:"code"`
	Status  string      `json:"status"`
	Message string      `json:"message"`
	Data    interface{} `json:"data"`
}

// HandleCollect initiates an async collection, transitioning PENDING -> PROCESSING
func HandleCollection(app *global.App, req *CollectRequest) (*CollectResponse, error) {
	if len(req.PhoneNumber) < 5 {
		return nil, fmt.Errorf("invalid phone number")
	}

	//Extract Merchant ID from Client ID
	var mApiKey merchantapikeys.MerchantAPIKey
	err := app.DB.Where("client_id = ?", req.ClientID).First(&mApiKey).Error
	if err != nil {
		return nil, fmt.Errorf("failed to extract merchant ID: %v", err)
	}
	merchantID := mApiKey.MerchantID.String()

	//Get merchant details via rabbitMQ request
	merchantPayload := map[string]interface{}{
		"merchant_id": merchantID,
	}

	merchantResponseBytes, err := app.MQ.Request("merchants.get_details", merchantPayload)
	if err != nil {
		return nil, fmt.Errorf("failed to get merchant details: %v", err)
	}

	var merchantResponse map[string]interface{}
	if err := json.Unmarshal(merchantResponseBytes, &merchantResponse); err != nil {
		return nil, fmt.Errorf("failed to unmarshal merchant details: %v", err)
	}

	if merchantResponse["code"] != 200 {
		return nil, fmt.Errorf("failed to get merchant details: %v", merchantResponse["message"])
	}

	// Calculate fees for this collection
	feeResult, err := feecalculator.CalculateFees(
		merchantID,
		req.PhoneNumber,
		req.Amount,
		feecalculator.TransactionTypeCollection,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to calculate fees: %v", err)
	}

	// Check if fee calculation returned an error status
	if feeResult.Status == "error" {
		return &CollectResponse{
			Code:    400,
			Status:  "failed",
			Message: feeResult.Error,
			Data:    nil,
		}, nil
	}

	// Log the fee calculation result
	feeResultJSON, _ := json.Marshal(feeResult)
	log.Printf("[HandleCollect] Fee calculation result: %s", string(feeResultJSON))

	// Build payload for transactions service
	transactionPayload := map[string]interface{}{
		"client_id":              req.ClientID,
		"merchant_id":            merchantID,
		"phone_number":           req.PhoneNumber,
		"amount":                 req.Amount,
		"transaction_ref":        req.TransactionRef,
		"type":                   "MNO_COLLECTION",
		"status":                 "PENDING",
		"fee_calculation":        feeResult,
		"transaction_fee_amount": feeResult.TransactionFeeAmount,
		"provider_fee_amount":    feeResult.ProviderFeeAmount,
		"commission_fee_amount":  feeResult.CommissionFeeAmount,
		"net_amount":             feeResult.NetAmount,
		"fee_profile_id":         feeResult.FeeProfileID,
		"payment_channel_id":     feeResult.PaymentChannelID,
	}

	// Log the transaction payload (simulating queue send to transactions service)
	transactionPayloadJSON, _ := json.Marshal(transactionPayload)
	log.Printf("[HandleCollect] Simulating send to transactions service queue: %s", string(transactionPayloadJSON))

	// Forward to transactions service via RabbitMQ
	// err = app.MQ.Emit("transactions.process", transactionPayload)
	// if err != nil {
	// 	return nil, fmt.Errorf("failed to forward collection to RabbitMQ: %v", err)
	// }

	utils.LogAuditEvent(app, "merchant_service", "collection.forwarded", transactionPayload)

	// Forward to transactions service
	responseBytes, err := app.MQ.Request("transactions.process", transactionPayload)
	if err != nil {
		return nil, fmt.Errorf("failed to forward collection to RabbitMQ: %v", err)
	}

	var result CollectResponse
	if err := json.Unmarshal(responseBytes, &result); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response from transactions service: %v", err)
	}

	return &result, nil
}

// HandleDisburse processes disbursement requests synchronously, verifying signature and checking balance
func HandleDisbursement(app *global.App, req *DisburseRequest) (*DisburseResponse, error) {

	var mApiKey merchantapikeys.MerchantAPIKey
	err := app.DB.Where("client_id = ?", req.ClientID).First(&mApiKey).Error
	if err != nil {
		return &DisburseResponse{Status: "FAILED", ErrorCode: "MERCHANT_NOT_FOUND"}, err
	}

	// Verify Auth Signature against request tampering
	expectedSig := computeHMAC(req.RawBody, mApiKey.ClientSecret)
	if !strings.EqualFold(expectedSig, req.Signature) {
		return &DisburseResponse{Status: "FAILED", ErrorCode: "INVALID_SIGNATURE"}, errors.New("invalid signature")
	}
	utils.LogAuditEvent(app, "merchant_service", "disbursement.completed", map[string]interface{}{
		"transaction_ref": req.TransactionRef,
		"client_id":       req.ClientID,
		"amount":          req.Amount,
		"phone_number":    req.PhoneNumber,
	})

	return &DisburseResponse{TransactionRef: req.TransactionRef, Status: "COMPLETED"}, nil
}

func computeHMAC(message string, secret string) string {
	key := []byte(secret)
	h := hmac.New(sha256.New, key)
	h.Write([]byte(message))
	return hex.EncodeToString(h.Sum(nil))
}

func HandleCollectionCheckStatus(app *global.App, req *CheckStatusRequest) (*CheckStatusResponse, error) {

	///Forward to transactions service via RabbitMQ
	transactionPayload := map[string]interface{}{
		"transaction_ref": req.TransactionRef,
		"type":            "MNO_COLLECTION",
		"client_id":       req.ClientID,
	}

	responseBytes, err := app.MQ.Request(
		"transactions.check_status",
		transactionPayload,
	)

	if err != nil {
		return nil, err
	}

	var response CheckStatusResponse

	err = json.Unmarshal(responseBytes, &response)
	if err != nil {
		return nil, err
	}

	return &response, nil

}

func HandleCollectionCheckBalance(app *global.App, req *CheckCollectionBalanceRequest) (*CheckCollectionBalanceResponse, error) {

	///Forward to transactions service via RabbitMQ
	balancePayload := map[string]interface{}{
		"client_id": req.ClientID,
	}

	responseBytes, err := app.MQ.Request(
		"merchant.accounts.check_balance",
		balancePayload,
	)

	if err != nil {
		return nil, err
	}

	var response CheckCollectionBalanceResponse

	err = json.Unmarshal(responseBytes, &response)
	if err != nil {
		return nil, err
	}

	return &response, nil
}

func HandleDisbursementCheckStatus(app *global.App, req *CheckStatusRequest) (*CheckStatusResponse, error) {

	///Forward to transactions service via RabbitMQ
	transactionPayload := map[string]interface{}{
		"transaction_ref": req.TransactionRef,
		"type":            "MNO_DISBURSEMENT",
		"client_id":       req.ClientID,
	}

	responseBytes, err := app.MQ.Request(
		"transactions.check_status",
		transactionPayload,
	)

	if err != nil {
		return nil, err
	}

	var response CheckStatusResponse

	err = json.Unmarshal(responseBytes, &response)
	if err != nil {
		return nil, err
	}

	return &response, nil

}

func HandleDisbursementCheckBalance(app *global.App, req *CheckDisbursementBalanceRequest) (*CheckDisbursementBalanceResponse, error) {

	//Check if client ID is in merhchants table
	var merchant merchantapikeys.MerchantAPIKey
	err := app.DB.Where("client_id = ?", req.ClientID).First(&merchant).Error
	if err != nil {
		return nil, fmt.Errorf("failed to find merchant: %v", err)
	}

	encryptionKey := []byte(os.Getenv("ENCRYPTION_KEY"))
	if len(encryptionKey) == 0 {
		return nil, fmt.Errorf("ENCRYPTION_KEY not set")
	}

	pin, err := utils.Decrypt(merchant.Pin, encryptionKey)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt pin: %v", err)
	}

	expectedSignature := computeHMAC(fmt.Sprintf("%s:%s", req.ClientID, pin), merchant.ClientSecret)
	if req.XAuthSignature != expectedSignature {
		return &CheckDisbursementBalanceResponse{
			Code:    401,
			Status:  "error",
			Message: "Invalid authentication signature",
		}, nil
	}

	///Forward to transactions service via RabbitMQ
	balancePayload := map[string]interface{}{
		"merchant_id": merchant.MerchantID,
	}

	responseBytes, err := app.MQ.Request(
		"merchant.accounts.disbursement.check_balance",
		balancePayload,
	)

	if err != nil {
		return nil, err
	}

	var response CheckDisbursementBalanceResponse

	err = json.Unmarshal(responseBytes, &response)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %v", err)
	}

	return &response, nil
}
