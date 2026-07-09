package merchantapis

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/Cork-Holdings/gp_payment_orchestration/internal/global"
	"github.com/Cork-Holdings/gp_payment_orchestration/internal/modules/feecalculator"
	"github.com/Cork-Holdings/gp_payment_orchestration/utils"
)

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
	Error          string `json:"error,omitempty"`
}

// HandleDisburse processes disbursement requests synchronously, verifying signature and checking balance
func HandleDisbursement(app *global.App, req *DisburseRequest) (*DisburseResponse, error) {

	merchant, err := FindMerchantByClientID(app, req.ClientID)

	if err != nil {
		return &DisburseResponse{
			Status: "failed",
			Error:  "MERCHANT_NOT_FOUND",
		}, err
	}
	//Generate expected signature using the client secret and the request body
	clientID := req.ClientID

	encryptionKey := []byte(os.Getenv("ENCRYPTION_KEY"))

	if len(encryptionKey) == 0 {
		return &DisburseResponse{
			Status: "failed",
			Error:  "ENCRYPTION_KEY_NOT_SET",
		}, errors.New("ENCRYPTION_KEY not set")
	}

	clientSecret := merchant.ClientSecret
	//Decrypt the Pin from the database
	// encryptionKey = []byte(os.Getenv("ENCRYPTION_KEY"))
	// if len(encryptionKey) == 0 {
	// 	return &DisburseResponse{Status: "failed", Error: "ENCRYPTION_KEY_NOT_SET"}, errors.New("ENCRYPTION_KEY not set")
	// }

	decryptedClientSecret, err := utils.Decrypt(clientSecret, encryptionKey)
	if err != nil {
		return &DisburseResponse{Status: "failed", Error: "CLIENT_SECRET_DECRYPTION_FAILED"}, err
	}

	pin := merchant.Pin

	decryptedPin, err := utils.Decrypt(pin, encryptionKey)
	if err != nil {
		return &DisburseResponse{Status: "failed", Error: "PIN_DECRYPTION_FAILED"}, err
	}

	message := fmt.Sprintf("%s:%s", clientID, decryptedPin)

	signature := hmac.New(sha256.New, []byte(decryptedClientSecret))
	signature.Write([]byte(message))

	expectedSig := hex.EncodeToString(signature.Sum(nil))

	if !strings.EqualFold(expectedSig, req.Signature) {
		return &DisburseResponse{Status: "failed", Error: "INVALID_SIGNATURE"}, errors.New("invalid signature")
	}

	merchantID := merchant.MerchantID.String()

	balanceRespBytes, err := app.MQ.Request("merchant.accounts.disbursement.check_balance", map[string]any{
		"merchant_id": merchantID,
		"amount":      req.Amount,
	})
	if err != nil {
		return &DisburseResponse{Status: "failed", Error: "BALANCE_CHECK_failed"}, fmt.Errorf("failed to check merchant float balance: %w", err)
	}

	var balanceResp struct {
		Code    int    `json:"code"`
		Status  string `json:"status"`
		Message string `json:"message"`
	}
	if err := json.Unmarshal(balanceRespBytes, &balanceResp); err != nil {
		return &DisburseResponse{Status: "failed", Error: "BALANCE_CHECK_failed"}, fmt.Errorf("invalid balance response payload: %w", err)
	}

	if balanceResp.Code != 200 {
		if balanceResp.Message == "" {
			balanceResp.Message = "insufficient float balance"
		}
		return &DisburseResponse{Status: "failed", Error: "INSUFFICIENT_FLOAT_BALANCE"}, errors.New(balanceResp.Message)
	}

	feeResult, err := feecalculator.CalculateFees(
		merchantID,
		req.PhoneNumber,
		req.Amount,
		feecalculator.TransactionTypeDisbursement,
	)
	if err != nil {
		if feeResult.Error == "" {
			feeResult.Error = err.Error()
		}
		return &DisburseResponse{Status: "failed", Error: "FEE_CALCULATION_failed"}, errors.New(feeResult.Error)
	}

	// Check if fee calculation returned an error status (e.g., unauthorized merchant-channel)
	if feeResult.Status == "error" {
		return &DisburseResponse{Status: "failed", Error: "FEE_CALCULATION_failed"}, errors.New(feeResult.Error)
	}

	transactionPayload := map[string]any{
		"client_id":              req.ClientID,
		"merchant_id":            merchantID,
		"phone_number":           req.PhoneNumber,
		"amount":                 req.Amount,
		"transaction_ref":        req.TransactionRef,
		"type":                   "MNO_DISBURSEMENT",
		"status":                 "pending",
		"fee_calculation":        feeResult,
		"transaction_fee_amount": feeResult.TransactionFeeAmount,
		"provider_fee_amount":    feeResult.ProviderFeeAmount,
		"commission_fee_amount":  feeResult.CommissionFeeAmount,
		"net_amount":             feeResult.NetAmount,
		"fee_profile_id":         feeResult.FeeProfileID,
		"payment_channel_id":     feeResult.PaymentChannelID,
		"description":            req.Narration,
	}

	if _, err := app.MQ.Request("transactions.create", transactionPayload); err != nil {
		return &DisburseResponse{Status: "failed", Error: "TRANSACTION_CREATE_failed"}, fmt.Errorf("failed to create disbursement transaction: %w", err)
	}

	provider := GetProviderFromPhoneNumber(req.PhoneNumber)
	if provider == "Unsupported Provider for "+req.PhoneNumber {
		return &DisburseResponse{Status: "failed", Error: "UNSUPPORTED_PROVIDER"}, errors.New(provider)
	}

	mnoPayload := map[string]any{
		"transaction_ref": req.TransactionRef,
		"phone_number":    req.PhoneNumber,
		"amount":          req.Amount,
		"currency":        "ZMW",
		"provider":        provider,
		"narration":       req.Narration,
	}

	mnoRespBytes, err := app.MQ.Request("mno.disbursement.requests", mnoPayload)
	if err != nil {
		return &DisburseResponse{Status: "failed", Error: "MNO_REQUEST_failed"}, fmt.Errorf("failed to forward disbursement to mno service: %w", err)
	}

	var mnoResp struct {
		Code    int    `json:"code"`
		Status  string `json:"status"`
		Message string `json:"message"`
	}
	if err := json.Unmarshal(mnoRespBytes, &mnoResp); err != nil {
		return &DisburseResponse{Status: "failed", Error: "MNO_RESPONSE_INVALID"}, fmt.Errorf("failed to parse mno response: %w", err)
	}

	if mnoResp.Code != 200 {
		if mnoResp.Message == "" {
			mnoResp.Message = "mno service rejected disbursement"
		}
		return &DisburseResponse{Status: "failed", Error: "MNO_REJECTED"}, errors.New(mnoResp.Message)
	}

	utils.LogAuditEvent(app, "merchant_service", "disbursement.completed", map[string]interface{}{
		"transaction_ref": req.TransactionRef,
		"client_id":       req.ClientID,
		"merchant_id":     merchantID,
		"amount":          req.Amount,
		"phone_number":    req.PhoneNumber,
	})

	return &DisburseResponse{TransactionRef: req.TransactionRef, Status: "PROCESSING"}, nil
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
