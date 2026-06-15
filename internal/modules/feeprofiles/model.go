package feeprofiles

import (
	"time"

	"github.com/Cork-Holdings/gp_payment_orchestration/internal/common"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type FeeProfile struct {
	ID                   uuid.UUID          `gorm:"type:uuid;primary_key"`
	Name                 string             `gorm:"type:varchar(255);not null"`
	Code                 string             `gorm:"type:varchar(255);not null"`
	PaymentChannelID     uuid.UUID          `gorm:"type:uuid;not null"`
	PaymentChannel       PaymentChannel     `gorm:"foreignKey:PaymentChannelID"`
	TransactionTypeID    uuid.UUID          `gorm:"type:uuid;not null"`
	TransactionType      TransactionType    `gorm:"foreignKey:TransactionTypeID"`
	SubTransactionTypeID uuid.UUID          `gorm:"type:uuid;default:null"`
	SubTransactionType   SubTransactionType `gorm:"foreignKey:SubTransactionTypeID"`
	Status               string             `gorm:"type:varchar(255);not null"`
	ChargeAmount         float64            `gorm:"type:decimal(10,2);not null"`
	ApprovalStatus       string             `gorm:"type:varchar(255);not null"`
	ApprovedBy           uuid.UUID          `gorm:"type:uuid;default:null"`
	ApprovedAt           *time.Time         `gorm:"type:timestamp;default:null"`
	RejectedBy           uuid.UUID          `gorm:"type:uuid;default:null"`
	RejectedAt           *time.Time         `gorm:"type:timestamp;default:null"`
	RejectedReason       string             `gorm:"type:varchar(255);default:null"`
	CalculationMode      string             `gorm:"type:varchar(255);not null"`
	common.Entity
}

type PaymentService struct {
	ID     uuid.UUID `gorm:"type:uuid;primary_key"`
	Name   string    `gorm:"type:varchar(255);not null"`
	Status string    `gorm:"type:varchar(255);not null"`
	Logo   string    `gorm:"type:varchar(255);not null"`
	common.Entity
}

type PaymentChannel struct {
	ID                   uuid.UUID          `gorm:"type:uuid;primary_key"`
	Name                 string             `gorm:"type:varchar(255);not null"`
	Code                 string             `gorm:"type:varchar(255);not null"`
	Logo                 string             `gorm:"type:varchar(255);not null"`
	Status               string             `gorm:"type:varchar(255);not null"`
	PaymentServiceID     uuid.UUID          `gorm:"type:uuid;not null"`
	PaymentService       PaymentService     `gorm:"foreignKey:PaymentServiceID"`
	TransactionTypeID    uuid.UUID          `gorm:"type:uuid;not null"`
	TransactionType      TransactionType    `gorm:"foreignKey:TransactionTypeID"`
	SubTransactionTypeID uuid.UUID          `gorm:"type:uuid;default:null"`
	SubTransactionType   SubTransactionType `gorm:"foreignKey:SubTransactionTypeID"`
	FeeType              string             `gorm:"type:varchar(255);not null"`
	ProviderFee          string             `gorm:"type:varchar(255);not null"`
	common.Entity
}

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

type MerchantFeeProfile struct {
	ID           uuid.UUID `gorm:"type:uuid;primary_key"`
	MerchantID   uuid.UUID `gorm:"type:uuid;not null"`
	FeeProfileID uuid.UUID `gorm:"type:uuid;not null"`
	Status       string    `gorm:"type:varchar(255);not null"`
	common.Entity
}

type ChannelFeeBands struct {
	ID               uuid.UUID      `gorm:"type:uuid;primary_key"`
	PaymentChannelID uuid.UUID      `gorm:"type:uuid;not null"`
	PaymentChannel   PaymentChannel `gorm:"foreignKey:PaymentChannelID"`
	MinAmount        float64        `gorm:"type:decimal(10,2);not null"`
	MaxAmount        float64        `gorm:"type:decimal(10,2);not null"`
	ChargeAmount     float64        `gorm:"type:decimal(10,2);not null"`
	Status           string         `gorm:"type:varchar(255);not null"`
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
	Status       string     `gorm:"type:varchar(255);not null"`
	common.Entity
}

type Prefix struct {
	ID     uuid.UUID `gorm:"type:uuid;primary_key"`
	Prefix string    `gorm:"type:varchar(255);not null"`
	common.Entity
}

type PrefixPaymentChannel struct {
	ID               uuid.UUID      `gorm:"type:uuid;primary_key"`
	PrefixID         uuid.UUID      `gorm:"type:uuid;not null"`
	PaymentChannelID uuid.UUID      `gorm:"type:uuid;not null"`
	PaymentChannel   PaymentChannel `gorm:"foreignKey:PaymentChannelID"`
	Prefix           Prefix         `gorm:"foreignKey:PrefixID"`
	common.Entity
}

func (FeeProfile) TableName() string {
	return "fee_profiles"
}

func (PaymentService) TableName() string {
	return "payment_services"
}

func (PaymentChannel) TableName() string {
	return "payment_channels"
}

func (TransactionType) TableName() string {
	return "transaction_types"
}

func (SubTransactionType) TableName() string {
	return "sub_transaction_types"
}

func (MerchantFeeProfile) TableName() string {
	return "merchant_fee_profiles"
}

func (ChannelFeeBands) TableName() string {
	return "channel_fee_bands"
}

func (ProfileFeeBands) TableName() string {
	return "profile_fee_bands"
}

func (Prefix) TableName() string {
	return "prefixes"
}

func (PrefixPaymentChannel) TableName() string {
	return "prefix_payment_channels"
}

func (f FeeProfile) Permissions() map[string][]string {
	return map[string][]string{
		"admin":    {"read", "write", "delete"},
		"merchant": {"read"},
	}
}

func (p PaymentService) Permissions() map[string][]string {
	return map[string][]string{
		"admin":    {"read", "write", "delete"},
		"merchant": {"read"},
	}
}

func (p PaymentChannel) Permissions() map[string][]string {
	return map[string][]string{
		"admin":    {"read", "write", "delete"},
		"merchant": {"read"},
	}
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

func (m MerchantFeeProfile) Permissions() map[string][]string {
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

func (p ProfileFeeBands) Permissions() map[string][]string {
	return map[string][]string{
		"admin":    {"read", "write", "delete"},
		"merchant": {"read"},
	}
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

func (feeProfile *FeeProfile) AutoMigrate(db *gorm.DB) error {
	return db.AutoMigrate(&FeeProfile{}, &PaymentService{}, &PaymentChannel{}, &TransactionType{}, &SubTransactionType{}, &MerchantFeeProfile{}, &ChannelFeeBands{}, &ProfileFeeBands{}, &Prefix{}, &PrefixPaymentChannel{})
}

func (p *PaymentService) AutoMigrate(db *gorm.DB) error {
	return db.AutoMigrate(&PaymentService{})
}

func (p *PaymentChannel) AutoMigrate(db *gorm.DB) error {
	return db.AutoMigrate(&PaymentChannel{})
}

func (t *TransactionType) AutoMigrate(db *gorm.DB) error {
	return db.AutoMigrate(&TransactionType{})
}

func (s *SubTransactionType) AutoMigrate(db *gorm.DB) error {
	return db.AutoMigrate(&SubTransactionType{})
}

func (m *MerchantFeeProfile) AutoMigrate(db *gorm.DB) error {
	return db.AutoMigrate(&MerchantFeeProfile{})
}

func (c *ChannelFeeBands) AutoMigrate(db *gorm.DB) error {
	return db.AutoMigrate(&ChannelFeeBands{})
}

func (p *ProfileFeeBands) AutoMigrate(db *gorm.DB) error {
	return db.AutoMigrate(&ProfileFeeBands{})
}

func (p *Prefix) AutoMigrate(db *gorm.DB) error {
	return db.AutoMigrate(&Prefix{})
}

func (p *PrefixPaymentChannel) AutoMigrate(db *gorm.DB) error {
	return db.AutoMigrate(&PrefixPaymentChannel{})
}
