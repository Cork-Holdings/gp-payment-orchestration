package common

import (
	"time"

	"github.com/Cork-Holdings/gp_payment_orchestration/internal/global"

	"gorm.io/gorm"
)

type Entity struct {
	// ExtID     *string    `json:"id" gorm:"column:ext_id;uniqueIndex"`
	CreatedAt *time.Time `json:"created_at" gorm:"column:created_at"`
	UpdatedAt *time.Time `json:"updated_at" gorm:"column:updated_at"`
	DeletedAt *time.Time `json:"-" gorm:"column:deleted_at"`
}

func (e Entity) TableName() string {
	return ""
}

func (e *Entity) Autofill(g global.Model) {
	now := time.Now()

	// e.ExtID = &id
	e.CreatedAt = &now
	e.UpdatedAt = &now
}

func (e Entity) BeforeCreate(tx *gorm.DB) error { return nil }
func (e Entity) AfterCreate(tx *gorm.DB) error  { return nil }
func (e Entity) AfterSave(tx *gorm.DB) error    { return nil }
func (e Entity) BeforeSave(tx *gorm.DB) error   { return nil }
func (e Entity) AfterFind(tx *gorm.DB) error    { return nil }
