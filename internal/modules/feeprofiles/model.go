package feeprofiles

import (
	"time"

	"github.com/Cork-Holdings/gp_payment_orchestration/internal/common"
	"github.com/Cork-Holdings/gp_payment_orchestration/internal/modules/paymentchannels"
	"github.com/Cork-Holdings/gp_payment_orchestration/internal/modules/transactiontypes"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type FeeProfile struct {
	ID                   uuid.UUID                           `gorm:"type:uuid;primary_key"`
	Name                 string                              `gorm:"type:varchar(255);not null"`
	Code                 string                              `gorm:"type:varchar(255);not null"`
	PaymentChannelID     uuid.UUID                           `gorm:"type:uuid;not null"`
	PaymentChannel       paymentchannels.PaymentChannel      `gorm:"foreignKey:PaymentChannelID"`
	TransactionTypeID    uuid.UUID                           `gorm:"type:uuid;not null"`
	TransactionType      transactiontypes.TransactionType    `gorm:"foreignKey:TransactionTypeID"`
	SubTransactionTypeID uuid.UUID                           `gorm:"type:uuid;default:null"`
	SubTransactionType   transactiontypes.SubTransactionType `gorm:"foreignKey:SubTransactionTypeID"`
	Status               string                              `gorm:"type:varchar(255);default:inactive"`
	ChargeAmount         float64                             `gorm:"type:decimal(10,2);not null"`
	ApprovalStatus       string                              `gorm:"type:varchar(255);default:pending"`
	ApprovedBy           uuid.UUID                           `gorm:"type:uuid;default:null"`
	ApprovedAt           *time.Time                          `gorm:"type:timestamp;default:null"`
	RejectedBy           uuid.UUID                           `gorm:"type:uuid;default:null"`
	RejectedAt           *time.Time                          `gorm:"type:timestamp;default:null"`
	RejectedReason       string                              `gorm:"type:varchar(255);default:null"`
	CalculationMode      string                              `gorm:"type:varchar(255);not null"`
	ChargeType           string                              `gorm:"type:varchar(255);default:percentage"`
	MinimumFee           float64                             `gorm:"type:decimal(10,2);default:0"`
	common.Entity
}

type ProfileFeeBands struct {
	ID           uuid.UUID  `gorm:"type:uuid;primary_key"`
	FeeProfileID uuid.UUID  `gorm:"type:uuid;not null"`
	FeeProfile   FeeProfile `gorm:"foreignKey:FeeProfileID"`
	MinAmount    float64    `gorm:"type:decimal(10,2);not null"`
	MaxAmount    float64    `gorm:"type:decimal(10,2);not null"`
	ChargeAmount float64    `gorm:"type:decimal(10,2);not null"`
	ChargeType   string     `gorm:"type:varchar(255);not null"`
	Status       string     `gorm:"type:varchar(255);default:inactive"`
	common.Entity
}

func (FeeProfile) TableName() string {
	return "fee_profiles"
}

func (ProfileFeeBands) TableName() string {
	return "profile_fee_bands"
}

func (f FeeProfile) Permissions() map[string][]string {
	return map[string][]string{
		"admin":    {"read", "write", "delete"},
		"merchant": {"read"},
	}
}

func (p ProfileFeeBands) Permissions() map[string][]string {
	return map[string][]string{
		"admin":    {"read", "write", "delete"},
		"merchant": {"read"},
	}
}

func (feeProfile *FeeProfile) AutoMigrate(db *gorm.DB) error {
	return db.AutoMigrate(&FeeProfile{})
}

func (p *ProfileFeeBands) AutoMigrate(db *gorm.DB) error {
	return db.AutoMigrate(&ProfileFeeBands{})
}
