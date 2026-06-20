package merchantips

import (
	"math"
	"time"

	"github.com/Cork-Holdings/gp_payment_orchestration/internal/global"
	"github.com/Cork-Holdings/gp_payment_orchestration/internal/proto/merchant_ips_proto"
	"github.com/google/uuid"
)

func CreateMerchantIP(req *merchant_ips_proto.CreateMerchantIPRequest) error {
	merchantID, err := uuid.Parse(req.MerchantId)
	if err != nil {
		return err
	}
	submittedBy, err := uuid.Parse(req.SubmittedBy)
	if err != nil {
		return err
	}
	merchantIP := MerchantIP{
		ID:          uuid.New(),
		MerchantID:  merchantID,
		IPAddress:   req.IpAddress,
		Status:      req.Status,
		SubmittedBy: submittedBy,
	}

	tx := global.GetDB().Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
			panic(r)
		}
	}()

	if err := tx.Model(&MerchantIP{}).Create(&merchantIP).Error; err != nil {
		tx.Rollback()
		return err
	}

	if err := tx.Commit().Error; err != nil {
		tx.Rollback()
		return err
	}
	return nil
}

func GetMerchantIP(id string) (*MerchantIP, error) {
	merchantIP := MerchantIP{}
	if err := global.GetDB().Model(&MerchantIP{}).Where("id = ?", id).First(&merchantIP).Error; err != nil {
		return nil, err
	}
	return &merchantIP, nil
}

func GetMerchantIPs(req *merchant_ips_proto.GetMerchantIPsRequest) (*merchant_ips_proto.GetMerchantIPsResponse, error) {

	var merchantIPs []MerchantIP

	page := req.Page
	pageSize := req.PageSize
	searchQuery := req.SearchQuery

	limit := uint(pageSize)
	offset := uint((page - 1) * pageSize)

	query := global.GetDB().Model(&MerchantIP{})

	if searchQuery != "" {
		if _, err := uuid.Parse(searchQuery); err == nil {
			query = query.Where("merchant_id = ?", req.MerchantId)
		}
	}

	total := int64(0)
	if err := query.Count(&total).Error; err != nil {
		return nil, err
	}

	if err := query.Offset(int(offset)).Limit(int(limit)).Find(&merchantIPs).Error; err != nil {
		return nil, err
	}
	totalPages := int32(math.Ceil(float64(total) / float64(pageSize)))

	var merchantRes []*merchant_ips_proto.MerchantIP
	for _, merchantIP := range merchantIPs {
		merchantRes = append(merchantRes, &merchant_ips_proto.MerchantIP{
			Id:             merchantIP.ID.String(),
			MerchantId:     merchantIP.MerchantID.String(),
			IpAddress:      merchantIP.IPAddress,
			Status:         merchantIP.Status,
			SubmittedBy:    merchantIP.SubmittedBy.String(),
			ApprovedBy:     merchantIP.ApprovedBy.String(),
			ApprovedAt:     merchantIP.ApprovedAt.Format(time.RFC3339),
			RejectedBy:     merchantIP.RejectedBy.String(),
			RejectedAt:     merchantIP.RejectedAt.Format(time.RFC3339),
			RejectedReason: merchantIP.RejectedReason,
		})
	}

	return &merchant_ips_proto.GetMerchantIPsResponse{
		MerchantIps: merchantRes,
		TotalPages:  totalPages,
		CurrentPage: page,
		HasMore:     page < totalPages,
	}, nil
}

func UpdateMerchantIP(req *merchant_ips_proto.EditMerchantIPRequest) error {

	updates := map[string]interface{}{}
	if req.IpAddress != "" {
		updates["ip_address"] = req.IpAddress
	}
	if req.Status != "" {
		updates["status"] = req.Status
	}
	if req.ApprovedBy != "" {
		updates["approved_by"] = uuid.MustParse(req.ApprovedBy)
	}
	if req.ApprovedAt != "" {
		approvedAt, err := time.Parse(time.RFC3339, req.ApprovedAt)
		if err != nil {
			return err
		}
		updates["approved_at"] = approvedAt
	}
	if req.RejectedBy != "" {
		updates["rejected_by"] = uuid.MustParse(req.RejectedBy)
	}
	if req.RejectedAt != "" {
		rejectedAt, err := time.Parse(time.RFC3339, req.RejectedAt)
		if err != nil {
			return err
		}
		updates["rejected_at"] = rejectedAt
	}
	if req.RejectedReason != "" {
		updates["rejected_reason"] = req.RejectedReason
	}
	tx := global.GetDB().Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
			panic(r)
		}
	}()

	if err := tx.Model(&MerchantIP{}).Where("id = ?", req.Id).Updates(updates).Error; err != nil {
		tx.Rollback()
		return err
	}

	if err := tx.Commit().Error; err != nil {
		tx.Rollback()
		return err
	}

	return nil
}

func DeleteMerchantIP(id string) error {
	tx := global.GetDB().Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
			panic(r)
		}
	}()

	if err := tx.Model(&MerchantIP{}).Where("id = ?", id).Delete(&MerchantIP{}).Error; err != nil {
		tx.Rollback()
		return err
	}

	if err := tx.Commit().Error; err != nil {
		tx.Rollback()
		return err
	}
	return nil
}
