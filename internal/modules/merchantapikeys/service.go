package merchantapikeys

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"os"

	"github.com/Cork-Holdings/gp_payment_orchestration/internal/common"
	"github.com/Cork-Holdings/gp_payment_orchestration/internal/global"
	"github.com/Cork-Holdings/gp_payment_orchestration/internal/proto/merchant_api_keys_proto"
	"github.com/Cork-Holdings/gp_payment_orchestration/utils"
	"github.com/google/uuid"
)

func CreateMerchantKeys(merchantID string) (*MerchantAPIKey, error) {

	//Parse the merchant ID to ensure it's a valid UUID
	if _, err := uuid.Parse(merchantID); err != nil {
		return nil, fmt.Errorf("invalid merchant ID: %v", err)
	}

	//Generate a random client ID and client secret
	clientID := common.GenerateRandomString(32)
	clientSecret := common.GenerateRandomString(32)

	merchantAPIKey := MerchantAPIKey{
		ID:           uuid.New(),
		MerchantID:   uuid.MustParse(merchantID),
		ClientID:     clientID,
		ClientSecret: clientSecret,
		Status:       "active",
	}

	tx := global.GetDB().Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
			panic(r)
		}
	}()

	if err := tx.Model(&MerchantAPIKey{}).Create(&merchantAPIKey).Error; err != nil {
		tx.Rollback()
		return nil, err
	}

	// Request account provisioning before committing transaction
	responseBytes, err := global.GetMQ().Request("merchant.accounts.create", map[string]any{
		"merchant_id": merchantID,
	})
	if err != nil {
		tx.Rollback()
		return nil, fmt.Errorf("account provisioning failed: %w", err)
	}

	var rpcResp struct {
		Code    int    `json:"code"`
		Status  string `json:"status"`
		Message string `json:"message"`
	}
	if err := json.Unmarshal(responseBytes, &rpcResp); err != nil {
		tx.Rollback()
		return nil, fmt.Errorf("invalid provisioning response: %w", err)
	}

	if rpcResp.Code != 200 {
		tx.Rollback()
		if rpcResp.Message == "" {
			rpcResp.Message = "unknown provisioning error"
		}
		return nil, fmt.Errorf("account provisioning failed: %s", rpcResp.Message)
	}

	// Commit transaction only after account provisioning succeeds
	if err := tx.Commit().Error; err != nil {
		tx.Rollback()
		return nil, err
	}

	// Emit event to RabbitMQ after successful commit
	eventPayload := map[string]interface{}{
		"merchant_id": merchantID,
		"client_id":   clientID,
		"status":      "active",
	}
	global.GetMQ().Emit("merchant.keys.created", eventPayload)

	return &merchantAPIKey, nil
}

func GetMerchantAPIKey(id string) (*MerchantAPIKey, error) {
	merchantAPIKey := MerchantAPIKey{}
	if err := global.GetDB().Model(&MerchantAPIKey{}).Where("id = ?", id).First(&merchantAPIKey).Error; err != nil {
		return nil, err
	}
	return &merchantAPIKey, nil
}

func GetMerchantAPIKeys(req *merchant_api_keys_proto.GetMerchantAPIKeysRequest) (*merchant_api_keys_proto.GetMerchantAPIKeysResponse, error) {

	var merchantAPIKeys []MerchantAPIKey

	page := req.Page
	pageSize := req.PageSize
	merchantID := req.MerchantId

	limit := uint(pageSize)
	offset := uint((page - 1) * pageSize)

	query := global.GetDB().Model(&MerchantAPIKey{})

	if merchantID != "" {
		query = query.Where("merchant_id = ?", merchantID)
	}

	total := int64(0)
	if err := query.Count(&total).Error; err != nil {
		return nil, err
	}

	if err := query.Offset(int(offset)).Limit(int(limit)).Find(&merchantAPIKeys).Error; err != nil {
		return nil, err
	}
	totalPages := int32(math.Ceil(float64(total) / float64(pageSize)))

	var merchantRes []*merchant_api_keys_proto.MerchantAPIKey
	for _, merchantAPIKey := range merchantAPIKeys {
		merchantRes = append(merchantRes, &merchant_api_keys_proto.MerchantAPIKey{
			Id:            merchantAPIKey.ID.String(),
			MerchantId:    merchantAPIKey.MerchantID.String(),
			ClientId:      merchantAPIKey.ClientID,
			ClientSecret:  merchantAPIKey.ClientSecret,
			Pin:           merchantAPIKey.Pin,
			AuthSignature: merchantAPIKey.AuthSignature,
			Status:        merchantAPIKey.Status,
		})
	}

	return &merchant_api_keys_proto.GetMerchantAPIKeysResponse{
		MerchantApiKeys: merchantRes,
		TotalPages:      totalPages,
		CurrentPage:     page,
		HasMore:         page < totalPages,
	}, nil
}

