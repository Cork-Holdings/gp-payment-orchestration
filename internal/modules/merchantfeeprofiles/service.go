package merchantfeeprofiles

import (
	"math"

	"github.com/Cork-Holdings/gp_payment_orchestration/internal/global"
	"github.com/Cork-Holdings/gp_payment_orchestration/internal/proto/merchant_fee_profile_proto"
	"github.com/google/uuid"
)

func CreateMerchantFeeProfile(req *merchant_fee_profile_proto.CreateMerchantFeeProfileRequest) error {
	merchantFeeProfile := MerchantFeeProfile{
		ID:           uuid.New(),
		MerchantID:   uuid.MustParse(req.MerchantId),
		FeeProfileID: uuid.MustParse(req.FeeProfileId),
		Status:       req.Status,
	}

	tx := global.GetDB().Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
			panic(r)
		}
	}()

	if err := tx.Model(&MerchantFeeProfile{}).Create(&merchantFeeProfile).Error; err != nil {
		tx.Rollback()
		return err
	}

	if err := tx.Commit().Error; err != nil {
		tx.Rollback()
		return err
	}
	return nil
}

func GetMerchantFeeProfile(id string) (*MerchantFeeProfile, error) {
	merchantFeeProfile := MerchantFeeProfile{}
	if err := global.GetDB().Model(&MerchantFeeProfile{}).Where("id = ?", id).First(&merchantFeeProfile).Error; err != nil {
		return nil, err
	}
	return &merchantFeeProfile, nil
}

func GetMerchantFeeProfiles(req *merchant_fee_profile_proto.GetMerchantFeeProfilesRequest) (*merchant_fee_profile_proto.GetMerchantFeeProfilesResponse, error) {

	var merchantFeeProfiles []MerchantFeeProfile

	page := req.Page
	pageSize := req.PageSize
	searchQuery := req.SearchQuery

	limit := uint(pageSize)
	offset := uint((page - 1) * pageSize)

	query := global.GetDB().Model(&MerchantFeeProfile{}).Preload("FeeProfile").Preload("Merchant")

	if searchQuery != "" {
		if _, err := uuid.Parse(searchQuery); err == nil {
			query = query.Where("merchant_id = ? OR fee_profile_id = ?", searchQuery, searchQuery)
		}
	}

	total := int64(0)
	if err := query.Count(&total).Error; err != nil {
		return nil, err
	}

	if err := query.Offset(int(offset)).Limit(int(limit)).Find(&merchantFeeProfiles).Error; err != nil {
		return nil, err
	}
	totalPages := int32(math.Ceil(float64(total) / float64(pageSize)))

	var merchantRes []*merchant_fee_profile_proto.MerchantFeeProfile
	for _, merchantFeeProfile := range merchantFeeProfiles {
		merchantRes = append(merchantRes, &merchant_fee_profile_proto.MerchantFeeProfile{
			Id:           merchantFeeProfile.ID.String(),
			MerchantId:   merchantFeeProfile.MerchantID.String(),
			FeeProfileId: merchantFeeProfile.FeeProfileID.String(),
			Status:       merchantFeeProfile.Status,
			MerchantName: "",
		})
	}

	return &merchant_fee_profile_proto.GetMerchantFeeProfilesResponse{
		MerchantFeeProfile: merchantRes,
		TotalPages:         totalPages,
		CurrentPage:        page,
		HasMore:            page < totalPages,
	}, nil
}

func UpdateMerchantFeeProfile(req *merchant_fee_profile_proto.EditMerchantFeeProfileRequest) error {

	updates := map[string]interface{}{}
	if req.MerchantId != "" {
		updates["merchant_id"] = uuid.MustParse(req.MerchantId)
	}
	if req.FeeProfileId != "" {
		updates["fee_profile_id"] = uuid.MustParse(req.FeeProfileId)
	}
	if req.Status != "" {
		updates["status"] = req.Status
	}

	tx := global.GetDB().Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
			panic(r)
		}
	}()

	if err := tx.Model(&MerchantFeeProfile{}).Where("id = ?", req.Id).Updates(updates).Error; err != nil {
		tx.Rollback()
		return err
	}

	if err := tx.Commit().Error; err != nil {
		tx.Rollback()
		return err
	}

	return nil
}

func DeleteMerchantFeeProfile(id string) error {
	tx := global.GetDB().Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
			panic(r)
		}
	}()

	if err := tx.Model(&MerchantFeeProfile{}).Where("id = ?", id).Delete(&MerchantFeeProfile{}).Error; err != nil {
		tx.Rollback()
		return err
	}

	if err := tx.Commit().Error; err != nil {
		tx.Rollback()
		return err
	}
	return nil
}
