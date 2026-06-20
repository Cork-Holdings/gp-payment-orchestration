package providers

import (
	"github.com/Cork-Holdings/gp_payment_orchestration/internal/common"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Provider struct {
	ID     uuid.UUID `gorm:"type:uuid;primary_key"`
	Name   string    `gorm:"type:varchar(255);not null"`
	Code   string    `gorm:"type:varchar(255);not null;unique"`
	Status string    `gorm:"type:varchar(255);default:active"`
	common.Entity
}

func (Provider) TableName() string {
	return "providers"
}

func (p Provider) Permissions() map[string][]string {
	return map[string][]string{
		"admin":    {"read", "write", "delete"},
		"merchant": {"read"},
	}
}

func (p *Provider) AutoMigrate(db *gorm.DB) error {
	return db.AutoMigrate(&Provider{})
}
