package merchantapis

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/Cork-Holdings/gp_payment_orchestration/internal/global"
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

// HandleGenerateToken validates credentials and generates a Bearer JWT token
// func HandleGenerateToken(app *global.App, req TokenRequest) (*TokenResponse, error) {

// 	_, err := FindMerchantByClientID(app, req.ClientID)
// 	if err != nil {
// 		return &TokenResponse{
// 			Error: "invalid client_id or client_secret",
// 		}, nil
// 	}

// 	expiresIn := int64(3600) // 1 hour
// 	expirationTime := time.Now().Add(time.Duration(expiresIn) * time.Second)

// 	claims := &jwt.MapClaims{
// 		"client_id": req.ClientID,
// 		"exp":       expirationTime.Unix(),
// 	}

// 	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
// 	tokenString, err := token.SignedString(jwtSecret)
// 	if err != nil {
// 		return nil, err
// 	}

// 	utils.LogAuditEvent(app, "auth_service", "token.generated", map[string]string{
// 		"client_id": req.ClientID,
// 	})

// 	return &TokenResponse{AccessToken: tokenString, ExpiresIn: expiresIn, TokenType: "Bearer"}, nil
// }

func HandleGenerateToken(
	app *global.App,
	req TokenRequest,
) (*TokenResponse, error) {

	merchant, err := FindMerchantByClientID(
		app,
		req.ClientID,
	)

	if err != nil {
		return &TokenResponse{
			Error: "invalid client_id or client_secret",
		}, nil
	}

	encryptionKey := []byte(os.Getenv("ENCRYPTION_KEY"))

	storedSecret, err := utils.Decrypt(
		merchant.ClientSecret,
		encryptionKey,
	)

	if err != nil {
		return nil, err
	}

	if storedSecret != req.ClientSecret {

		return &TokenResponse{
			Error: "invalid client_id or client_secret",
		}, nil
	}

	expiresIn := int64(3600)

	expirationTime := time.Now().
		Add(time.Duration(expiresIn) * time.Second)

	claims := jwt.MapClaims{
		"client_id":   req.ClientID,
		"merchant_id": merchant.MerchantID.String(),
		"exp":         expirationTime.Unix(),
	}

	token := jwt.NewWithClaims(
		jwt.SigningMethodHS256,
		claims,
	)

	tokenString, err := token.SignedString(jwtSecret)

	if err != nil {
		return nil, err
	}

	utils.LogAuditEvent(
		app,
		"auth_service",
		"token.generated",
		map[string]string{
			"merchant_id": merchant.MerchantID.String(),
		},
	)

	return &TokenResponse{
		AccessToken: tokenString,
		ExpiresIn:   expiresIn,
		TokenType:   "Bearer",
	}, nil
}

// VerifyTokenAndIPDirect validates token signature, client ID, and IP whitelist (direct call, no RabbitMQ)
// func VerifyTokenAndIPDirect(app *global.App, req *VerifyRequest) *VerifyResponse {
// 	token, err := jwt.Parse(req.Token, func(token *jwt.Token) (interface{}, error) {
// 		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
// 			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
// 		}
// 		return jwtSecret, nil
// 	})

// 	if err != nil || !token.Valid {
// 		return &VerifyResponse{Valid: false, ErrorMessage: "invalid or expired token"}
// 	}

// 	claims, ok := token.Claims.(jwt.MapClaims)
// 	if !ok {
// 		return &VerifyResponse{Valid: false, ErrorMessage: "invalid claims structure"}
// 	}

// 	claimClientID, _ := claims["client_id"].(string)
// 	if claimClientID != req.ClientID {
// 		fmt.Printf("[Auth] Client ID mismatch: token=%s, header=%s\n", claimClientID, req.ClientID)
// 		return &VerifyResponse{Valid: false, ErrorMessage: "token does not match Client ID"}
// 	}

