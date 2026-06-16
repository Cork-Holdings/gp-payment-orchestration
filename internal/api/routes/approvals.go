package routes

import (
	"errors"
	"net/http"

	"github.com/Cork-Holdings/gp_payment_orchestration/internal/global"
	"github.com/Cork-Holdings/gp_payment_orchestration/internal/modules/approvals"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func RegisterApprovalRoutes(e *gin.Engine, app *global.App) {
	e.POST("/approvals/:id/approve", func(c *gin.Context) {
		id := c.Param("id")
		var req struct {
			ApproverID string `json:"approver_id" binding:"required"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		var approval approvals.ApprovalRequest
		err := app.DB.Where("ext_id = ?", id).First(&approval).Error
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				c.JSON(http.StatusNotFound, gin.H{"error": "approval request not found"})
				return
			}
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		if approval.Status != "PENDING" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "approval request is already processed"})
			return
		}

		// Separation of Duties Enforcement: Initiator cannot approve!
		if approval.InitiatorID == req.ApproverID {
			c.JSON(http.StatusForbidden, gin.H{"error": "FORBIDDEN", "message": "Separation of duties violation: Creator of approval request is not authorized to approve it."})
			return
		}

		approval.ApproverID = req.ApproverID
		approval.Status = "APPROVED"
		
		err = app.DB.Save(&approval).Error
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{"status": "APPROVED", "approval_id": approval.ExtID})
	})

	e.POST("/approvals/:id/reject", func(c *gin.Context) {
		id := c.Param("id")
		var req struct {
			ApproverID string `json:"approver_id" binding:"required"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		var approval approvals.ApprovalRequest
		err := app.DB.Where("ext_id = ?", id).First(&approval).Error
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				c.JSON(http.StatusNotFound, gin.H{"error": "approval request not found"})
				return
			}
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		if approval.Status != "PENDING" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "approval request is already processed"})
			return
		}

		// Separation of Duties Enforcement: Initiator cannot reject
		if approval.InitiatorID == req.ApproverID {
			c.JSON(http.StatusForbidden, gin.H{"error": "FORBIDDEN", "message": "Separation of duties violation: Creator of approval request is not authorized to reject it."})
			return
		}

		approval.ApproverID = req.ApproverID
		approval.Status = "REJECTED"

		err = app.DB.Save(&approval).Error
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{"status": "REJECTED", "approval_id": approval.ExtID})
	})
}
