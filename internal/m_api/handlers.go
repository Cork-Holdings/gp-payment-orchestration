package m_api

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"strings"
	"time"

	"github.com/Cork-Holdings/gp_payment_orchestration/internal/common"
	"github.com/Cork-Holdings/gp_payment_orchestration/internal/global"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type TokenRequest struct {
	ClientID     string `json:"client_id"`
	ClientSecret string `json:"client_secret"`
	IPAddress    string `json:"ip_address"`
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
	Currency    string  `json:"currency"`
	TrackingRef string  `json:"tracking_ref"`
}

type DisburseRequest struct {
	ClientID    string  `json:"client_id"`
	PhoneNumber string  `json:"phone_number"`
	Amount      float64 `json:"amount"`
	Currency    string  `json:"currency"`
	Signature   string  `json:"signature"`
	RawBody     string  `json:"raw_body"`
	TrackingRef string  `json:"tracking_ref"`
}

type DisburseResponse struct {
	TrackingRef string `json:"tracking_ref,omitempty"`
	Status      string `json:"status"`
	ErrorCode   string `json:"error_code,omitempty"`
}

// HandleGenerateToken validates credentials, IP whitelist, and generates a Bearer JWT token
func HandleGenerateToken(app *global.App, payload []byte) ([]byte, error) {
	var req TokenRequest
	if err := json.Unmarshal(payload, &req); err != nil {
		return nil, err
	}

	var mProfile MerchantProfile
	err := app.DB.Where("client_id = ?", req.ClientID).First(&mProfile).Error
	if err != nil {
		LogAuditEvent(app, "auth_service", "token.denied", map[string]string{
			"client_id": req.ClientID,
			"reason":    "invalid_credentials",
			"ip":        req.IPAddress,
		})
		resp, _ := json.Marshal(TokenResponse{Error: "invalid client_id or client_secret"})
		return resp, nil
	}

	if !verifyClientSecret(mProfile.ClientSecret, req.ClientSecret) {
		LogAuditEvent(app, "auth_service", "token.denied", map[string]string{
			"client_id": req.ClientID,
			"reason":    "invalid_credentials",
			"ip":        req.IPAddress,
		})
		resp, _ := json.Marshal(TokenResponse{Error: "invalid client_id or client_secret"})
		return resp, nil
	}

	if req.IPAddress != "" && !isIPAllowed(req.IPAddress, mProfile.AllowedIPs) {
		LogAuditEvent(app, "auth_service", "token.denied", map[string]string{
			"client_id": req.ClientID,
			"reason":    "ip_not_whitelisted",
			"ip":        req.IPAddress,
		})
		resp, _ := json.Marshal(TokenResponse{Error: "IP address not whitelisted"})
		return resp, nil
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

	LogAuditEvent(app, "auth_service", "token.generated", map[string]string{
		"client_id": req.ClientID,
	})

	resp, _ := json.Marshal(TokenResponse{
		AccessToken: tokenString,
		ExpiresIn:   expiresIn,
		TokenType:   "Bearer",
	})
	return resp, nil
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
		LogAuditEvent(app, "auth_service", "token.verify_failed", map[string]string{
			"client_id": req.ClientID,
			"reason":    "invalid_or_expired_token",
			"ip":        req.IPAddress,
		})
		resp, _ := json.Marshal(VerifyResponse{Valid: false, ErrorMessage: "invalid or expired token"})
		return resp, nil
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		LogAuditEvent(app, "auth_service", "token.verify_failed", map[string]string{
			"client_id": req.ClientID,
			"reason":    "invalid_claims",
			"ip":        req.IPAddress,
		})
		resp, _ := json.Marshal(VerifyResponse{Valid: false, ErrorMessage: "invalid claims structure"})
		return resp, nil
	}

	claimClientID, _ := claims["client_id"].(string)
	if claimClientID != req.ClientID {
		LogAuditEvent(app, "auth_service", "token.verify_failed", map[string]string{
			"client_id": req.ClientID,
			"reason":    "client_id_mismatch",
			"ip":        req.IPAddress,
		})
		resp, _ := json.Marshal(VerifyResponse{Valid: false, ErrorMessage: "token does not match Client ID"})
		return resp, nil
	}

	var mProfile MerchantProfile
	err = app.DB.Where("client_id = ?", claimClientID).First(&mProfile).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			LogAuditEvent(app, "auth_service", "token.verify_failed", map[string]string{
				"client_id": req.ClientID,
				"reason":    "merchant_not_found",
				"ip":        req.IPAddress,
			})
			resp, _ := json.Marshal(VerifyResponse{Valid: false, ErrorMessage: "merchant profile not found"})
			return resp, nil
		}
		return nil, err
	}

	if !isIPAllowed(req.IPAddress, mProfile.AllowedIPs) {
		LogAuditEvent(app, "auth_service", "token.verify_failed", map[string]string{
			"client_id": req.ClientID,
			"reason":    "ip_not_whitelisted",
			"ip":        req.IPAddress,
		})
		resp, _ := json.Marshal(VerifyResponse{Valid: false, ErrorMessage: "IP address not whitelisted"})
		return resp, nil
	}

	resp, _ := json.Marshal(VerifyResponse{
		Valid:    true,
		TenantID: mProfile.ID.String(),
	})
	return resp, nil
}

