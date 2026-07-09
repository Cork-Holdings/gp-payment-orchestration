package mq

import (
	"log"

	"github.com/Cork-Holdings/gp_payment_orchestration/internal/global"
	"github.com/rabbitmq/amqp091-go"
)

func Reciever(app *global.App, msg amqp091.Delivery) error {
	log.Printf("[MQ Consumer] Received routing key: %s", msg.RoutingKey)

	switch msg.RoutingKey {
	}

	return nil
}

func Respond(app *global.App, msg amqp091.Delivery, response []byte) error {
	if msg.ReplyTo == "" {
		return nil
	}
	return app.MQ.Channel.Publish(
		"", // Default exchange
		msg.ReplyTo,
		false,
		false,
		amqp091.Publishing{
			ContentType:   "application/json",
			CorrelationId: msg.CorrelationId,
			Body:          response,
		},
	)
}
