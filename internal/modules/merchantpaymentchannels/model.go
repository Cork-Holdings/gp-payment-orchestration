package merchantpaymentchannels

import (
	"time"

	"github.com/Cork-Holdings/gp_payment_orchestration/internal/common"
	"github.com/Cork-Holdings/gp_payment_orchestration/internal/modules/paymentchannels"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type MerchantPaymentChannel struct {
	ID               uuid.UUID                      `gorm:"type:uuid;primary_key"`
	MerchantID       uuid.UUID                      `gorm:"type:uuid;not null"`
	PaymentChannelID uuid.UUID                      `gorm:"type:uuid;not null"`
	PaymentChannel   paymentchannels.PaymentChannel `gorm:"foreignKey:PaymentChannelID"`
	Status           string                         `gorm:"type:varchar(255);inactive:active"`
	RejectionReason  string                         `gorm:"type:varchar(255);default:null"`
	RejectedBy       uuid.UUID                      `gorm:"type:uuid;default:null"`
	RejectedAt       time.Time                      `gorm:"type:timestamp;default:null"`
	AssignedBy       uuid.UUID                      `gorm:"type:uuid;default:null"`
	AssignedAt       time.Time                      `gorm:"type:timestamp;default:null"`
	ApprovedBy       uuid.UUID                      `gorm:"type:uuid;default:null"`
	ApprovedAt       time.Time                      `gorm:"type:timestamp;default:null"`
	ApprovalStatus   string                         `gorm:"type:varchar(255);default:pending"`
	common.Entity
}

func (MerchantPaymentChannel) TableName() string {
	return "merchant_payment_channels"
}

func (m MerchantPaymentChannel) Permissions() map[string][]string {
	return map[string][]string{
		"admin":    {"read", "write", "delete"},
		"merchant": {"read"},
	}
}

func (m *MerchantPaymentChannel) AutoMigrate(db *gorm.DB) error {
	return db.AutoMigrate(&MerchantPaymentChannel{})
}
