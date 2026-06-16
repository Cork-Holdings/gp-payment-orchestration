package approvals

import (
	"github.com/Cork-Holdings/gp_payment_orchestration/internal/common"
	"gorm.io/gorm"
)

type ApprovalRequest struct {
	common.Entity
	WorkflowName string `gorm:"column:workflow_name"`
	InitiatorID  string `gorm:"column:initiator_id"`
	ApproverID   string `gorm:"column:approver_id"`
	Status       string `gorm:"column:status;default:'PENDING'"` // PENDING, APPROVED, REJECTED
}

func (a *ApprovalRequest) TableName() string {
	return "approval_requests"
}

func (a *ApprovalRequest) Permissions() map[string][]string {
	return map[string][]string{
		"admin": {"read", "write"},
	}
}

func (a *ApprovalRequest) AutoMigrate(db *gorm.DB) error {
	err := db.AutoMigrate(a)
	if err != nil {
		return err
	}
	
	// Seed a mock approval request if table is empty
	var count int64
	db.Model(&ApprovalRequest{}).Count(&count)
	if count == 0 {
		mockApproval := &ApprovalRequest{
			WorkflowName: "merchant_onboarding_payment_enablement",
			InitiatorID:  "user_123", // Initiated by user_123
			Status:       "PENDING",
		}
		mockApproval.Autofill(mockApproval)
		db.Create(mockApproval)
	}
	return nil
}
