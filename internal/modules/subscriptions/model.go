package subscriptions

import (
	"github.com/Cork-Holdings/gp_payment_orchestration/internal/common"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Subscription struct {
	ID   uuid.UUID `gorm:"type:uuid;primary_key"`
	Name string    `gorm:"type:varchar(255);not null"`
	// Code        string    `gorm:"type:varchar(255);not null"`
	Status      string `gorm:"type:varchar(255);not null"`
	Description string `gorm:"type:varchar(255);not null"`
	common.Entity
}

type MerchantSubscription struct {
	ID             uuid.UUID    `gorm:"type:uuid;primary_key"`
	MerchantID     uuid.UUID    `gorm:"type:uuid;not null"`
	SubscriptionID uuid.UUID    `gorm:"type:uuid;not null"`
	Subscription   Subscription `gorm:"foreignKey:SubscriptionID"`
	Status         string       `gorm:"type:varchar(255);not null"`
	common.Entity
}

func (s Subscription) TableName() string {
	return "subscriptions"
}

func (m MerchantSubscription) TableName() string {
	return "merchant_subscriptions"
}

func (s Subscription) Permissions() map[string][]string {
	return map[string][]string{
		"admin":    {"read", "write", "delete"},
		"merchant": {"read"},
	}
}

func (m MerchantSubscription) Permissions() map[string][]string {
	return map[string][]string{
		"admin":    {"read", "write", "delete"},
		"merchant": {"read"},
	}
}

func (s *Subscription) AutoMigrate(db *gorm.DB) error {
	return db.AutoMigrate(&Subscription{})
}

func (m *MerchantSubscription) AutoMigrate(db *gorm.DB) error {
	return db.AutoMigrate(&MerchantSubscription{})
}
