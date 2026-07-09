package merchantapis

import (
	"encoding/json"
	"log"
	"strings"

	"github.com/Cork-Holdings/gp_payment_orchestration/internal/global"
	"github.com/Cork-Holdings/gp_payment_orchestration/utils"
)

type NameLookupRequest struct {
	ClientID    string `json:"client_id"`
	PhoneNumber string `json:"phone_number"`
}

type NameLookupData struct {
	PhoneNumber string `json:"phone_number"`
	Provider    string `json:"provider"`
	Status      string `json:"status"`
	Names       string `json:"names"`
	Message     string `json:"message"`
	ProcessedAt string `json:"processed_at"`
}

type NameLookupResponse struct {
	Code    int             `json:"code"`
	Status  string          `json:"status"`
	Message string          `json:"message"`
	Data    *NameLookupData `json:"data,omitempty"`
}

func HandleNameLookup(app *global.App, req *NameLookupRequest) (*NameLookupResponse, error) {

	// Find merchant using decrypted client_id lookup helper
	merchant, err := FindMerchantByClientID(app, req.ClientID)
	if err != nil {
		return &NameLookupResponse{
			Code:    400,
			Status:  "error",
			Message: "Merchant not found for client_id: " + req.ClientID,
		}, nil
	}

	// Validate phone number format: must be 12+ digits (Zambian format: 260XXXXXXXXX)
	if len(req.PhoneNumber) < 12 {
		return &NameLookupResponse{
			Code:    400,
			Status:  "error",
			Message: "Invalid phone number format. Phone number must be at least 12 digits long (e.g., 260956587842).",
		}, nil
	}

	// Check if phone prefix is supported by any provider
	provider := GetProviderFromPhoneNumber(req.PhoneNumber)
	if strings.HasPrefix(provider, "Unsupported Provider") {
		return &NameLookupResponse{
			Code:    400,
			Status:  "error",
			Message: provider,
		}, nil
	}

	log.Printf(
		"[HandleNameLookup] Initiating name lookup for phone: %s, provider: %s, client_id: %s, merchant_id: %s",
		req.PhoneNumber,
		provider,
		req.ClientID,
		merchant.MerchantID.String(),
	)

	lookupPayload := map[string]interface{}{
		"phone_number": req.PhoneNumber,
	}

	responseBytes, err := app.MQ.Request(
		"mno.namelookup.requests",
		lookupPayload,
	)

	if err != nil {
		log.Printf("[HandleNameLookup] RPC request failed: %v", err)

		return &NameLookupResponse{
			Code:    500,
			Status:  "error",
			Message: "Failed to contact MNO service: " + err.Error(),
		}, nil
	}

	var mnoData NameLookupData

	if err := json.Unmarshal(responseBytes, &mnoData); err != nil {
		log.Printf("[HandleNameLookup] Failed to unmarshal MNO response: %v", err)

		return &NameLookupResponse{
			Code:    500,
			Status:  "error",
			Message: "Failed to parse MNO service response",
		}, nil
	}

	log.Printf(
		"[HandleNameLookup] Name lookup completed. Status: %s, Provider: %s, Message: %s",
		mnoData.Status,
		mnoData.Provider,
		mnoData.Message,
	)

	respCode := 200

	if mnoData.Status == "failed" || mnoData.Status == "error" {
		respCode = 400
	}

	resp := NameLookupResponse{
		Code:    respCode,
		Status:  mnoData.Status,
		Message: mnoData.Message,
		Data:    &mnoData,
	}

	utils.LogAuditEvent(app, "merchant_service", "namelookup.completed", map[string]interface{}{
		"client_id":   req.ClientID,
		"merchant_id": merchant.MerchantID.String(),
		"provider":    mnoData.Provider,
		"status":      mnoData.Status,
		"message":     mnoData.Message,
	})

	return &resp, nil
}
