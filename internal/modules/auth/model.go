package auth

import (
	"github.com/Cork-Holdings/gp_payment_orchestration/internal/common"
	"gorm.io/gorm"
)

type User struct {
	common.Entity
	Username string `gorm:"column:username;uniqueIndex"`
	Password string `gorm:"column:password"`
	Role     string `gorm:"column:role"`
}

func (u *User) TableName() string {
	return "users"
}

func (u *User) Permissions() map[string][]string {
	return map[string][]string{
		"admin": {"read", "write", "approve"},
		"user":  {"read"},
	}
}

func (u *User) AutoMigrate(db *gorm.DB) error {
	err := db.AutoMigrate(u)
	if err != nil {
		return err
	}
	
	// Seed a default user if database is empty
	var count int64
	db.Model(&User{}).Count(&count)
	if count == 0 {
		defaultUser := &User{
			Username: "admin",
			Password: "password",
			Role:     "admin",
		}
		defaultUser.Autofill(defaultUser)
		db.Create(defaultUser)
	}
	return nil
}
