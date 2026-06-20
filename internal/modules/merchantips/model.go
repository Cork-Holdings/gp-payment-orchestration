package merchantips

import (
	"time"

	"github.com/Cork-Holdings/gp_payment_orchestration/internal/common"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type MerchantIP struct {
	ID             uuid.UUID  `gorm:"type:uuid;primary_key"`
	MerchantID     uuid.UUID  `gorm:"type:uuid;not null"`
	IPAddress      string     `gorm:"type:varchar(255);not null"`
	Status         string     `gorm:"type:varchar(255);not null"`
	SubmittedBy    uuid.UUID  `gorm:"type:uuid;not null"`
	ApprovedBy     uuid.UUID  `gorm:"type:uuid;default:null"`
	ApprovedAt     *time.Time `gorm:"type:timestamp;default:null"`
	RejectedBy     uuid.UUID  `gorm:"type:uuid;default:null"`
	RejectedAt     *time.Time `gorm:"type:timestamp;default:null"`
	RejectedReason string     `gorm:"type:varchar(255);default:null"`
	common.Entity
}

func (MerchantIP) TableName() string {
	return "merchant_ips"
}

func (m MerchantIP) Permissions() map[string][]string {
	return map[string][]string{
		"admin":    {"read", "write", "delete"},
		"merchant": {"read"},
	}
}

func (m *MerchantIP) AutoMigrate(db *gorm.DB) error {
	return db.AutoMigrate(&MerchantIP{})
}