func UpdateMerchantAPIKey(req *merchant_api_keys_proto.EditMerchantAPIKeyRequest) error {

	updates := map[string]interface{}{}
	if req.ClientSecret != "" {
		updates["client_secret"] = req.ClientSecret
	}
	if req.Pin != "" {
		updates["pin"] = req.Pin
	}
	if req.AuthSignature != "" {
		updates["auth_signature"] = req.AuthSignature
	}
	if req.Status != "" {
		updates["status"] = req.Status
	}
	tx := global.GetDB().Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
			panic(r)
		}
	}()

	if err := tx.Model(&MerchantAPIKey{}).Where("id = ?", req.Id).Updates(updates).Error; err != nil {
		tx.Rollback()
		return err
	}

	if err := tx.Commit().Error; err != nil {
		tx.Rollback()
		return err
	}

	return nil
}

func DeleteMerchantAPIKey(id string) error {
	tx := global.GetDB().Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
			panic(r)
		}
	}()

	if err := tx.Model(&MerchantAPIKey{}).Where("id = ?", id).Delete(&MerchantAPIKey{}).Error; err != nil {
		tx.Rollback()
		return err
	}

	if err := tx.Commit().Error; err != nil {
		tx.Rollback()
		return err
	}
	return nil
}

func GenerateAuthSignature(req *merchant_api_keys_proto.GenerateAuthSignatureRequest) (string, error) {

	merchantID, err := uuid.Parse(req.MerchantId)
	if err != nil {
		return "", err
	}

	//Check if merchant has any API keys
	var merchantAPIKeys []MerchantAPIKey
	if err := global.GetDB().Model(&MerchantAPIKey{}).Where("merchant_id = ?", merchantID).Find(&merchantAPIKeys).Error; err != nil {
		return "", err
	}

	if len(merchantAPIKeys) == 0 {
		return "", errors.New("merchant has no API keys")
	}

	//If Pin is not set update it with the pin from the request
	if merchantAPIKeys[0].Pin == "" {

		encryptionKey := []byte(os.Getenv("ENCRYPTION_KEY"))
		if len(encryptionKey) == 0 {
			return "", errors.New("ENCRYPTION_KEY not set")
		}

		// Encrypt the Pin
		ePIN, err := utils.Encrypt(req.Pin, encryptionKey)
		if err != nil {
			return "", err
		}

		merchantAPIKeys[0].Pin = ePIN
		if err := global.GetDB().Model(&MerchantAPIKey{}).Where("id = ?", merchantAPIKeys[0].ID).Updates(map[string]interface{}{"pin": ePIN}).Error; err != nil {
			return "", err
		}
	}

	//Decrypt the Pin from the database
	encryptionKey := []byte(os.Getenv("ENCRYPTION_KEY"))
	if len(encryptionKey) == 0 {
		return "", errors.New("ENCRYPTION_KEY not set")
	}

	decryptedPIN, err := utils.Decrypt(merchantAPIKeys[0].Pin, encryptionKey)
	if err != nil {
		return "", err
	}

	//If Pin exists check if it is correct
	if merchantAPIKeys[0].Pin != "" {
		if decryptedPIN != req.Pin {
			return "", errors.New("invalid pin")
		}
	}

	//Generate the auth signature

	clientID := merchantAPIKeys[0].ClientID
	clientSecret := merchantAPIKeys[0].ClientSecret
	pin := merchantAPIKeys[0].Pin

	message := fmt.Sprintf("%s:%s", clientID, pin)

	signature := hmac.New(sha256.New, []byte(clientSecret))
	signature.Write([]byte(message))

	signatureString := hex.EncodeToString(signature.Sum(nil))

	//Update the auth signature in the database
	if err := global.GetDB().Model(&MerchantAPIKey{}).Where("id = ?", merchantAPIKeys[0].ID).Updates(map[string]interface{}{"auth_signature": signatureString}).Error; err != nil {
		return "", err
	}

	return signatureString, nil
}

func SetPin(req *merchant_api_keys_proto.SetPinRequest) error {

	parsedMerchantID, err := uuid.Parse(req.MerchantId)
	if err != nil {
		return err
	}

	// Encrypt the Pin
	encryptionKey := []byte(os.Getenv("ENCRYPTION_KEY"))
	if len(encryptionKey) == 0 {
		return errors.New("ENCRYPTION_KEY not set")
	}

	//Ensure Pin is not empty
	if req.Pin == "" {
		return errors.New("pin cannot be empty")
	}

	//Ensure pin is at least 8 characters long
	if len(req.Pin) < 8 {
		return errors.New("pin must be at least 8 characters long")
	}

	ePIN, err := utils.Encrypt(req.Pin, encryptionKey)
	if err != nil {
		return err
	}

	tx := global.GetDB().Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
			panic(r)
		}
	}()

	if err := global.GetDB().Model(&MerchantAPIKey{}).Where("merchant_id = ?", parsedMerchantID).Updates(map[string]interface{}{"pin": ePIN}).Error; err != nil {
		tx.Rollback()
		return err
	}

	if err := tx.Commit().Error; err != nil {
		tx.Rollback()
		return err
	}

	return nil
}
