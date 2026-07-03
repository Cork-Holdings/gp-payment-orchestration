package subscriptions

import (
	"math"
	"time"

	"github.com/Cork-Holdings/gp_payment_orchestration/internal/global"
	"github.com/Cork-Holdings/gp_payment_orchestration/internal/proto/subscriptions_proto"
	"github.com/google/uuid"
)

func CreateSubscription(req *subscriptions_proto.CreateSubscriptionRequest) error {
	subscription := Subscription{
		ID:          uuid.New(),
		Name:        req.Name,
		Status:      req.Status,
		Description: req.Description,
	}

	tx := global.GetDB().Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
			panic(r)
		}
	}()

	if err := tx.Create(&subscription).Error; err != nil {
		tx.Rollback()
		return err
	}

	if err := tx.Commit().Error; err != nil {
		tx.Rollback()
		return err
	}

	return nil
}

func GetSubscriptions(req *subscriptions_proto.GetSubscriptionsRequest) (*subscriptions_proto.GetSubscriptionsResponse, error) {

	var subscriptionsModel []Subscription

	query := global.GetDB().Model(&Subscription{})
	if req.SearchQuery != "" {
		query = query.Where("name LIKE ?", "%"+req.SearchQuery+"%")
	}

	total := int64(0)
	if err := query.Count(&total).Error; err != nil {
		return nil, err
	}

	if req.Page == 0 {
		req.Page = 1
	}
	if req.PageSize == 0 {
		req.PageSize = 10
	}

	offset := (req.Page - 1) * req.PageSize
	query = query.Offset(int(offset)).Limit(int(req.PageSize))

	if err := query.Find(&subscriptionsModel).Error; err != nil {
		return nil, err
	}

	totalPages := int32(math.Ceil(float64(total) / float64(req.PageSize)))

	var subRes []*subscriptions_proto.Subscription
	for _, subscription := range subscriptionsModel {
		subRes = append(subRes, &subscriptions_proto.Subscription{
			Id:          subscription.ID.String(),
			Name:        subscription.Name,
			Status:      subscription.Status,
			Description: subscription.Description,
		})
	}

	return &subscriptions_proto.GetSubscriptionsResponse{
		Subscription: subRes,
		TotalPages:   totalPages,
		CurrentPage:  req.Page,
		HasMore:      req.Page < totalPages,
	}, nil
}

func UpdateSubscription(req *subscriptions_proto.EditSubscriptionRequest) error {
	updates := map[string]interface{}{}

	if req.Name != "" {
		updates["name"] = req.Name
	}
	if req.Status != "" {
		updates["status"] = req.Status
	}
	if req.Description != "" {
		updates["description"] = req.Description
	}

	tx := global.GetDB().Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
			panic(r)
		}
	}()

	if err := tx.Model(&Subscription{}).Where("id = ?", req.Id).Updates(updates).Error; err != nil {
		tx.Rollback()
		return err
	}

	if err := tx.Commit().Error; err != nil {
		tx.Rollback()
		return err
	}

	return nil
}

func DeleteSubscription(id string) error {
	tx := global.GetDB().Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
			panic(r)
		}
	}()

	// Delete all merchant subscriptions for this subscription
	if err := tx.Model(&MerchantSubscription{}).Where("subscription_id = ?", id).Delete(&MerchantSubscription{}).Error; err != nil {
		tx.Rollback()
		return err
	}
	if err := tx.Model(&Subscription{}).Where("id = ?", id).Delete(&Subscription{}).Error; err != nil {
		tx.Rollback()
		return err
	}
	if err := tx.Commit().Error; err != nil {
		tx.Rollback()
		return err
	}

	return nil
}

func CreateMerchantSubscription(req *subscriptions_proto.CreateMerchantSubscriptionRequest) error {
	merchantSubscription := MerchantSubscription{
		ID:             uuid.New(),
		MerchantID:     uuid.MustParse(req.MerchantId),
		SubscriptionID: uuid.MustParse(req.SubscriptionId),
		Status:         req.Status,
	}

	tx := global.GetDB().Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
			panic(r)
		}
	}()

	if err := tx.Create(&merchantSubscription).Error; err != nil {
		tx.Rollback()
		return err
	}

	if err := tx.Commit().Error; err != nil {
		tx.Rollback()
		return err
	}

	return nil
}

func GetMerchantSubscriptions(req *subscriptions_proto.GetMerchantSubscriptionsRequest) (*subscriptions_proto.GetMerchantSubscriptionsResponse, error) {
	merchantSubscriptionsModel := []MerchantSubscription{}

	query := global.GetDB().Preload("Subscription").Model(&MerchantSubscription{})
	if req.SearchQuery != "" {
		if _, err := uuid.Parse(req.SearchQuery); err == nil {
			query = query.Where("merchant_id = ?", req.SearchQuery)
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

	if err := query.Find(&merchantSubscriptionsModel).Error; err != nil {
		return nil, err
	}

	totalPages := int32(math.Ceil(float64(total) / float64(req.PageSize)))

	var merchantSubRes []*subscriptions_proto.MerchantSubscription
	for _, merchantSubscription := range merchantSubscriptionsModel {
		createdAt := ""
		if merchantSubscription.CreatedAt != nil {
			createdAt = merchantSubscription.CreatedAt.Format(time.RFC3339)
		}
		updatedAt := ""
		if merchantSubscription.UpdatedAt != nil {
			updatedAt = merchantSubscription.UpdatedAt.Format(time.RFC3339)
		}
		merchantSubRes = append(merchantSubRes, &subscriptions_proto.MerchantSubscription{
			Id:               merchantSubscription.ID.String(),
			MerchantId:       merchantSubscription.MerchantID.String(),
			SubscriptionId:   merchantSubscription.SubscriptionID.String(),
			Status:           merchantSubscription.Status,
			CreatedAt:        createdAt,
			UpdatedAt:        updatedAt,
			SubscriptionName: merchantSubscription.Subscription.Name,
		})
	}
	return &subscriptions_proto.GetMerchantSubscriptionsResponse{
		MerchantSubscriptions: merchantSubRes,
		TotalPages:            totalPages,
		CurrentPage:           req.Page,
		HasMore:               req.Page < totalPages,
	}, nil
}

func UpdateMerchantSubscription(req *subscriptions_proto.EditMerchantSubscriptionRequest) error {
	updates := map[string]interface{}{}

	if req.Status != "" {
		updates["status"] = req.Status
	}

	if req.MerchantId != "" {
		updates["merchant_id"] = uuid.MustParse(req.MerchantId)
	}
	if req.SubscriptionId != "" {
		updates["subscription_id"] = uuid.MustParse(req.SubscriptionId)
	}

	tx := global.GetDB().Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
			panic(r)
		}
	}()

	if err := tx.Model(&MerchantSubscription{}).Where("id = ?", req.Id).Updates(updates).Error; err != nil {
		tx.Rollback()
		return err
	}
	if err := tx.Commit().Error; err != nil {
		tx.Rollback()
		return err
	}

	return nil
}

func DeleteMerchantSubscription(id string) error {

	merchantSubscriptionId, err := uuid.Parse(id)
	if err != nil {
		return err
	}
	tx := global.GetDB().Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
			panic(r)
		}
	}()
	if err := tx.Model(&MerchantSubscription{}).Where("id = ?", merchantSubscriptionId).Delete(&MerchantSubscription{}).Error; err != nil {
		tx.Rollback()
		return err
	}
	if err := tx.Commit().Error; err != nil {
		tx.Rollback()
		return err
	}
	return nil
}
