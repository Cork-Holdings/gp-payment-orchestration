package providers

import (
	"math"
	"time"

	"github.com/Cork-Holdings/gp_payment_orchestration/internal/global"
	"github.com/Cork-Holdings/gp_payment_orchestration/internal/proto/providers_proto"
	"github.com/google/uuid"
)

func CreateProvider(req *providers_proto.CreateProviderRequest) error {

	provider := Provider{
		ID:     uuid.New(),
		Name:   req.Name,
		Code:   req.Code,
		Status: req.Status,
	}
	tx := global.GetDB().Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
			panic(r)
		}
	}()

	if err := tx.Create(&provider).Error; err != nil {
		tx.Rollback()
		return err
	}

	if err := tx.Commit().Error; err != nil {
		tx.Rollback()
		return err
	}

	return nil
}

func GetProvider(id string) (*Provider, error) {
	var provider Provider
	if err := global.GetDB().Where("id = ?", id).First(&provider).Error; err != nil {
		return nil, err
	}
	providerRes := Provider{
		ID:     provider.ID,
		Name:   provider.Name,
		Code:   provider.Code,
		Status: provider.Status,
	}
	return &providerRes, nil
}

func GetProviders(req *providers_proto.GetProvidersRequest) (*providers_proto.GetProvidersResponse, error) {
	var providers []Provider
	query := global.GetDB().Model(&Provider{})
	if req.ProviderName != "" {
		query = query.Where("name LIKE ?", "%"+req.ProviderName+"%")
	}

	total := int64(0)
	if err := query.Count(&total).Error; err != nil {
		return nil, err
	}

	if req.ProviderName != "" {
		query = query.Where("name LIKE ?", "%"+req.ProviderName+"%")
	}

	if req.Page > 0 {
		query = query.Offset(int((req.Page - 1) * req.PageSize))
	}
	if req.PageSize > 0 {
		query = query.Limit(int(req.PageSize))
	}
	if err := query.Find(&providers).Error; err != nil {
		return nil, err
	}
	totalPages := int32(math.Ceil(float64(total) / float64(req.PageSize)))

	var providerRes []*providers_proto.Provider
	for _, provider := range providers {
		providerRes = append(providerRes, &providers_proto.Provider{
			Id:        provider.ID.String(),
			Name:      provider.Name,
			Code:      provider.Code,
			Status:    provider.Status,
			CreatedAt: provider.CreatedAt.Format(time.RFC3339),
			UpdatedAt: provider.UpdatedAt.Format(time.RFC3339),
		})
	}

	return &providers_proto.GetProvidersResponse{
		Providers:   providerRes,
		TotalPages:  totalPages,
		CurrentPage: req.Page,
		HasMore:     req.Page < totalPages,
	}, nil
}

func UpdateProvider(req *providers_proto.EditProviderRequest) error {

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

	tx := global.GetDB().Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
			panic(r)
		}
	}()

	if err := tx.Model(&Provider{}).Where("id = ?", req.Id).Updates(updates).Error; err != nil {
		tx.Rollback()
		return err
	}

	if err := tx.Commit().Error; err != nil {
		tx.Rollback()
		return err
	}
	return nil
}

func DeleteProvider(Id string) error {
	tx := global.GetDB().Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
			panic(r)
		}
	}()

	providerId, err := uuid.Parse(Id)
	if err != nil {
		return err
	}

	//Check if the provider is associated with any fee profiles
	// var count int64
	// if err := tx.Model(&feeprofiles.FeeProfile{}).Where("provider_id = ?", providerId).Count(&count).Error; err != nil {
	// 	tx.Rollback()
	// 	return err
	// }
	// if count > 0 {
	// 	tx.Rollback()
	// 	return fmt.Errorf("cannot delete provider that is linked to fee profiles")
	// }

	if err := tx.Model(&Provider{}).Where("id = ?", providerId).Delete(&Provider{}).Error; err != nil {
		tx.Rollback()
		return err
	}

	if err := tx.Commit().Error; err != nil {
		tx.Rollback()
		return err
	}

	return nil
}
