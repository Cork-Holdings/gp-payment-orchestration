package paymentservices

import (
	"github.com/Cork-Holdings/gp_payment_orchestration/internal/common"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type PaymentService struct {
	ID     uuid.UUID `gorm:"type:uuid;primary_key"`
	Name   string    `gorm:"type:varchar(255);not null"`
	Status string    `gorm:"type:varchar(255);not null"`
	Logo   string    `gorm:"type:varchar(255);not null"`
	common.Entity
}

func (PaymentService) TableName() string {
	return "payment_services"
}

func (p PaymentService) Permissions() map[string][]string {
	return map[string][]string{
		"admin":    {"read", "write", "delete"},
		"merchant": {"read"},
	}
}

func (p *PaymentService) AutoMigrate(db *gorm.DB) error {
	return db.AutoMigrate(&PaymentService{})
}
