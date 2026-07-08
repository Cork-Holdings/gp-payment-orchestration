package feeprofiles

import (
	"math"
	"strconv"
	"time"

	"github.com/Cork-Holdings/gp_payment_orchestration/internal/global"
	"github.com/Cork-Holdings/gp_payment_orchestration/internal/proto/fee_profiles_proto"
	"github.com/google/uuid"
)

func CreateFeeProfile(req *fee_profiles_proto.CreateFeeProfileRequest) error {

	chargeAmount, _ := strconv.ParseFloat(req.ChargeAmount, 64)
	minimumFee, _ := strconv.ParseFloat(req.MinimumFee, 64)

	feeprofile := FeeProfile{
		ID:                uuid.New(),
		Name:              req.Name,
		Code:              req.Code,
		PaymentChannelID:  uuid.MustParse(req.PaymentChannelId),
		TransactionTypeID: uuid.MustParse(req.TransactionTypeId),
		Status:            req.Status,
		ChargeAmount:      chargeAmount,
		ApprovalStatus:    "pending",
		CalculationMode:   req.CalculationMode,
		ChargeType:        req.ChargeType,
		MinimumFee:        minimumFee,
	}
	tx := global.GetDB().Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
			panic(r)
		}
	}()

	if err := tx.Create(&feeprofile).Error; err != nil {
		tx.Rollback()
		return err
	}

	if err := tx.Commit().Error; err != nil {
		tx.Rollback()
		return err
	}

	return nil
}

func GetFeeProfile(id string) (*FeeProfile, error) {
	var feeprofile FeeProfile
	if err := global.GetDB().Where("id = ?", id).First(&feeprofile).Error; err != nil {
		return nil, err
	}
	feeProfile := FeeProfile{
		ID:                feeprofile.ID,
		Name:              feeprofile.Name,
		Code:              feeprofile.Code,
		PaymentChannelID:  feeprofile.PaymentChannelID,
		TransactionTypeID: feeprofile.TransactionTypeID,
		Status:            feeprofile.Status,
		ChargeAmount:      feeprofile.ChargeAmount,
		ApprovalStatus:    feeprofile.ApprovalStatus,
		ApprovedBy:        feeprofile.ApprovedBy,
		ApprovedAt:        feeprofile.ApprovedAt,
		RejectedBy:        feeprofile.RejectedBy,
		RejectedAt:        feeprofile.RejectedAt,
		RejectedReason:    feeprofile.RejectedReason,
		CalculationMode:   feeprofile.CalculationMode,
	}
	return &feeProfile, nil
}

