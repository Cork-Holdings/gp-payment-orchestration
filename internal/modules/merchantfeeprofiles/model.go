package merchantfeeprofiles

import (
	"github.com/Cork-Holdings/gp_payment_orchestration/internal/common"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type MerchantFeeProfile struct {
	ID           uuid.UUID `gorm:"type:uuid;primary_key"`
	MerchantID   uuid.UUID `gorm:"type:uuid;not null"`
	FeeProfileID uuid.UUID `gorm:"type:uuid;not null"`
	Status       string    `gorm:"type:varchar(255);not null"`
	common.Entity
}

func (MerchantFeeProfile) TableName() string {
	return "merchant_fee_profiles"
}

func (m MerchantFeeProfile) Permissions() map[string][]string {
	return map[string][]string{
		"admin":    {"read", "write", "delete"},
		"merchant": {"read"},
	}
}

func (m *MerchantFeeProfile) AutoMigrate(db *gorm.DB) error {
	return db.AutoMigrate(&MerchantFeeProfile{})
}
