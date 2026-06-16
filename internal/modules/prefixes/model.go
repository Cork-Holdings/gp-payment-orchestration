package prefixes

import (
	"github.com/Cork-Holdings/gp_payment_orchestration/internal/common"
	"github.com/Cork-Holdings/gp_payment_orchestration/internal/modules/paymentchannels"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Prefix struct {
	ID     uuid.UUID `gorm:"type:uuid;primary_key"`
	Prefix string    `gorm:"type:varchar(255);not null"`
	common.Entity
}

type PrefixPaymentChannel struct {
	ID               uuid.UUID                      `gorm:"type:uuid;primary_key"`
	PrefixID         uuid.UUID                      `gorm:"type:uuid;not null"`
	PaymentChannelID uuid.UUID                      `gorm:"type:uuid;not null"`
	PaymentChannel   paymentchannels.PaymentChannel `gorm:"foreignKey:PaymentChannelID"`
	Prefix           Prefix                         `gorm:"foreignKey:PrefixID"`
	common.Entity
}

func (Prefix) TableName() string {
	return "prefixes"
}

func (PrefixPaymentChannel) TableName() string {
	return "prefix_payment_channels"
}

func (p Prefix) Permissions() map[string][]string {
	return map[string][]string{
		"admin":    {"read", "write", "delete"},
		"merchant": {"read"},
	}
}

func (p PrefixPaymentChannel) Permissions() map[string][]string {
	return map[string][]string{
		"admin":    {"read", "write", "delete"},
		"merchant": {"read"},
	}
}

func (p *Prefix) AutoMigrate(db *gorm.DB) error {
	return db.AutoMigrate(&Prefix{})
}

func (p *PrefixPaymentChannel) AutoMigrate(db *gorm.DB) error {
	return db.AutoMigrate(&PrefixPaymentChannel{})
}
