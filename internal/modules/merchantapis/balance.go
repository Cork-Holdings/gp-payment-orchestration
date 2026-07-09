package merchantapis

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/Cork-Holdings/gp_payment_orchestration/internal/global"
	"github.com/Cork-Holdings/gp_payment_orchestration/utils"
)

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

func HandleCollectionCheckBalance(app *global.App, req *CheckCollectionBalanceRequest) (*CheckCollectionBalanceResponse, error) {

	merchant, err := FindMerchantByClientID(app, req.ClientID)

	if err != nil {
		return nil, fmt.Errorf("merchant not found")
	}

	///Forward to transactions service via RabbitMQ
	balancePayload := map[string]interface{}{
		"merchant_id": merchant.MerchantID,
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

func HandleDisbursementCheckBalance(app *global.App, req *CheckDisbursementBalanceRequest) (*CheckDisbursementBalanceResponse, error) {

	//Check if client ID is in merhchants table
	merchant, err := FindMerchantByClientID(app, req.ClientID)

	if err != nil {
		return nil, fmt.Errorf("failed to find merchant: %v", err)
	}

	// Decrypt PIN stored in database
	encryptionKey := []byte(os.Getenv("ENCRYPTION_KEY"))

	if len(encryptionKey) == 0 {
		return nil, fmt.Errorf("ENCRYPTION_KEY not set")
	}

	pin, err := utils.Decrypt(merchant.Pin, encryptionKey)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt pin: %v", err)
	}

	clientSecret, err := utils.Decrypt(
		merchant.ClientSecret,
		encryptionKey,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to decrypt client secret: %v", err)
	}

	expectedSignature := computeHMAC(
		fmt.Sprintf("%s:%s", req.ClientID, pin),
		clientSecret,
	)
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
		"amount":      0.0,
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