// 	var mApiKey merchantapikeys.MerchantAPIKey
// 	lookupErr := app.DB.Where("client_id = ?", claimClientID).First(&mApiKey).Error
// 	if lookupErr != nil {
// 		if errors.Is(lookupErr, gorm.ErrRecordNotFound) {
// 			fmt.Printf("[Auth] Merchant profile not found for client_id: %s\n", claimClientID)
// 			return &VerifyResponse{Valid: false, ErrorMessage: "merchant profile not found"}
// 		}
// 		fmt.Printf("[Auth] DB error: %v\n", lookupErr)
// 		return &VerifyResponse{Valid: false, ErrorMessage: "database error"}
// 	}

// 	var mIP merchantips.MerchantIP
// 	err = app.DB.Where("merchant_id = ? AND ip_address = ? AND status = ?", mApiKey.MerchantID, req.IPAddress, "approved").First(&mIP).Error
// 	if err != nil {
// 		if errors.Is(err, gorm.ErrRecordNotFound) {
// 			fmt.Printf("[Auth] IP not whitelisted: merchant=%s, ip=%s\n", mApiKey.MerchantID, req.IPAddress)
// 			return &VerifyResponse{Valid: false, ErrorMessage: "IP address not whitelisted or not approved"}
// 		}
// 		fmt.Printf("[Auth] DB error checking IP: %v\n", err)
// 		return &VerifyResponse{Valid: false, ErrorMessage: "database error"}
// 	}

// 	return &VerifyResponse{
// 		Valid:      true,
// 		TenantID:   mApiKey.ID.String(),
// 		MerchantID: mApiKey.MerchantID.String(),
// 	}
// }

func VerifyTokenAndIPDirect(app *global.App, req *VerifyRequest) *VerifyResponse {

	token, err := jwt.Parse(req.Token, func(token *jwt.Token) (interface{}, error) {

		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf(
				"unexpected signing method: %v",
				token.Header["alg"],
			)
		}

		return jwtSecret, nil
	})

	if err != nil || !token.Valid {
		return &VerifyResponse{
			Valid:        false,
			ErrorMessage: "invalid or expired token",
		}
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return &VerifyResponse{
			Valid:        false,
			ErrorMessage: "invalid claims structure",
		}
	}

	claimClientID, ok := claims["client_id"].(string)

	if !ok || claimClientID == "" {
		return &VerifyResponse{
			Valid:        false,
			ErrorMessage: "client_id missing from token",
		}
	}

	if claimClientID != req.ClientID {

		return &VerifyResponse{
			Valid:        false,
			ErrorMessage: "token does not match client ID",
		}
	}

	// Find merchant using decrypted client ID
	mApiKey, err := FindMerchantByClientID(
		app,
		claimClientID,
	)

	if err != nil {

		if errors.Is(err, gorm.ErrRecordNotFound) {
			return &VerifyResponse{
				Valid:        false,
				ErrorMessage: "merchant profile not found",
			}
		}

		return &VerifyResponse{
			Valid:        false,
			ErrorMessage: "merchant lookup failed",
		}
	}

	// Verify merchant IP whitelist

	var mIP merchantips.MerchantIP

	err = app.DB.
		Where(
			"merchant_id = ? AND ip_address = ? AND status = ?",
			mApiKey.MerchantID,
			req.IPAddress,
			"approved",
		).
		First(&mIP).
		Error

	if err != nil {

		if errors.Is(err, gorm.ErrRecordNotFound) {

			return &VerifyResponse{
				Valid:        false,
				ErrorMessage: "IP address not whitelisted",
			}
		}

		return &VerifyResponse{
			Valid:        false,
			ErrorMessage: "database error",
		}
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

func computeHMAC(message string, secret string) string {
	key := []byte(secret)
	h := hmac.New(sha256.New, key)
	h.Write([]byte(message))
	return hex.EncodeToString(h.Sum(nil))
}

// HandleNameLookup performs name lookup via MNO service (MTN/Airtel/Zamtel APIs)
// Validates merchant, phone number format, and provider support before forwarding to MNO service.
// Returns account holder name from the MNO API via RPC response.
