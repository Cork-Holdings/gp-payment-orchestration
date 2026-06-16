package ledger

import (
	"github.com/Cork-Holdings/gp_payment_orchestration/internal/common"
	"gorm.io/gorm"
)

type Account struct {
	common.Entity
	Name       string  `gorm:"column:name"`
	Currency   string  `gorm:"column:currency"`
	Balance    float64 `gorm:"column:balance"`
	Version    int64   `gorm:"column:version;default:1"`
	WalletType string  `gorm:"column:wallet_type;default:'emoney'"`
}

func (a *Account) TableName() string {
	return "accounts"
}

func (a *Account) Permissions() map[string][]string {
	return map[string][]string{
		"admin": {"read", "write"},
		"user":  {"read"},
	}
}

func (a *Account) AutoMigrate(db *gorm.DB) error {
	return db.AutoMigrate(a)
}
