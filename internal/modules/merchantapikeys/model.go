package merchantapikeys

import (
	"github.com/Cork-Holdings/gp_payment_orchestration/internal/common"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type MerchantAPIKey struct {
	ID            uuid.UUID `gorm:"type:uuid;primary_key"`
	MerchantID    uuid.UUID `gorm:"type:uuid;unique:not null"`
	ClientID      string    `gorm:"type:varchar(255);not null"`
	ClientSecret  string    `gorm:"type:varchar(255);not null"`
	Pin           string    `gorm:"type:varchar(255);default:null"`
	AuthSignature string    `gorm:"type:varchar(255);default:null"`
	Status        string    `gorm:"type:varchar(255);not null"`
	common.Entity
}

func (MerchantAPIKey) TableName() string {
	return "merchant_api_keys"
}

func (m MerchantAPIKey) Permissions() map[string][]string {
	return map[string][]string{
		"admin":    {"read", "write", "delete"},
		"merchant": {"read"},
	}
}

func (m *MerchantAPIKey) AutoMigrate(db *gorm.DB) error {
	return db.AutoMigrate(&MerchantAPIKey{})
}
