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
	"gorm.io/gorm"
)

var (
	ErrAuthSignaturePinRequired = errors.New("merchant_id and PIN are required to retrieve an auth signature")
	ErrInvalidPin               = errors.New("invalid pin")
)

func CreateMerchantKeys(merchantID string) (*MerchantAPIKey, error) {

	//Parse the merchant ID to ensure it's a valid UUID
	if _, err := uuid.Parse(merchantID); err != nil {
		return nil, fmt.Errorf("invalid merchant ID: %v", err)
	}

	//Generate a random client ID and client secret
	clientID := common.GenerateRandomString(32)
	clientSecret := common.GenerateRandomString(32)

	encryptionKey := []byte(os.Getenv("ENCRYPTION_KEY"))

	if len(encryptionKey) == 0 {
		return nil, errors.New("ENCRYPTION_KEY not set")
	}

	eClientSecret, err := utils.Encrypt(clientSecret, encryptionKey)
	if err != nil {
		return nil, fmt.Errorf("failed to encrypt client secret: %v", err)
	}

	eClientID, err := utils.Encrypt(clientID, encryptionKey)
	if err != nil {
		return nil, fmt.Errorf("failed to encrypt client ID: %v", err)
	}

	merchantAPIKey := MerchantAPIKey{
		ID:           uuid.New(),
		MerchantID:   uuid.MustParse(merchantID),
		ClientID:     eClientID,
		ClientSecret: eClientSecret,
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

	encryptionKey := []byte(os.Getenv("ENCRYPTION_KEY"))
	if len(encryptionKey) == 0 {
		return nil, errors.New("ENCRYPTION_KEY not set")
	}

	var merchantRes []*merchant_api_keys_proto.MerchantAPIKey
	for _, merchantAPIKey := range merchantAPIKeys {

		var decryptedClientID, decryptedClientSecret string

		decryptedClientID, err := utils.Decrypt(merchantAPIKey.ClientID, encryptionKey)
		if err != nil {
			return nil, fmt.Errorf("failed to decrypt client ID: %v", err)
		}

		decryptedClientSecret, err = utils.Decrypt(merchantAPIKey.ClientSecret, encryptionKey)
		if err != nil {
			return nil, fmt.Errorf("failed to decrypt client secret: %v", err)
		}

		authSignature := ""
		//Check if PIN is set
		decryptedPIN := ""
		if merchantAPIKey.Pin != "" {
			decryptedPIN, err = utils.Decrypt(merchantAPIKey.Pin, encryptionKey)
			if err != nil {
				return nil, fmt.Errorf("failed to decrypt pin: %w", err)
			}
		}

		authSignature, err = decryptAuthSignature(merchantAPIKey.AuthSignature, encryptionKey)
		if err != nil {
			return nil, err
		}

		merchantRes = append(merchantRes, &merchant_api_keys_proto.MerchantAPIKey{
			Id:            merchantAPIKey.ID.String(),
			MerchantId:    merchantAPIKey.MerchantID.String(),
			ClientId:      decryptedClientID,
			ClientSecret:  decryptedClientSecret,
			AuthSignature: authSignature,
			Pin:           decryptedPIN,
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

		encryptedSecret, err := encryptSecret(req.ClientSecret)

		if err != nil {
			return err
		}

		updates["client_secret"] = encryptedSecret
	}
	if req.Pin != "" {

		encryptedPin, err := encryptSecret(req.Pin)

		if err != nil {
			return err
		}

		updates["pin"] = encryptedPin
	}
	if req.AuthSignature != "" {
		encryptedSignature, err := encryptSecret(req.AuthSignature)
		if err != nil {
			return err
		}
		updates["auth_signature"] = encryptedSignature
	}
	if req.Status != "" {
		updates["status"] = req.Status
	}
	if len(updates) == 0 {
		return errors.New("no fields to update")
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
		return "", fmt.Errorf("invalid merchant id: %w", err)
	}

	encryptionKey := []byte(os.Getenv("ENCRYPTION_KEY"))
	if len(encryptionKey) == 0 {
		return "", errors.New("ENCRYPTION_KEY not set")
	}

	db := global.GetDB()

	// Get active merchant API key
	var merchantAPIKey MerchantAPIKey

	err = db.
		Where(
			"merchant_id = ? AND status = ?",
			merchantID,
			"active",
		).
		First(&merchantAPIKey).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return "", errors.New("merchant has no active API keys")
		}
		return "", err
	}

	// Handle PIN
	var decryptedPIN string

	if merchantAPIKey.Pin == "" {

		if req.Pin == "" {
			return "", errors.New("pin is required")
		}

		encryptedPin, err := utils.Encrypt(req.Pin, encryptionKey)
		if err != nil {
			return "", fmt.Errorf("failed to encrypt pin: %w", err)
		}

		if err := db.
			Model(&MerchantAPIKey{}).
			Where("id = ?", merchantAPIKey.ID).
			Update("pin", encryptedPin).Error; err != nil {
			return "", err
		}

		decryptedPIN = req.Pin

	} else {

		decryptedPIN, err = utils.Decrypt(
			merchantAPIKey.Pin,
			encryptionKey,
		)

		if err != nil {
			return "", fmt.Errorf("failed to decrypt pin: %w", err)
		}

		if decryptedPIN != req.Pin {
			return "", errors.New("invalid pin")
		}
	}

	// Decrypt client credentials

	clientID, err := utils.Decrypt(
		merchantAPIKey.ClientID,
		encryptionKey,
	)

	if err != nil {
		return "", fmt.Errorf("failed to decrypt client id: %w", err)
	}

	clientSecret, err := utils.Decrypt(
		merchantAPIKey.ClientSecret,
		encryptionKey,
	)

	if err != nil {
		return "", fmt.Errorf("failed to decrypt client secret: %w", err)
	}

	// Generate HMAC signature
	message := fmt.Sprintf(
		"%s:%s",
		clientID,
		decryptedPIN,
	)

	h := hmac.New(
		sha256.New,
		[]byte(clientSecret),
	)

	h.Write([]byte(message))

	signature := hex.EncodeToString(h.Sum(nil))

	encryptedSignature, err := utils.Encrypt(signature, encryptionKey)
	if err != nil {
		return "", fmt.Errorf("failed to encrypt auth signature: %w", err)
	}

	// Save generated signature
	err = db.
		Model(&MerchantAPIKey{}).
		Where("id = ?", merchantAPIKey.ID).
		Update(
			"auth_signature",
			encryptedSignature,
		).Error

	if err != nil {
		return "", fmt.Errorf("failed to save auth signature: %w", err)
	}

	return signature, nil
}
func encryptSecret(value string) (string, error) {

	key := []byte(os.Getenv("ENCRYPTION_KEY"))

	if len(key) == 0 {
		return "", errors.New("ENCRYPTION_KEY not set")
	}

	return utils.Encrypt(value, key)
}

func decryptAuthSignature(value string, key []byte) (string, error) {
	if value == "" {
		return "", nil
	}

	signature, err := utils.Decrypt(value, key)
	if err == nil {
		return signature, nil
	}

	if len(value) == sha256.Size*2 {
		if _, decodeErr := hex.DecodeString(value); decodeErr == nil {
			return value, nil
		}
	}

	return "", fmt.Errorf("failed to decrypt auth signature: %w", err)
}

func SetPin(req *merchant_api_keys_proto.SetPinRequest) error {

	merchantID, err := uuid.Parse(req.MerchantId)
	if err != nil {
		return fmt.Errorf("invalid merchant id: %w", err)
	}

	if req.Pin == "" {
		return errors.New("pin cannot be empty")
	}

	if len(req.Pin) < 8 {
		return errors.New("pin must be at least 8 characters long")
	}

	encryptionKey := []byte(os.Getenv("ENCRYPTION_KEY"))
	if len(encryptionKey) == 0 {
		return errors.New("ENCRYPTION_KEY not set")
	}

	encryptedPIN, err := utils.Encrypt(req.Pin, encryptionKey)
	if err != nil {
		return fmt.Errorf("failed to encrypt pin: %w", err)
	}

	db := global.GetDB()

	tx := db.Begin()

	if tx.Error != nil {
		return tx.Error
	}

	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
			panic(r)
		}
	}()

	var merchantAPIKey MerchantAPIKey

	if err := tx.
		Where(
			"merchant_id = ? AND status = ?",
			merchantID,
			"active",
		).
		First(&merchantAPIKey).Error; err != nil {

		tx.Rollback()

		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("merchant API key not found")
		}

		return err
	}

	if err := tx.
		Model(&MerchantAPIKey{}).
		Where("id = ?", merchantAPIKey.ID).
		Update("pin", encryptedPIN).Error; err != nil {

		tx.Rollback()
		return err
	}

	if err := tx.Commit().Error; err != nil {
		return err
	}

	return nil
}
