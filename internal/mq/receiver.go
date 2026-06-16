package mq

import (
	"github.com/Cork-Holdings/gp_payment_orchestration/internal/global"
	"github.com/Cork-Holdings/gp_payment_orchestration/internal/mq/handlers"
	"github.com/rabbitmq/amqp091-go"
)

func Reciever(app *global.App, msg amqp091.Delivery) error {
	switch msg.RoutingKey {
	case "transaction.completed":
		return handlers.HandleTransactionCompleted(app, msg)
	case "transaction.failed":
		return handlers.HandleTransactionFailed(app, msg)
	case "settlement.completed":
		return handlers.HandleSettlementCompleted(app, msg)
	case "kyc.approved":
		return handlers.HandleKycApproved(app, msg)
	}
	return nil
}
