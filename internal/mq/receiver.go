package mq

import (
	"log"

	"github.com/Cork-Holdings/gp_payment_orchestration/internal/global"
	"github.com/Cork-Holdings/gp_payment_orchestration/internal/m_api"
	"github.com/rabbitmq/amqp091-go"
)

func Reciever(app *global.App, msg amqp091.Delivery) error {
	log.Printf("[MQ Consumer] Received routing key: %s", msg.RoutingKey)

	switch msg.RoutingKey {
	case "auth.generate_token":
		resp, err := m_api.HandleGenerateToken(app, msg.Body)
		if err != nil {
			log.Printf("[MQ] Failed to generate token: %v", err)
			return err
		}
		return respond(app, msg, resp)

	case "auth.verify_token_ip":
		resp, err := m_api.HandleVerifyTokenAndIP(app, msg.Body)
		if err != nil {
			log.Printf("[MQ] Failed to verify token/IP: %v", err)
			return err
		}
		return respond(app, msg, resp)

	case "collection.collect":
		err := m_api.HandleCollect(app, msg.Body)
		if err != nil {
			log.Printf("[MQ] Failed to process collection: %v", err)
			return err
		}
		return nil

	case "disbursement.disburse":
		resp, err := m_api.HandleDisburse(app, msg.Body)
		if err != nil {
			log.Printf("[MQ] Failed to process disbursement: %v", err)
			return err
		}
		return respond(app, msg, resp)
	}

	return nil
}

func respond(app *global.App, msg amqp091.Delivery, response []byte) error {
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
