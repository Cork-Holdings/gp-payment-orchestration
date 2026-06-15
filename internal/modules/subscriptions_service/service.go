package subscriptionsservice

import (
	"github.com/Cork-Holdings/gp_payment_orchestration/internal/global"
	"github.com/Cork-Holdings/gp_payment_orchestration/internal/proto/subscriptions_proto"
)

func CreateSubscription(req *subscriptions_proto.CreateSubscriptionRequest) error {
	subscription := &subscriptions_proto.Subscription{
		Name:   req.Name,
		Status: req.Status,
		Code:   req.Code,
	}

	tx := global.GetDB().Begin()

	if err := tx.Create(subscription).Error; err != nil {
		tx.Rollback()
		return err
	}
	return nil
}

func GetSubscriptions(req *subscriptions_proto.GetSubscriptionsRequest) ([]*subscriptions_proto.Subscription, error) {
	subscriptions := []*subscriptions_proto.Subscription{}
	if err := global.GetDB().Find(&subscriptions).Error; err != nil {
		return nil, err
	}
	return subscriptions, nil
}
