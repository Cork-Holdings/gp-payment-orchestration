package merchantapis

import (
	"encoding/json"
	"log"

	"github.com/Cork-Holdings/gp_payment_orchestration/internal/global"
	"github.com/Cork-Holdings/gp_payment_orchestration/internal/modules/feecalculator"
)

type CollectRequest struct {
	ClientID       string  `json:"client_id"`
	MerchantID     string  `json:"merchant_id"`
	PhoneNumber    string  `json:"phone_number"`
	Amount         float64 `json:"amount"`
	TransactionRef string  `json:"transaction_ref"`
}

type CollectResponse struct {
	Code    int         `json:"code"`
	Status  string      `json:"status"`
	Message string      `json:"message"`
	Data    interface{} `json:"data"`
}

func HandleCollection(app *global.App, req *CollectRequest) *CollectResponse {
	log.Printf("collection request received transaction_ref=%s client_id=%s amount=%.2f currency=ZMW", req.TransactionRef, req.ClientID, req.Amount)
	if len(req.PhoneNumber) < 12 {
		return &CollectResponse{
			Code:    400,
			Status:  "failed",
			Message: "Invalid Phone number",
			Data:    nil,
		}
	}

	if req.Amount <= 0.01 {
		return &CollectResponse{
			Code:    400,
			Status:  "failed",
			Message: "Minimum amount is K1",
			Data:    nil,
		}
	}

	merchant, err := FindMerchantByClientID(app, req.ClientID)

	if err != nil {
		return &CollectResponse{
			Code:    400,
			Status:  "failed",
			Message: "Invalid client_id",
			Data:    nil,
		}
	}

	merchantID := merchant.MerchantID.String()
	log.Printf("collection merchant resolved transaction_ref=%s merchant_id=%s amount=%.2f", req.TransactionRef, merchantID, req.Amount)
	// Calculate fees for this collection
	feeResult, err := feecalculator.CalculateFees(
		merchantID,
		req.PhoneNumber,
		req.Amount,
		feecalculator.TransactionTypeCollection,
	)
	if err != nil {
		log.Printf("collection fee calculation failed transaction_ref=%s merchant_id=%s amount=%.2f error=%v", req.TransactionRef, merchantID, req.Amount, err)
		return &CollectResponse{
			Code:    400,
			Status:  "failed",
			Message: feeResult.Error,
			Data:    nil,
		}
	}

	// Check if fee calculation returned an error status (e.g., unauthorized merchant-channel)
	if feeResult.Status == "error" {
		log.Printf("collection fee calculation rejected transaction_ref=%s merchant_id=%s amount=%.2f error=%s", req.TransactionRef, merchantID, req.Amount, feeResult.Error)
		return &CollectResponse{
			Code:    400,
			Status:  "failed",
			Message: feeResult.Error,
			Data:    nil,
		}
	}

	log.Printf("collection fees calculated transaction_ref=%s merchant_id=%s amount=%.2f net_amount=%.2f transaction_fee=%.2f provider_fee=%.2f commission_fee=%.2f fee_profile_id=%s payment_channel_id=%s",
		req.TransactionRef, merchantID, req.Amount, feeResult.NetAmount, feeResult.TransactionFeeAmount, feeResult.ProviderFeeAmount, feeResult.CommissionFeeAmount, feeResult.FeeProfileID, feeResult.PaymentChannelID)

	// Build payload for transactions service
	transactionPayload := map[string]interface{}{
		"client_id":              req.ClientID,
		"merchant_id":            merchantID,
		"phone_number":           req.PhoneNumber,
		"amount":                 req.Amount,
		"transaction_ref":        req.TransactionRef,
		"type":                   "MNO_COLLECTION",
		"status":                 "pending",
		"fee_calculation":        feeResult,
		"transaction_fee_amount": feeResult.TransactionFeeAmount,
		"provider_fee_amount":    feeResult.ProviderFeeAmount,
		"commission_fee_amount":  feeResult.CommissionFeeAmount,
		"net_amount":             feeResult.NetAmount,
		"fee_profile_id":         feeResult.FeeProfileID,
		"payment_channel_id":     feeResult.PaymentChannelID,
	}

	// Forward to transactions service via RabbitMQ
	// err = app.MQ.Emit("transactions.process", transactionPayload)
	// if err != nil {
	// 	return nil, fmt.Errorf("failed to forward collection to RabbitMQ: %v", err)
	// }

	// utils.LogAuditEvent(app, "merchant_service", "collection.forwarded", transactionPayload)

	// // Forward to transactions service
	responseBytes, err := app.MQ.Request("transactions.create", transactionPayload)

	if err != nil {
		log.Printf("collection transaction creation failed transaction_ref=%s merchant_id=%s amount=%.2f error=%v", req.TransactionRef, merchantID, req.Amount, err)
		return &CollectResponse{
			Code:    400,
			Status:  "failed",
			Message: "Unable to create transaction",
			Data:    nil,
		}
	}

	var transactionsResp CollectResponse

	if err := json.Unmarshal(responseBytes, &transactionsResp); err != nil {
		return &CollectResponse{
			Code:    400,
			Status:  "failed",
			Message: "Unable to unmarshal transaction response",
			Data:    nil,
		}
	}
	log.Printf("collection transaction created transaction_ref=%s merchant_id=%s amount=%.2f status=%s code=%d", req.TransactionRef, merchantID, req.Amount, transactionsResp.Status, transactionsResp.Code)

	//From this step transaction has been created
	//Get provider via prefix
	provider := GetProviderFromPhoneNumber(req.PhoneNumber)

	if provider == "Unsupported Provider for "+req.PhoneNumber {
		log.Printf("collection provider resolution failed transaction_ref=%s merchant_id=%s amount=%.2f", req.TransactionRef, merchantID, req.Amount)
		return &CollectResponse{
			Code:    400,
			Status:  "failed",
			Message: "Unsupported Provider for " + req.PhoneNumber,
			Data:    nil,
		}
	}

	//Send request to MNO service
	mnoPayload := map[string]interface{}{
		"transaction_ref": req.TransactionRef,
		"phone_number":    req.PhoneNumber,
		"amount":          req.Amount,
		"currency":        "ZMW", // Default currency
		"provider":        provider,
		"callback_url":    "", // Optional: can be set from config if needed
	}

	//Log the request to the MNO service
	log.Printf("collection MNO dispatch request transaction_ref=%s merchant_id=%s provider=%s amount=%.2f request=%+v", req.TransactionRef, merchantID, provider, req.Amount, mnoPayload)

	mnoRespBytes, err := app.MQ.Request("mno.collection.requests", mnoPayload)

	if err != nil {
		log.Printf("collection MNO dispatch failed transaction_ref=%s merchant_id=%s provider=%s amount=%.2f error=%v", req.TransactionRef, merchantID, provider, req.Amount, err)
		return &CollectResponse{
			Code:    400,
			Status:  "failed",
			Message: "MNO service request failed: " + err.Error(),
			Data:    nil,
		}
	}

	//Example successful response from MNO service
	// {
	//    "code": 200,
	//    "status": "success",
	//    "message": "Collection request accepted and processing",
	//    "data": {
	//      "transaction_ref":"e162f10b-06ab-4168-bdab-e24a2cff99b2",
	//      "phone_number":"260956587842",
	//      "amount":5,
	//      "currency":"ZMW",
	//      "provider":"zamtel",
	//      "external_reference":"V406HWGD",
	//      "status":"pending",
	//      "processed_at":"2026-06-29T08:13:06Z"
	//    }
	// }

	var mnoResp CollectResponse

	if err := json.Unmarshal(mnoRespBytes, &mnoResp); err != nil {
		return &CollectResponse{
			Code:    400,
			Status:  "failed",
			Message: "Unable to unmarshal mno response",
			Data:    nil,
		}
	}

	// Check if MNO service returned an error
	if mnoResp.Code != 200 {
		log.Printf("collection MNO rejected transaction_ref=%s merchant_id=%s provider=%s amount=%.2f code=%d status=%s message=%s", req.TransactionRef, merchantID, provider, req.Amount, mnoResp.Code, mnoResp.Status, mnoResp.Message)
		return &CollectResponse{
			Code:    mnoResp.Code,
			Status:  "failed",
			Message: "MNO service error: " + mnoResp.Message,
			Data:    nil,
		}
	}

	// MNO service accepted the request. Log and return success.
	// The MNO service will continue processing asynchronously and callback with results.
	log.Printf("collection forwarded to MNO transaction_ref=%s merchant_id=%s provider=%s amount=%.2f status=%s", req.TransactionRef, merchantID, provider, req.Amount, mnoResp.Status)

	return &CollectResponse{
		Code:    200,
		Status:  "success",
		Message: "Collection request forwarded to MNO service and is now processing",
		Data:    mnoResp.Data,
	}
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
		log.Printf("collection status check failed transaction_ref=%s client_id=%s error=%v", req.TransactionRef, req.ClientID, err)
		return nil, err
	}

	var response CheckStatusResponse

	err = json.Unmarshal(responseBytes, &response)
	if err != nil {
		return nil, err
	}
	log.Printf("collection status checked transaction_ref=%s client_id=%s status=%s", req.TransactionRef, req.ClientID, response.Status)

	return &response, nil

}
