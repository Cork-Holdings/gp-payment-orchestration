package merchantpaymentchannels

import (
	"math"
	"time"

	"github.com/Cork-Holdings/gp_payment_orchestration/internal/global"
	"github.com/Cork-Holdings/gp_payment_orchestration/internal/proto/merchant_payment_channels_proto"
	"github.com/google/uuid"
)

func CreateMerchantPaymentChannel(req *merchant_payment_channels_proto.CreateMerchantPaymentChannelRequest) error {

	merchantID, err := uuid.Parse(req.MerchantId)
	if err != nil {
		return err
	}
	paymentChannelID, err := uuid.Parse(req.PaymentChannelId)
	if err != nil {
		return err
	}
	assignedBy, err := uuid.Parse(req.AssignedBy)
	if err != nil {
		return err
	}
	approvedBy, err := uuid.Parse(req.ApprovedBy)
	if err != nil {
		return err
	}
	merchantPaymentChannel := MerchantPaymentChannel{
		ID:               uuid.New(),
		MerchantID:       merchantID,
		PaymentChannelID: paymentChannelID,
		Status:           req.Status,
		RejectionReason:  req.RejectionReason,
		AssignedBy:       assignedBy,
		ApprovedBy:       approvedBy,
		ApprovalStatus:   "pending",
	}

	tx := global.GetDB().Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
			panic(r)
		}
	}()

	if err := tx.Create(&merchantPaymentChannel).Error; err != nil {
		tx.Rollback()
		return err
	}

	if err := tx.Commit().Error; err != nil {
		tx.Rollback()
		return err
	}
	return nil
}

func GetMerchantPaymentChannel(id string) (*MerchantPaymentChannel, error) {
	merchantPaymentChannel := MerchantPaymentChannel{}
	if err := global.GetDB().Model(&MerchantPaymentChannel{}).Where("id = ?", id).First(&merchantPaymentChannel).Error; err != nil {
		return nil, err
	}
	return &merchantPaymentChannel, nil
}

func GetMerchantPaymentChannels(req *merchant_payment_channels_proto.GetMerchantPaymentChannelsRequest) (*merchant_payment_channels_proto.GetMerchantPaymentChannelsResponse, error) {
	merchantPaymentChannels := []MerchantPaymentChannel{}
	query := global.GetDB().Preload("PaymentChannel").Model(&MerchantPaymentChannel{}).Where("merchant_id = ?", req.MerchantId)
	if req.SearchQuery != "" {
		query = query.Where("name LIKE ?", "%"+req.SearchQuery+"%")
	}

	total := int64(0)
	if err := query.Count(&total).Error; err != nil {
		return nil, err
	}

	if err := query.Find(&merchantPaymentChannels).Error; err != nil {
		return nil, err
	}

	if req.Page > 0 {
		query = query.Offset(int((req.Page - 1) * req.PageSize))
	}
	if req.PageSize > 0 {
		query = query.Limit(int(req.PageSize))
	}

	var merchantPaymentChannelsRes []*merchant_payment_channels_proto.MerchantPaymentChannel
	for _, merchantPaymentChannel := range merchantPaymentChannels {
		paymentChannelID := ""
		paymentChannelName := ""
		if merchantPaymentChannel.PaymentChannelID != uuid.Nil {
			paymentChannelID = merchantPaymentChannel.PaymentChannelID.String()
			paymentChannelName = merchantPaymentChannel.PaymentChannel.Name
		}
		merchantName := ""
		// if merchantPaymentChannel.MerchantID != uuid.Nil {
		// 	merchantName = merchantPaymentChannel.Merchant.Name
		// }
		createdAt := ""
		if merchantPaymentChannel.CreatedAt != nil {
			createdAt = merchantPaymentChannel.CreatedAt.Format(time.RFC3339)
		}
		updatedAt := ""
		if merchantPaymentChannel.UpdatedAt != nil {
			updatedAt = merchantPaymentChannel.UpdatedAt.Format(time.RFC3339)
		}
		assignedBy := ""
		if merchantPaymentChannel.AssignedBy != uuid.Nil {
			assignedBy = merchantPaymentChannel.AssignedBy.String()
		}
		approvedBy := ""
		if merchantPaymentChannel.ApprovedBy != uuid.Nil {
			approvedBy = merchantPaymentChannel.ApprovedBy.String()
		}
		merchantPaymentChannelsRes = append(merchantPaymentChannelsRes, &merchant_payment_channels_proto.MerchantPaymentChannel{
			Id:                 merchantPaymentChannel.ID.String(),
			MerchantId:         merchantPaymentChannel.MerchantID.String(),
			PaymentChannelId:   paymentChannelID,
			Status:             merchantPaymentChannel.Status,
			PaymentChannelName: paymentChannelName,
			MerchantName:       merchantName,
			ApprovalStatus:     merchantPaymentChannel.ApprovalStatus,
			RejectionReason:    merchantPaymentChannel.RejectionReason,
			AssignedBy:         assignedBy,
			ApprovedBy:         approvedBy,
			CreatedAt:          createdAt,
			UpdatedAt:          updatedAt,
		})
	}
	return &merchant_payment_channels_proto.GetMerchantPaymentChannelsResponse{
		MerchantPaymentChannels: merchantPaymentChannelsRes,
		TotalPages:              int32(math.Ceil(float64(total) / float64(req.PageSize))),
		CurrentPage:             req.Page,
		HasMore:                 total > int64(req.Page*req.PageSize),
	}, nil
}

func UpdateMerchantPaymentChannel(req *merchant_payment_channels_proto.EditMerchantPaymentChannelRequest) error {
	updates := map[string]interface{}{}

	if req.Status != "" {
		updates["status"] = req.Status
	}
	if req.RejectionReason != "" {
		updates["rejection_reason"] = req.RejectionReason
	}
	if req.AssignedBy != "" {
		updates["assigned_by"] = req.AssignedBy
	}
	if req.ApprovedBy != "" {
		updates["approved_by"] = req.ApprovedBy
	}
	if req.ApprovalStatus != "" {
		updates["approval_status"] = req.ApprovalStatus
	}
	if req.MerchantId != "" {
		merchantID, err := uuid.Parse(req.MerchantId)
		if err != nil {
			return err
		}
		updates["merchant_id"] = merchantID
	}
	if req.PaymentChannelId != "" {
		paymentChannelID, err := uuid.Parse(req.PaymentChannelId)
		if err != nil {
			return err
		}
		updates["payment_channel_id"] = paymentChannelID
	}

	tx := global.GetDB().Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
			panic(r)
		}
	}()
	if err := tx.Model(&MerchantPaymentChannel{}).Where("id = ?", req.Id).Updates(updates).Error; err != nil {
		tx.Rollback()
		return err
	}
	if err := tx.Commit().Error; err != nil {
		tx.Rollback()
		return err
	}
	return nil
}

func DeleteMerchantPaymentChannel(id string) error {
	merchantPaymentChannel := MerchantPaymentChannel{}
	if err := global.GetDB().Model(&MerchantPaymentChannel{}).Where("id = ?", id).First(&merchantPaymentChannel).Error; err != nil {
		return err
	}
	tx := global.GetDB().Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
			panic(r)
		}
	}()
	if err := tx.Model(&MerchantPaymentChannel{}).Where("id = ?", id).Delete(&MerchantPaymentChannel{}).Error; err != nil {
		tx.Rollback()
		return err
	}
	if err := tx.Commit().Error; err != nil {
		tx.Rollback()
		return err
	}
	return nil
}
