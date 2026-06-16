package channelfeebands

import (
	"math"
	"strconv"
	"time"

	"github.com/Cork-Holdings/gp_payment_orchestration/internal/global"
	"github.com/Cork-Holdings/gp_payment_orchestration/internal/modules/paymentchannels"
	"github.com/Cork-Holdings/gp_payment_orchestration/internal/proto/channel_fee_bands_proto"
	"github.com/google/uuid"
)

func CreateChannelFeeBand(req *channel_fee_bands_proto.CreateChannelFeeBandRequest) error {

	paymentChannelID, err := uuid.Parse(req.PaymentChannelId)
	if err != nil {
		return err
	}
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
	channelFeeBands := paymentchannels.ChannelFeeBands{
		ID:               uuid.New(),
		Name:             req.Name,
		PaymentChannelID: paymentChannelID,
		MinAmount:        minAmount,
		MaxAmount:        maxAmount,
		ChargeAmount:     chargeAmount,
		Status:           req.Status,
	}

	tx := global.GetDB().Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
			panic(r)
		}
	}()
	if err := tx.Create(&channelFeeBands).Error; err != nil {
		tx.Rollback()
		return err
	}
	if err := tx.Commit().Error; err != nil {
		tx.Rollback()
		return err
	}
	return nil
}

func GetChannelFeeBand(id string) (*paymentchannels.ChannelFeeBands, error) {
	channelFeeBands := paymentchannels.ChannelFeeBands{}
	if err := global.GetDB().Model(&paymentchannels.ChannelFeeBands{}).Where("id = ?", id).First(&channelFeeBands).Error; err != nil {
		return nil, err
	}
	return &channelFeeBands, nil
}

func GetChannelFeeBands(req *channel_fee_bands_proto.GetChannelFeeBandsRequest) (*channel_fee_bands_proto.GetChannelFeeBandsResponse, error) {
	channelFeeBands := []paymentchannels.ChannelFeeBands{}
	query := global.GetDB().Model(&paymentchannels.ChannelFeeBands{})
	if req.SearchQuery != "" {
		query = query.Where("name = ?", req.SearchQuery)
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
	if err := query.Find(&channelFeeBands).Error; err != nil {
		return nil, err
	}
	totalPages := int32(math.Ceil(float64(total) / float64(req.PageSize)))
	var channelFeeBandsRes []*channel_fee_bands_proto.ChannelFeeBands
	for _, channelFeeBand := range channelFeeBands {
		paymentChannelID := ""
		paymentChannelName := ""
		if channelFeeBand.PaymentChannelID != uuid.Nil {
			paymentChannelID = channelFeeBand.PaymentChannelID.String()
			paymentChannelName = channelFeeBand.PaymentChannel.Name
		}

		createdAt := ""
		if channelFeeBand.CreatedAt != nil {
			createdAt = channelFeeBand.CreatedAt.Format(time.RFC3339)
		}
		updatedAt := ""
		if channelFeeBand.UpdatedAt != nil {
			updatedAt = channelFeeBand.UpdatedAt.Format(time.RFC3339)
		}

		channelFeeBandsRes = append(channelFeeBandsRes, &channel_fee_bands_proto.ChannelFeeBands{
			Id:                 channelFeeBand.ID.String(),
			Name:               channelFeeBand.Name,
			Status:             channelFeeBand.Status,
			PaymentChannelId:   paymentChannelID,
			PaymentChannelName: paymentChannelName,
			MinAmount:          strconv.FormatFloat(channelFeeBand.MinAmount, 'f', -1, 64),
			MaxAmount:          strconv.FormatFloat(channelFeeBand.MaxAmount, 'f', -1, 64),
			ChargeAmount:       strconv.FormatFloat(channelFeeBand.ChargeAmount, 'f', -1, 64),
			CreatedAt:          createdAt,
			UpdatedAt:          updatedAt,
		})
	}
	return &channel_fee_bands_proto.GetChannelFeeBandsResponse{
		ChannelFeeBands: channelFeeBandsRes,
		TotalPages:      totalPages,
		CurrentPage:     req.Page,
		HasMore:         req.Page < totalPages,
	}, nil
}

func UpdateChannelFeeBand(req *channel_fee_bands_proto.EditChannelFeeBandsRequest) error {
	channelFeeBands := paymentchannels.ChannelFeeBands{}
	if err := global.GetDB().Model(&paymentchannels.ChannelFeeBands{}).Where("id = ?", req.Id).First(&channelFeeBands).Error; err != nil {
		return err
	}
	updates := map[string]interface{}{}
	if req.Name != "" {
		updates["name"] = req.Name
	}
	if req.PaymentChannelId != "" {
		paymentChannelID, err := uuid.Parse(req.PaymentChannelId)
		if err != nil {
			return err
		}
		updates["payment_channel_id"] = paymentChannelID
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
	if err := tx.Model(&paymentchannels.ChannelFeeBands{}).Where("id = ?", req.Id).Updates(updates).Error; err != nil {
		tx.Rollback()
		return err
	}
	if err := tx.Commit().Error; err != nil {
		tx.Rollback()
		return err
	}
	return nil
}

func DeleteChannelFeeBand(id string) error {
	channelFeeBands := paymentchannels.ChannelFeeBands{}
	if err := global.GetDB().Model(&paymentchannels.ChannelFeeBands{}).Where("id = ?", id).First(&channelFeeBands).Delete(&paymentchannels.ChannelFeeBands{}).Error; err != nil {
		return err
	}
	tx := global.GetDB().Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
			panic(r)
		}
	}()
	if err := tx.Model(&paymentchannels.ChannelFeeBands{}).Where("id = ?", id).Delete(&paymentchannels.ChannelFeeBands{}).Error; err != nil {
		tx.Rollback()
		return err
	}
	if err := tx.Commit().Error; err != nil {
		tx.Rollback()
		return err
	}

	return nil
}
