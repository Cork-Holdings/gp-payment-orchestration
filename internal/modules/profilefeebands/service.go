package profilefeebands

import (
	"math"
	"strconv"
	"time"

	"github.com/Cork-Holdings/gp_payment_orchestration/internal/global"
	"github.com/Cork-Holdings/gp_payment_orchestration/internal/modules/feeprofiles"
	"github.com/Cork-Holdings/gp_payment_orchestration/internal/proto/profile_fee_bands_proto"
	"github.com/google/uuid"
)

func CreateProfileFeeBands(req *profile_fee_bands_proto.CreateProfileFeeBandsRequest) error {

	minAmount, err := strconv.ParseFloat(req.MinAmount, 64)
	if err != nil {
		return err
	}
	maxAmount, err := strconv.ParseFloat(req.MaxAmount, 64)
	if err != nil {
		return err
	}
	chargeAmount, err := strconv.ParseFloat(req.ChargeAmount, 64)
	if err != nil {
		return err
	}
	profileFeeBands := feeprofiles.ProfileFeeBands{
		ID:           uuid.New(),
		FeeProfileID: uuid.MustParse(req.FeeProfileId),
		MinAmount:    minAmount,
		MaxAmount:    maxAmount,
		ChargeAmount: chargeAmount,
		ChargeType:   req.ChargeType,
		Status:       req.Status,
	}
	tx := global.GetDB().Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
			panic(r)
		}
	}()
	if err := tx.Create(&profileFeeBands).Error; err != nil {
		tx.Rollback()
		return err
	}
	if err := tx.Commit().Error; err != nil {
		tx.Rollback()
		return err
	}
	return nil
}

func GetProfileFeeBand(id string) (*feeprofiles.ProfileFeeBands, error) {
	profileFeeBands := feeprofiles.ProfileFeeBands{}
	if err := global.GetDB().Model(&feeprofiles.ProfileFeeBands{}).Where("id = ?", id).First(&profileFeeBands).Error; err != nil {
		return nil, err
	}
	return &profileFeeBands, nil
}

func GetProfileFeeBands(req *profile_fee_bands_proto.GetProfileFeeBandsRequest) (*profile_fee_bands_proto.GetProfileFeeBandsResponse, error) {
	profileFeeBands := []feeprofiles.ProfileFeeBands{}
	query := global.GetDB().Model(&feeprofiles.ProfileFeeBands{})
	if req.SearchQuery != "" {
		if _, err := uuid.Parse(req.SearchQuery); err == nil {
			query = query.Where("fee_profile_id = ?", req.SearchQuery)
		}
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
	if err := query.Find(&profileFeeBands).Error; err != nil {
		return nil, err
	}
	totalPages := int32(math.Ceil(float64(total) / float64(req.PageSize)))
	var profileFeeBandsRes []*profile_fee_bands_proto.ProfileFeeBands
	for _, profileFeeBand := range profileFeeBands {
		feeProfileId := ""
		feeProfileName := ""
		if profileFeeBand.FeeProfileID != uuid.Nil {
			feeProfileId = profileFeeBand.FeeProfileID.String()
			feeProfileName = profileFeeBand.FeeProfile.Name
		}
		createdAt := ""
		if profileFeeBand.CreatedAt != nil {
			createdAt = profileFeeBand.CreatedAt.Format(time.RFC3339)
		}
		updatedAt := ""
		if profileFeeBand.UpdatedAt != nil {
			updatedAt = profileFeeBand.UpdatedAt.Format(time.RFC3339)
		}
		profileFeeBandsRes = append(profileFeeBandsRes, &profile_fee_bands_proto.ProfileFeeBands{
			Id:             profileFeeBand.ID.String(),
			CreatedAt:      createdAt,
			UpdatedAt:      updatedAt,
			FeeProfileId:   feeProfileId,
			FeeProfileName: feeProfileName,
			MinAmount:      strconv.FormatFloat(profileFeeBand.MinAmount, 'f', -1, 64),
			MaxAmount:      strconv.FormatFloat(profileFeeBand.MaxAmount, 'f', -1, 64),
			ChargeAmount:   strconv.FormatFloat(profileFeeBand.ChargeAmount, 'f', -1, 64),
			ChargeType:     profileFeeBand.ChargeType,
			Status:         profileFeeBand.Status,
		})
	}
	return &profile_fee_bands_proto.GetProfileFeeBandsResponse{
		ProfileFeeBands: profileFeeBandsRes,
		TotalPages:      totalPages,
		CurrentPage:     req.Page,
		HasMore:         req.Page < totalPages,
	}, nil
}

func UpdateProfileFeeBand(req *profile_fee_bands_proto.EditProfileFeeBandsRequest) error {
	profileFeeBands := feeprofiles.ProfileFeeBands{}
	if err := global.GetDB().Model(&feeprofiles.ProfileFeeBands{}).Where("id = ?", req.Id).First(&profileFeeBands).Error; err != nil {
		return err
	}
	updates := map[string]interface{}{}
	if req.Name != "" {
		updates["name"] = req.Name
	}
	if req.FeeProfileId != "" {
		updates["fee_profile_id"] = uuid.MustParse(req.FeeProfileId)
	}
	if req.MinAmount != "" {
		minAmount, err := strconv.ParseFloat(req.MinAmount, 64)
		if err != nil {
			return err
		}
		updates["min_amount"] = minAmount
	}
	if req.MaxAmount != "" {
		maxAmount, err := strconv.ParseFloat(req.MaxAmount, 64)
		if err != nil {
			return err
		}
		updates["max_amount"] = maxAmount
	}
	if req.ChargeAmount != "" {
		chargeAmount, err := strconv.ParseFloat(req.ChargeAmount, 64)
		if err != nil {
			return err
		}
		updates["charge_amount"] = chargeAmount
	}
	if req.ChargeType != "" {
		updates["charge_type"] = req.ChargeType
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
	if err := tx.Model(&feeprofiles.ProfileFeeBands{}).Where("id = ?", req.Id).Updates(updates).Error; err != nil {
		tx.Rollback()
		return err
	}
	if err := tx.Commit().Error; err != nil {
		tx.Rollback()
		return err
	}
	return nil
}

func DeleteProfileFeeBand(id string) error {
	profileFeeBands := feeprofiles.ProfileFeeBands{}
	if err := global.GetDB().Model(&feeprofiles.ProfileFeeBands{}).Where("id = ?", id).First(&profileFeeBands).Delete(&feeprofiles.ProfileFeeBands{}).Error; err != nil {
		return err
	}
	tx := global.GetDB().Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
			panic(r)
		}
	}()
	if err := tx.Model(&feeprofiles.ProfileFeeBands{}).Where("id = ?", id).Delete(&feeprofiles.ProfileFeeBands{}).Error; err != nil {
		tx.Rollback()
		return err
	}
	if err := tx.Commit().Error; err != nil {
		tx.Rollback()
		return err
	}
	return nil
}