func GetFeeProfiles(req *fee_profiles_proto.GetFeeProfilesRequest) (*fee_profiles_proto.GetFeeProfilesResponse, error) {
	var feeprofiles []FeeProfile
	query := global.GetDB().Model(&FeeProfile{})
	if req.SearchQuery != "" {
		query = query.Where("name LIKE ?", "%"+req.SearchQuery+"%")
	}

	total := int64(0)
	if err := query.Count(&total).Error; err != nil {
		return nil, err
	}

	if req.Page > 0 {
		query = query.Offset(int((req.Page - 1) * req.PageSize))
	}
	if req.PageSize > 0 {
		query = query.Limit(int(req.PageSize))
	}
	if err := query.Find(&feeprofiles).Error; err != nil {
		return nil, err
	}
	totalPages := int32(math.Ceil(float64(total) / float64(req.PageSize)))

	var feeProfileRes []*fee_profiles_proto.FeeProfile
	for _, feeprofile := range feeprofiles {
		paymentChannel := ""
		paymentChannelId := ""
		if feeprofile.PaymentChannelID != uuid.Nil {
			paymentChannel = feeprofile.PaymentChannel.Name
			paymentChannelId = feeprofile.PaymentChannelID.String()
		}
		transactionType := ""
		transactionTypeId := ""
		if feeprofile.TransactionTypeID != uuid.Nil {
			transactionType = feeprofile.TransactionType.Name
			transactionTypeId = feeprofile.TransactionTypeID.String()
		}
		subTransactionType := ""
		subTransactionTypeId := ""
		if feeprofile.SubTransactionTypeID != uuid.Nil {
			subTransactionType = feeprofile.SubTransactionType.Name
			subTransactionTypeId = feeprofile.SubTransactionTypeID.String()
		}
		approvedAt := ""
		if feeprofile.ApprovedAt != nil {
			approvedAt = feeprofile.ApprovedAt.Format(time.RFC3339)
		}
		approvedBy := ""
		if feeprofile.ApprovedBy != uuid.Nil {
			approvedBy = feeprofile.ApprovedBy.String()
		}
		rejectedBy := ""
		if feeprofile.RejectedBy != uuid.Nil {
			rejectedBy = feeprofile.RejectedBy.String()
		}
		createdAt := ""
		if feeprofile.CreatedAt != nil {
			createdAt = feeprofile.CreatedAt.Format(time.RFC3339)
		}
		updatedAt := ""
		if feeprofile.UpdatedAt != nil {
			updatedAt = feeprofile.UpdatedAt.Format(time.RFC3339)
		}
		feeProfileRes = append(feeProfileRes, &fee_profiles_proto.FeeProfile{
			Id:                     feeprofile.ID.String(),
			Name:                   feeprofile.Name,
			Code:                   feeprofile.Code,
			Status:                 feeprofile.Status,
			ChargeAmount:           strconv.FormatFloat(feeprofile.ChargeAmount, 'f', -1, 64),
			ApprovalStatus:         feeprofile.ApprovalStatus,
			ApprovedBy:             approvedBy,
			ApprovedAt:             approvedAt,
			RejectedBy:             rejectedBy,
			RejectedReason:         feeprofile.RejectedReason,
			CalculationMode:        feeprofile.CalculationMode,
			PaymentChannelName:     paymentChannel,
			TransactionTypeName:    transactionType,
			SubTransactionTypeName: subTransactionType,
			PaymentChannelId:       paymentChannelId,
			TransactionTypeId:      transactionTypeId,
			SubTransactionTypeId:   subTransactionTypeId,
			CreatedAt:              createdAt,
			UpdatedAt:              updatedAt,
			ChargeType:             feeprofile.ChargeType,
			MinimumFee:             strconv.FormatFloat(feeprofile.MinimumFee, 'f', -1, 64),
		})
	}
	return &fee_profiles_proto.GetFeeProfilesResponse{
		FeeProfile:  feeProfileRes,
		TotalPages:  totalPages,
		CurrentPage: req.Page,
		HasMore:     req.Page < totalPages,
	}, nil
}

func UpdateFeeProfile(req *fee_profiles_proto.EditFeeProfileRequest) error {

	updates := map[string]interface{}{}

	if req.Name != "" {
		updates["name"] = req.Name
	}
	if req.Code != "" {
		updates["code"] = req.Code
	}
	if req.Status != "" {
		updates["status"] = req.Status
	}
	if req.ChargeAmount != "" {
		chargeAmount, _ := strconv.ParseFloat(req.ChargeAmount, 64)
		updates["charge_amount"] = chargeAmount

	}
	if req.PaymentChannelId != "" {
		updates["payment_channel_id"] = uuid.MustParse(req.PaymentChannelId)
	}
	if req.TransactionTypeId != "" {
		updates["transaction_type_id"] = uuid.MustParse(req.TransactionTypeId)
	}
	if req.SubTransactionTypeId != "" {
		updates["sub_transaction_type_id"] = uuid.MustParse(req.SubTransactionTypeId)
	}

	tx := global.GetDB().Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
			panic(r)
		}
	}()

	if err := tx.Model(&FeeProfile{}).Where("id = ?", req.Id).Updates(updates).Error; err != nil {
		tx.Rollback()
		return err
	}

	if err := tx.Commit().Error; err != nil {
		tx.Rollback()
		return err
	}
	return nil
}

func DeleteFeeProfile(Id string) error {
	tx := global.GetDB().Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
			panic(r)
		}
	}()

	feeprofileId, err := uuid.Parse(Id)
	if err != nil {
		return err
	}
	if err := tx.Model(&ProfileFeeBands{}).Where("fee_profile_id = ?", feeprofileId).Delete(&ProfileFeeBands{}).Error; err != nil {
		tx.Rollback()
		return err
	}

	if err := tx.Model(&FeeProfile{}).Where("id = ?", feeprofileId).Delete(&FeeProfile{}).Error; err != nil {
		tx.Rollback()
		return err
	}

	if err := tx.Commit().Error; err != nil {
		tx.Rollback()
		return err
	}

	return nil
}
