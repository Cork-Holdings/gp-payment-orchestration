package merchantapis

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"strings"

	"github.com/Cork-Holdings/gp_payment_orchestration/internal/global"
	"github.com/Cork-Holdings/gp_payment_orchestration/internal/proto/batch_name_lookup_proto"
	"github.com/Cork-Holdings/gp_payment_orchestration/utils"
	"github.com/gin-gonic/gin"
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

func BatchNameLookup(c *gin.Context, app *global.App, req *batch_name_lookup_proto.BatchNameLookupRequest) (*batch_name_lookup_proto.BatchNameLookupResponse, error) {

	fileHeader, err := c.FormFile("file")
	if err != nil {
		return nil, fmt.Errorf("failed to get uploaded file: %w", err)
	}

	// Open uploaded file
	file, err := fileHeader.Open()
	if err != nil {
		return nil, fmt.Errorf("failed to open uploaded file: %w", err)
	}
	defer file.Close()

	reader := csv.NewReader(file)
	reader.FieldsPerRecord = -1

	// Read header
	header, err := reader.Read()
	if err != nil {
		return nil, fmt.Errorf("failed to read CSV header: %w", err)
	}

	// Clean header (remove BOM and extra spaces)
	for i, col := range header {
		// Remove UTF-8 BOM if present
		col = strings.TrimPrefix(col, "\ufeff")
		header[i] = strings.TrimSpace(col)
	}

	log.Printf("[BatchNameLookup] Cleaned CSV header: %v", header)

	phoneColIdx := -1
	for i, col := range header {
		colLower := strings.ToLower(col)
		if colLower == "phone_number" || colLower == "phone number" || colLower == "phonenumber" || colLower == "mobile" || colLower == "msisdn" || colLower == "phone" {
			phoneColIdx = i
			break
		}
	}

	if phoneColIdx == -1 {
		log.Printf("[BatchNameLookup] Phone column not found. Available columns: %v", header)
		return nil, fmt.Errorf("phone number column not found in CSV")
	}

	var results []*batch_name_lookup_proto.BatchNameLookupData

	for {
		record, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Printf("[BatchNameLookup] Error reading CSV record: %v", err)
			continue
		}

		if len(record) <= phoneColIdx {
			continue
		}

		phoneNumber := strings.TrimSpace(record[phoneColIdx])
		if phoneNumber == "" {
			continue
		}

		// Validate phone number: must be exactly 12 digits
		isDigits := true
		for _, char := range phoneNumber {
			if char < '0' || char > '9' {
				isDigits = false
				break
			}
		}

		if len(phoneNumber) != 12 || !isDigits {
			log.Printf("[BatchNameLookup] Invalid phone number format: %s", phoneNumber)
			continue
		}

		provider := GetProviderFromPhoneNumber(phoneNumber)
		if strings.HasPrefix(provider, "Unsupported Provider") {
			log.Printf("[BatchNameLookup] Unsupported provider for: %s", phoneNumber)
			continue
		}

		// Request name lookup from MNO service
		lookupPayload := map[string]interface{}{
			"phone_number": phoneNumber,
		}

		responseBytes, err := app.MQ.Request(
			"mno.namelookup.requests",
			lookupPayload,
		)

		if err != nil {
			log.Printf("[BatchNameLookup] MQ request failed for %s: %v", phoneNumber, err)
			continue
		}

		var mnoData NameLookupData
		if err := json.Unmarshal(responseBytes, &mnoData); err != nil {
			log.Printf("[BatchNameLookup] Failed to unmarshal MNO response for %s: %v", phoneNumber, err)
			continue
		}

		if mnoData.Status == "success" {
			results = append(results, &batch_name_lookup_proto.BatchNameLookupData{
				PhoneNumber: phoneNumber,
				Name:        mnoData.Names,
				Provider:    mnoData.Provider,
			})
		}
	}

	return &batch_name_lookup_proto.BatchNameLookupResponse{
		Data:        results,
		TotalPages:  1,
		CurrentPage: 1,
		HasMore:     false,
	}, nil
}
