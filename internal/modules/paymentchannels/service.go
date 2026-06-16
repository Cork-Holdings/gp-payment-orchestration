package paymentchannels

import (
	"math"
	"time"

	"github.com/Cork-Holdings/gp_payment_orchestration/internal/global"
	"github.com/Cork-Holdings/gp_payment_orchestration/internal/proto/payment_channels_proto"
	"github.com/google/uuid"
)

func CreatePaymentChannel(req *payment_channels_proto.CreatePaymentChannelRequest) error {

	transactionTypeID, err := uuid.Parse(req.TransactionTypeId)
	if err != nil {
		return err
	}
	subTransactionTypeID, err := uuid.Parse(req.SubTransactionTypeId)
	if err != nil {
		return err
	}
	paymentServiceID, err := uuid.Parse(req.PaymentServiceId)
	if err != nil {
		return err
	}

	paymentChannel := PaymentChannel{
		ID:                   uuid.New(),
		Name:                 req.Name,
		Status:               req.Status,
		PaymentServiceID:     paymentServiceID,
		FeeType:              req.FeeType,
		ProviderFee:          req.ProviderFee,
		TransactionTypeID:    transactionTypeID,
		SubTransactionTypeID: subTransactionTypeID,
	}

	tx := global.GetDB().Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
			panic(r)
		}
	}()

	if err := tx.Model(&PaymentChannel{}).Create(&paymentChannel).Error; err != nil {
		tx.Rollback()
		return err
	}

	if err := tx.Commit().Error; err != nil {
		tx.Rollback()
		return err
	}
	return nil
}

func GetPaymentChannel(id string) (*PaymentChannel, error) {
	paymentChannel := PaymentChannel{}
	if err := global.GetDB().Model(&PaymentChannel{}).Where("id = ?", id).First(&paymentChannel).Error; err != nil {
		return nil, err
	}
	return &paymentChannel, nil
}

func GetPaymentChannels(req *payment_channels_proto.GetPaymentChannelsRequest) (*payment_channels_proto.GetPaymentChannelsResponse, error) {
	paymentChannels := []PaymentChannel{}

	query := global.GetDB().Model(&PaymentChannel{})

	if req.SearchQuery != "" {
		query = query.Where("name LIKE ?", "%"+req.SearchQuery+"%")
	}
	if req.PaymentServiceId != "" {
		query = query.Where("payment_service_id = ?", req.PaymentServiceId)
	}
	if req.TransactionTypeId != "" {
		query = query.Where("transaction_type_id = ?", req.TransactionTypeId)
	}
	if req.SubTransactionTypeId != "" {
		query = query.Where("sub_transaction_type_id = ?", req.SubTransactionTypeId)
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

	if err := query.Find(&paymentChannels).Error; err != nil {
		return nil, err
	}
	totalPages := int32(math.Ceil(float64(total) / float64(req.PageSize)))

	var paymentChannelRes []*payment_channels_proto.PaymentChannel
	for _, paymentChannel := range paymentChannels {

		transactionId := ""
		subTransactionId := ""
		if paymentChannel.TransactionTypeID != uuid.Nil {
			transactionId = paymentChannel.TransactionTypeID.String()
		}
		if paymentChannel.SubTransactionTypeID != uuid.Nil {
			subTransactionId = paymentChannel.SubTransactionTypeID.String()
		}

		createdAt := ""
		if paymentChannel.CreatedAt != nil {
			createdAt = paymentChannel.CreatedAt.Format(time.RFC3339)
		}
		updatedAt := ""
		if paymentChannel.UpdatedAt != nil {
			updatedAt = paymentChannel.UpdatedAt.Format(time.RFC3339)
		}

		paymentChannelRes = append(paymentChannelRes, &payment_channels_proto.PaymentChannel{
			Id:                   paymentChannel.ID.String(),
			Name:                 paymentChannel.Name,
			Status:               paymentChannel.Status,
			PaymentServiceId:     paymentChannel.PaymentServiceID.String(),
			TransactionTypeId:    transactionId,
			SubTransactionTypeId: subTransactionId,
			FeeType:              paymentChannel.FeeType,
			ProviderFee:          paymentChannel.ProviderFee,
			Logo:                 paymentChannel.Logo,
			CreatedAt:            createdAt,
			UpdatedAt:            updatedAt,
		})
	}
	return &payment_channels_proto.GetPaymentChannelsResponse{
		PaymentChannel: paymentChannelRes,
		TotalPages:     totalPages,
		CurrentPage:    req.Page,
		HasMore:        int32(req.Page) < totalPages,
	}, nil
}

func UpdatePaymentChannel(req *payment_channels_proto.EditPaymentChannelRequest) error {

	updates := map[string]interface{}{}
	if req.Name != "" {
		updates["name"] = req.Name
	}
	if req.Status != "" {
		updates["status"] = req.Status
	}
	if req.PaymentServiceId != "" {
		updates["payment_service_id"] = req.PaymentServiceId
	}
	if req.TransactionTypeId != "" {
		updates["transaction_type_id"] = req.TransactionTypeId
	}
	if req.SubTransactionTypeId != "" {
		updates["sub_transaction_type_id"] = req.SubTransactionTypeId
	}
	if req.FeeType != "" {
		updates["fee_type"] = req.FeeType
	}
	if req.ProviderFee != "" {
		updates["provider_fee"] = req.ProviderFee
	}

	tx := global.GetDB().Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
			panic(r)
		}
	}()

	if err := tx.Model(&PaymentChannel{}).Where("id = ?", req.Id).Updates(updates).Error; err != nil {
		tx.Rollback()
		return err
	}

	if err := tx.Commit().Error; err != nil {
		tx.Rollback()
		return err
	}
	return nil
}

func DeletePaymentChannel(id string) error {
	tx := global.GetDB().Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
			panic(r)
		}
	}()

	if err := tx.Model(&PaymentChannel{}).Where("id = ?", id).Delete(&PaymentChannel{}).Error; err != nil {
		tx.Rollback()
		return err
	}
	if err := tx.Commit().Error; err != nil {
		tx.Rollback()
		return err
	}
	return nil
}
