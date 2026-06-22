package m_api

import (
	"log"
	"os"
	"time"

	"github.com/Cork-Holdings/gp_payment_orchestration/internal/common"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type MerchantProfile struct {
	ID             uuid.UUID `gorm:"type:uuid;primary_key"`
	ClientID       string    `gorm:"type:varchar(255);uniqueIndex;not null"`
	ClientSecret   string    `gorm:"type:varchar(255);not null"`
	MerchantName   string    `gorm:"type:varchar(255);not null"`
	AllowedIPs     string    `gorm:"type:text;not null"` // comma-separated allowed IPs/CIDRs
	WalletBalance  float64   `gorm:"type:decimal(15,2);default:0.0"`
	WalletCurrency string    `gorm:"type:varchar(10);default:'USD'"`
	common.Entity
}

func (MerchantProfile) TableName() string {
	return "merchant_profiles"
}

func (MerchantProfile) Permissions() map[string][]string {
	return map[string][]string{
		"admin":    {"read", "write"},
		"merchant": {"read"},
	}
}

func (m *MerchantProfile) AutoMigrate(db *gorm.DB) error {
	err := db.AutoMigrate(m)
	if err != nil {
		return err
	}

	var count int64
	db.Model(&MerchantProfile{}).Count(&count)
	if count == 0 && os.Getenv("SEED_DEFAULT_MERCHANT") == "true" {
		plainSecret := "secret_456"
		protectedSecret, err := protectSecret(plainSecret)
		if err != nil {
			return err
		}

		now := time.Now()
		defaultMerchant := &MerchantProfile{
			ID:             uuid.New(),
			ClientID:       "merchant_123",
			ClientSecret:   protectedSecret,
			MerchantName:   "Test Merchant",
			AllowedIPs:     "127.0.0.1,::1,10.0.0.0/8,192.168.1.0/24",
			WalletBalance:  1000.0,
			WalletCurrency: "USD",
			Entity: common.Entity{
				CreatedAt: &now,
				UpdatedAt: &now,
			},
		}
		if err := db.Create(defaultMerchant).Error; err != nil {
			return err
		}
		log.Println("Seeded default merchant (SEED_DEFAULT_MERCHANT=true)")
	}
	return nil
}

type MerchantTransaction struct {
	ID          uuid.UUID `gorm:"type:uuid;primary_key"`
	TrackingRef string    `gorm:"type:varchar(255);uniqueIndex;not null"`
	ClientID    string    `gorm:"type:varchar(255);index;not null"`
	Type        string    `gorm:"type:varchar(20);not null"` // "COLLECT" or "DISBURSE"
	PhoneNumber string    `gorm:"type:varchar(20);not null"`
	Amount      float64   `gorm:"type:decimal(15,2);not null"`
	Currency    string    `gorm:"type:varchar(10);not null"`
	Status      string    `gorm:"type:varchar(20);not null"` // "PENDING", "PROCESSING", "COMPLETED", "FAILED"
	common.Entity
}

func (MerchantTransaction) TableName() string {
	return "merchant_transactions"
}

func (MerchantTransaction) Permissions() map[string][]string {
	return map[string][]string{
		"admin":    {"read", "write"},
		"merchant": {"read"},
	}
}

func (t *MerchantTransaction) AutoMigrate(db *gorm.DB) error {
	return db.AutoMigrate(t)
}