// HandleCollect initiates an async collection, transitioning PENDING -> PROCESSING
func HandleCollect(app *global.App, payload []byte) error {
	var req CollectRequest
	if err := json.Unmarshal(payload, &req); err != nil {
		return err
	}

	now := time.Now()
	txn := &MerchantTransaction{
		ID:          uuid.New(),
		TrackingRef: req.TrackingRef,
		ClientID:    req.ClientID,
		Type:        "COLLECT",
		PhoneNumber: req.PhoneNumber,
		Amount:      req.Amount,
		Currency:    req.Currency,
		Status:      "PENDING",
		Entity: common.Entity{
			CreatedAt: &now,
			UpdatedAt: &now,
		},
	}

	if err := app.DB.Create(txn).Error; err != nil {
		return err
	}

	LogAuditEvent(app, "merchant_service", "collection.initiated", txn)

	// Asynchronously transition PENDING -> PROCESSING
	go func() {
		time.Sleep(200 * time.Millisecond)
		var updatedTxn MerchantTransaction
		err := app.DB.Where("tracking_ref = ?", req.TrackingRef).First(&updatedTxn).Error
		if err == nil {
			updatedTxn.Status = "PROCESSING"
			tNow := time.Now()
			updatedTxn.UpdatedAt = &tNow
			app.DB.Save(&updatedTxn)
			LogAuditEvent(app, "merchant_service", "collection.processing", updatedTxn)
		}
	}()

	return nil
}

// HandleDisburse processes disbursement requests synchronously, verifying signature and checking balance
func HandleDisburse(app *global.App, payload []byte) ([]byte, error) {
	var req DisburseRequest
	if err := json.Unmarshal(payload, &req); err != nil {
		return nil, err
	}

	var mProfile MerchantProfile
	err := app.DB.Where("client_id = ?", req.ClientID).First(&mProfile).Error
	if err != nil {
		resp, _ := json.Marshal(DisburseResponse{Status: "FAILED", ErrorCode: "MERCHANT_NOT_FOUND"})
		return resp, nil
	}

	// Verify Auth Signature against request tampering
	signingSecret, err := secretForHMAC(mProfile.ClientSecret)
	if err != nil {
		return nil, err
	}
	if !verifyHMAC(req.RawBody, signingSecret, req.Signature) {
		LogAuditEvent(app, "merchant_service", "disbursement.denied", map[string]string{
			"client_id": req.ClientID,
			"reason":    "invalid_signature",
		})
		resp, _ := json.Marshal(DisburseResponse{Status: "FAILED", ErrorCode: "INVALID_SIGNATURE"})
		return resp, nil
	}

	if req.Amount <= 0 {
		resp, _ := json.Marshal(DisburseResponse{Status: "FAILED", ErrorCode: "INVALID_AMOUNT"})
		return resp, nil
	}

	if strings.TrimSpace(req.Currency) == "" {
		resp, _ := json.Marshal(DisburseResponse{Status: "FAILED", ErrorCode: "INVALID_CURRENCY"})
		return resp, nil
	}

	if !phoneValidationRegex.MatchString(req.PhoneNumber) {
		resp, _ := json.Marshal(DisburseResponse{Status: "FAILED", ErrorCode: "INVALID_PHONE_NUMBER"})
		return resp, nil
	}

	// Verify balance constraints (short-circuit on low balance)
	if mProfile.WalletBalance < req.Amount {
		resp, _ := json.Marshal(DisburseResponse{Status: "FAILED", ErrorCode: "INSUFFICIENT_BALANCE"})
		return resp, nil
	}

	now := time.Now()
	err = app.DB.Transaction(func(tx *gorm.DB) error {
		mProfile.WalletBalance -= req.Amount
		mProfile.UpdatedAt = &now
		if err := tx.Save(&mProfile).Error; err != nil {
			return err
		}

		txn := &MerchantTransaction{
			ID:          uuid.New(),
			TrackingRef: req.TrackingRef,
			ClientID:    req.ClientID,
			Type:        "DISBURSE",
			PhoneNumber: req.PhoneNumber,
			Amount:      req.Amount,
			Currency:    req.Currency,
			Status:      "COMPLETED",
			Entity: common.Entity{
				CreatedAt: &now,
				UpdatedAt: &now,
			},
		}
		if err := tx.Create(txn).Error; err != nil {
			return err
		}
		return nil
	})

	if err != nil {
		return nil, err
	}

	LogAuditEvent(app, "merchant_service", "disbursement.completed", map[string]interface{}{
		"tracking_ref": req.TrackingRef,
		"client_id":    req.ClientID,
		"amount":       req.Amount,
		"phone_number": req.PhoneNumber,
	})

	resp, _ := json.Marshal(DisburseResponse{
		TrackingRef: req.TrackingRef,
		Status:      "COMPLETED",
	})
	return resp, nil
}

func isIPAllowed(clientIP string, allowedIPsStr string) bool {
	clientIP = strings.TrimSpace(clientIP)
	ip := net.ParseIP(clientIP)
	if ip == nil {
		return false
	}

	allowedIPs := strings.Split(allowedIPsStr, ",")
	for _, allowed := range allowedIPs {
		allowed = strings.TrimSpace(allowed)
		if allowed == "" {
			continue
		}

		if strings.Contains(allowed, "/") {
			_, subnet, err := net.ParseCIDR(allowed)
			if err == nil {
				if subnet.Contains(ip) {
					return true
				}
			}
		} else {
			if allowed == clientIP {
				return true
			}
		}
	}
	return false
}

func computeHMAC(message string, secret string) string {
	key := []byte(secret)
	h := hmac.New(sha256.New, key)
	h.Write([]byte(message))
	return hex.EncodeToString(h.Sum(nil))
}
