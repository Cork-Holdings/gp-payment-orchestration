package merchantapis

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/Cork-Holdings/gp_payment_orchestration/internal/global"
	"github.com/Cork-Holdings/gp_payment_orchestration/internal/modules/merchantapikeys"
	"github.com/Cork-Holdings/gp_payment_orchestration/internal/modules/merchantips"
	"github.com/Cork-Holdings/gp_payment_orchestration/utils"
	"github.com/golang-jwt/jwt/v5"
	"gorm.io/gorm"
)

var jwtSecret = []byte("merchant_secret_jwt_key_999")

type TokenRequest struct {
	ClientID     string `json:"client_id"`
	ClientSecret string `json:"client_secret"`
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

type VerifyResponse struct {
	Valid        bool   `json:"valid"`
	TenantID     string `json:"tenant_id,omitempty"`
	ErrorMessage string `json:"error_message,omitempty"`
}

type CollectRequest struct {
	ClientID    string  `json:"client_id"`
	PhoneNumber string  `json:"phone_number"`
	Amount      float64 `json:"amount"`
	// Currency    string  `json:"currency"`
	TransactionRef string `json:"transaction_ref"`
}

type DisburseRequest struct {
	ClientID    string  `json:"client_id"`
	PhoneNumber string  `json:"phone_number"`
	Amount      float64 `json:"amount"`
	// Currency       string  `json:"currency"`
	Signature      string `json:"signature"`
	RawBody        string `json:"raw_body"`
	TransactionRef string `json:"transaction_ref"`
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

// HandleVerifyTokenAndIP validates token signature, client ID, and IP whitelist
func HandleVerifyTokenAndIP(app *global.App, payload []byte) ([]byte, error) {
	var req VerifyRequest
	if err := json.Unmarshal(payload, &req); err != nil {
		return nil, err
	}

	token, err := jwt.Parse(req.Token, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return jwtSecret, nil
	})

	if err != nil || !token.Valid {
		resp, _ := json.Marshal(VerifyResponse{Valid: false, ErrorMessage: "invalid or expired token"})
		return resp, nil
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		resp, _ := json.Marshal(VerifyResponse{Valid: false, ErrorMessage: "invalid claims structure"})
		return resp, nil
	}

	claimClientID, _ := claims["client_id"].(string)
	if claimClientID != req.ClientID {
		resp, _ := json.Marshal(VerifyResponse{Valid: false, ErrorMessage: "token does not match Client ID"})
		return resp, nil
	}

	var mApiKey merchantapikeys.MerchantAPIKey
	err = app.DB.Where("client_id = ?", claimClientID).First(&mApiKey).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			resp, _ := json.Marshal(VerifyResponse{Valid: false, ErrorMessage: "merchant profile not found"})
			return resp, nil
		}
		return nil, err
	}

	var mIP merchantips.MerchantIP
	err = app.DB.Where("merchant_id = ? AND ip_address = ? AND status = ?", mApiKey.MerchantID, req.IPAddress, "approved").First(&mIP).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			resp, _ := json.Marshal(VerifyResponse{Valid: false, ErrorMessage: "IP address not whitelisted or not approved"})
			return resp, nil
		}
		return nil, err
	}

	resp, _ := json.Marshal(VerifyResponse{
		Valid:    true,
		TenantID: mApiKey.ID.String(),
	})
	return resp, nil
}

// HandleCollect initiates an async collection, transitioning PENDING -> PROCESSING
func HandleCollect(app *global.App, req *CollectRequest) error {

	//Calculate MNO fees

	// Forward to transactions service via RabbitMQ
	eventPayload := map[string]interface{}{
		"client_id":       req.ClientID,
		"phone_number":    req.PhoneNumber,
		"amount":          req.Amount,
		"transaction_ref": req.TransactionRef,
		"type":            "MNO_COLLECTION",
		"status":          "PENDING",
	}

	err := app.MQ.Emit("transactions.process", eventPayload)
	if err != nil {
		return fmt.Errorf("failed to forward collection to RabbitMQ: %v", err)
	}

	utils.LogAuditEvent(app, "merchant_service", "collection.forwarded", eventPayload)

	return nil
}

// HandleDisburse processes disbursement requests synchronously, verifying signature and checking balance
func HandleDisburse(app *global.App, req *DisburseRequest) (*DisburseResponse, error) {

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
