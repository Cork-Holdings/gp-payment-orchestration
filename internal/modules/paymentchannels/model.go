package paymentchannels

import (
	"github.com/Cork-Holdings/gp_payment_orchestration/internal/common"
	"github.com/Cork-Holdings/gp_payment_orchestration/internal/modules/providers"
	"github.com/Cork-Holdings/gp_payment_orchestration/internal/modules/subscriptions"
	"github.com/Cork-Holdings/gp_payment_orchestration/internal/modules/transactiontypes"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type PaymentChannel struct {
	ID                   uuid.UUID                           `gorm:"type:uuid;primary_key"`
	Name                 string                              `gorm:"type:varchar(255);not null"`
	Code                 string                              `gorm:"type:varchar(255);not null"`
	Logo                 string                              `gorm:"type:varchar(255);default:null"`
	Status               string                              `gorm:"type:varchar(255);default:inactive"`
	ProviderID           uuid.UUID                           `gorm:"type:uuid;not null"`
	Provider             providers.Provider                  `gorm:"foreignKey:ProviderID"`
	SubscriptionID       uuid.UUID                           `gorm:"type:uuid;not null"`
	Subscription         subscriptions.Subscription          `gorm:"foreignKey:SubscriptionID"`
	TransactionTypeID    uuid.UUID                           `gorm:"type:uuid;not null"`
	TransactionType      transactiontypes.TransactionType    `gorm:"foreignKey:TransactionTypeID"`
	SubTransactionTypeID uuid.UUID                           `gorm:"type:uuid;default:null"`
	SubTransactionType   transactiontypes.SubTransactionType `gorm:"foreignKey:SubTransactionTypeID"`
	FeeType              string                              `gorm:"type:varchar(255);not null"`
	ProviderFee          string                              `gorm:"type:varchar(255);not null"`
	common.Entity
}

type ChannelFeeBands struct {
	ID               uuid.UUID      `gorm:"type:uuid;primary_key"`
	Name             string         `gorm:"type:varchar(255);not null"`
	PaymentChannelID uuid.UUID      `gorm:"type:uuid;not null"`
	PaymentChannel   PaymentChannel `gorm:"foreignKey:PaymentChannelID"`
	MinAmount        float64        `gorm:"type:decimal(10,2);not null"`
	MaxAmount        float64        `gorm:"type:decimal(10,2);not null"`
	ChargeAmount     float64        `gorm:"type:decimal(10,2);not null"`
	ChargeType       string         `gorm:"type:varchar(255);default:percentage"`
	Status           string         `gorm:"type:varchar(255);default:inactive"`
	common.Entity
}

func (PaymentChannel) TableName() string {
	return "payment_channels"
}

func (ChannelFeeBands) TableName() string {
	return "channel_fee_bands"
}

func (p PaymentChannel) Permissions() map[string][]string {
	return map[string][]string{
		"admin":    {"read", "write", "delete"},
		"merchant": {"read"},
	}
}

func (c ChannelFeeBands) Permissions() map[string][]string {
	return map[string][]string{
		"admin":    {"read", "write", "delete"},
		"merchant": {"read"},
	}
}

func (p *PaymentChannel) AutoMigrate(db *gorm.DB) error {
	return db.AutoMigrate(&PaymentChannel{})
}

func (c *ChannelFeeBands) AutoMigrate(db *gorm.DB) error {
	return db.AutoMigrate(&ChannelFeeBands{})
}
