package transactiontypes

import (
	"github.com/Cork-Holdings/gp_payment_orchestration/internal/common"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type TransactionType struct {
	ID        uuid.UUID `gorm:"type:uuid;primary_key"`
	Name      string    `gorm:"type:varchar(255);not null"`
	Code      string    `gorm:"type:varchar(255);not null"`
	MaxAmount string    `gorm:"type:varchar(255);not null"`
	MinAmount string    `gorm:"type:varchar(255);not null"`
	Status    string    `gorm:"type:varchar(255);not null"`
	common.Entity
}

type SubTransactionType struct {
	ID                uuid.UUID       `gorm:"type:uuid;primary_key"`
	Name              string          `gorm:"type:varchar(255);not null"`
	TransactionTypeID uuid.UUID       `gorm:"type:uuid;not null"`
	TransactionType   TransactionType `gorm:"foreignKey:TransactionTypeID"`
	Code              string          `gorm:"type:varchar(255);not null"`
	Status            string          `gorm:"type:varchar(255);not null"`
	MaxAmount         string          `gorm:"type:varchar(255);not null"`
	MinAmount         string          `gorm:"type:varchar(255);not null"`
	common.Entity
}

func (TransactionType) TableName() string {
	return "transaction_types"
}

func (SubTransactionType) TableName() string {
	return "sub_transaction_types"
}

func (t TransactionType) Permissions() map[string][]string {
	return map[string][]string{
		"admin":    {"read", "write", "delete"},
		"merchant": {"read"},
	}
}

func (s SubTransactionType) Permissions() map[string][]string {
	return map[string][]string{
		"admin":    {"read", "write", "delete"},
		"merchant": {"read"},
	}
}

func (t *TransactionType) AutoMigrate(db *gorm.DB) error {
	return db.AutoMigrate(&TransactionType{})
}

func (s *SubTransactionType) AutoMigrate(db *gorm.DB) error {
	return db.AutoMigrate(&SubTransactionType{})
}
