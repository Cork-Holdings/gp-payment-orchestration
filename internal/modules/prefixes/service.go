package prefixes

import (
	"math"
	"time"

	"github.com/Cork-Holdings/gp_payment_orchestration/internal/global"
	"github.com/Cork-Holdings/gp_payment_orchestration/internal/proto/prefixes_proto"
	"github.com/google/uuid"
)

func CreatePrefix(req *prefixes_proto.CreatePrefixRequest) error {
	prefix := Prefix{
		ID:     uuid.New(),
		Prefix: req.Name,
	}
	tx := global.GetDB().Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
			panic(r)
		}
	}()
	if err := tx.Create(&prefix).Error; err != nil {
		tx.Rollback()
		return err
	}
	if err := tx.Commit().Error; err != nil {
		tx.Rollback()
		return err
	}
	return nil
}

func GetPrefix(id string) (*Prefix, error) {
	prefix := Prefix{}
	if err := global.GetDB().Model(&Prefix{}).Where("id = ?", id).First(&prefix).Error; err != nil {
		return nil, err
	}
	return &prefix, nil
}

func GetPrefixes(req *prefixes_proto.GetPrefixesRequest) (*prefixes_proto.GetPrefixesResponse, error) {
	prefixes := []Prefix{}
	query := global.GetDB().Model(&Prefix{})
	if req.SearchQuery != "" {
		query = query.Where("prefix LIKE ?", "%"+req.SearchQuery+"%")
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
	if err := query.Find(&prefixes).Error; err != nil {
		return nil, err
	}
	totalPages := int32(math.Ceil(float64(total) / float64(req.PageSize)))

	var prefixRes []*prefixes_proto.Prefix
	for _, prefix := range prefixes {
		prefixRes = append(prefixRes, &prefixes_proto.Prefix{
			Id:   prefix.ID.String(),
			Name: prefix.Prefix,
		})
	}

	return &prefixes_proto.GetPrefixesResponse{
		Prefixes:    prefixRes,
		TotalPages:  totalPages,
		CurrentPage: req.Page,
		HasMore:     req.Page < totalPages,
	}, nil
}

func UpdatePrefix(req *prefixes_proto.EditPrefixRequest) error {
	prefix := Prefix{}
	if err := global.GetDB().Model(&Prefix{}).Where("id = ?", req.Id).First(&prefix).Error; err != nil {
		return err
	}
	updates := map[string]interface{}{}
	if req.Name != "" {
		updates["prefix"] = req.Name
	}
	tx := global.GetDB().Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
			panic(r)
		}
	}()
	if err := tx.Model(&Prefix{}).Where("id = ?", req.Id).Updates(updates).Error; err != nil {
		tx.Rollback()
		return err
	}
	if err := tx.Commit().Error; err != nil {
		tx.Rollback()
		return err
	}
	return nil
}

func DeletePrefix(id string) error {
	prefix := Prefix{}
	if err := global.GetDB().Model(&Prefix{}).Where("id = ?", id).First(&prefix).Delete(&Prefix{}).Error; err != nil {
		return err
	}
	tx := global.GetDB().Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
			panic(r)
		}
	}()
	if err := tx.Model(&Prefix{}).Where("id = ?", id).Delete(&Prefix{}).Error; err != nil {
		tx.Rollback()
		return err
	}
	if err := tx.Commit().Error; err != nil {
		tx.Rollback()
		return err
	}
	return nil
}

func CreatePrefixPaymentChannel(req *prefixes_proto.CreatePrefixPaymentChannelRequest) error {
	prefixPaymentChannel := PrefixPaymentChannel{
		ID:               uuid.New(),
		PrefixID:         uuid.MustParse(req.PrefixId),
		PaymentChannelID: uuid.MustParse(req.PaymentChannelId),
	}
	tx := global.GetDB().Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
			panic(r)
		}
	}()
	if err := tx.Create(&prefixPaymentChannel).Error; err != nil {
		tx.Rollback()
		return err
	}
	if err := tx.Commit().Error; err != nil {
		tx.Rollback()
		return err
	}
	return nil
}

func GetPrefixPaymentChannels(req *prefixes_proto.GetPrefixPaymentChannelsRequest) (*prefixes_proto.GetPrefixPaymentChannelsResponse, error) {
	prefixPaymentChannels := []PrefixPaymentChannel{}
	query := global.GetDB().Preload("PaymentChannel").Model(&PrefixPaymentChannel{})
	if req.PrefixId != "" {
		query = query.Where("prefix_id = ?", req.PrefixId)
	}
	if req.PaymentChannelId != "" {
		query = query.Where("payment_channel_id = ?", req.PaymentChannelId)
	}

	total := int64(0)
	if err := query.Count(&total).Error; err != nil {
		return nil, err
	}

	if err := query.Find(&prefixPaymentChannels).Error; err != nil {
		return nil, err
	}
	totalPages := int32(math.Ceil(float64(total) / float64(req.PageSize)))

	var prefixPaymentChannelRes []*prefixes_proto.PrefixPaymentChannel
	for _, prefixPaymentChannel := range prefixPaymentChannels {

		paymentChannelId := ""
		paymentChannelName := ""
		if prefixPaymentChannel.PaymentChannelID != uuid.Nil {
			paymentChannelId = prefixPaymentChannel.PaymentChannelID.String()
			paymentChannelName = prefixPaymentChannel.PaymentChannel.Name
		}
		createdAt := ""
		if prefixPaymentChannel.CreatedAt != nil {
			createdAt = prefixPaymentChannel.CreatedAt.Format(time.RFC3339)
		}
		updatedAt := ""
		if prefixPaymentChannel.UpdatedAt != nil {
			updatedAt = prefixPaymentChannel.UpdatedAt.Format(time.RFC3339)
		}
		prefixPaymentChannelRes = append(prefixPaymentChannelRes, &prefixes_proto.PrefixPaymentChannel{
			PrefixId:           prefixPaymentChannel.PrefixID.String(),
			PaymentChannelId:   paymentChannelId,
			CreatedAt:          createdAt,
			UpdatedAt:          updatedAt,
			PaymentChannelName: paymentChannelName,
		})
	}
	return &prefixes_proto.GetPrefixPaymentChannelsResponse{
		PrefixPaymentChannels: prefixPaymentChannelRes,
		TotalPages:            totalPages,
		CurrentPage:           req.Page,
		HasMore:               req.Page < totalPages,
	}, nil
}

func DeletePrefixPaymentChannel(id string) error {
	prefixPaymentChannel := PrefixPaymentChannel{}
	if err := global.GetDB().Model(&PrefixPaymentChannel{}).Where("id = ?", id).First(&prefixPaymentChannel).Delete(&PrefixPaymentChannel{}).Error; err != nil {
		return err
	}
	tx := global.GetDB().Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
			panic(r)
		}
	}()
	if err := tx.Model(&PrefixPaymentChannel{}).Where("id = ?", id).Delete(&PrefixPaymentChannel{}).Error; err != nil {
		tx.Rollback()
		return err
	}
	if err := tx.Commit().Error; err != nil {
		tx.Rollback()
		return err
	}
	return nil
}
