package mq

import (
	"github.com/Cork-Holdings/gp_payment_orchestration/internal/global"
	"github.com/rabbitmq/amqp091-go"
)

func Reciever(app *global.App, msg amqp091.Delivery) error {
	switch msg.RoutingKey {
	}
	return nil
}
