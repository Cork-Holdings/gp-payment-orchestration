package common

import "github.com/golang-jwt/jwt/v5"

type Claims struct {
	jwt.RegisteredClaims

	// Present in all token types
	ID        string `json:"id"`
	Email     string `json:"email,omitempty"`
	Type      string `json:"type,omitempty"` // "merchant" | "admin" | ""
	SessionID string `json:"session_id,omitempty"`

	// User / merchant tokens
	FirstName      string `json:"first_name,omitempty"`
	LastName       string `json:"last_name,omitempty"`
	IsStaff        bool   `json:"is_staff,omitempty"`
	UserStatus     string `json:"user_status,omitempty"`
	MerchantID     string `json:"merchant_id,omitempty"`
	MerchantActive bool   `json:"merchant_active,omitempty"`

	// Admin tokens
	IsSuperAdmin bool   `json:"is_super_admin,omitempty"`
	AdminStatus  string `json:"admin_status,omitempty"`
}
