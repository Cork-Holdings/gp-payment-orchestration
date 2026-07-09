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
	// Calculate fees for this collection
	feeResult, err := feecalculator.CalculateFees(
		merchantID,
		req.PhoneNumber,
		req.Amount,
		feecalculator.TransactionTypeCollection,
	)
	if err != nil {
		return &CollectResponse{
			Code:    400,
			Status:  "failed",
			Message: feeResult.Error,
			Data:    nil,
		}
	}

	// Check if fee calculation returned an error status (e.g., unauthorized merchant-channel)
	if feeResult.Status == "error" {
		return &CollectResponse{
			Code:    400,
			Status:  "failed",
			Message: feeResult.Error,
			Data:    nil,
		}
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
		"status":                 "pending",
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

	// utils.LogAuditEvent(app, "merchant_service", "collection.forwarded", transactionPayload)

	// // Forward to transactions service
	responseBytes, err := app.MQ.Request("transactions.create", transactionPayload)

	if err != nil {
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

	//From this step transaction has been created
	//Get provider via prefix
	provider := GetProviderFromPhoneNumber(req.PhoneNumber)

	if provider == "Unsupported Provider for "+req.PhoneNumber {
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

	mnoRespBytes, err := app.MQ.Request("mno.collection.requests", mnoPayload)

	if err != nil {
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
		return &CollectResponse{
			Code:    mnoResp.Code,
			Status:  "failed",
			Message: "MNO service error: " + mnoResp.Message,
			Data:    nil,
		}
	}

	// MNO service accepted the request. Log and return success.
	// The MNO service will continue processing asynchronously and callback with results.
	log.Printf("[HandleCollect] MNO service accepted collection request for transaction_ref: %s", req.TransactionRef)

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
		return nil, err
	}

	var response CheckStatusResponse

	err = json.Unmarshal(responseBytes, &response)
	if err != nil {
		return nil, err
	}

	return &response, nil

}
