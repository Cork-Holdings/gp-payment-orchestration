package mq

import (
	"encoding/json"
	"log"

	"github.com/Cork-Holdings/gp_payment_orchestration/internal/global"
	"github.com/Cork-Holdings/gp_payment_orchestration/internal/modules/merchantapis"
	"github.com/rabbitmq/amqp091-go"
)

func Reciever(app *global.App, msg amqp091.Delivery) error {
	log.Printf("[MQ Consumer] Received routing key: %s", msg.RoutingKey)

	switch msg.RoutingKey {
	case "auth.generate_token":
		var req merchantapis.TokenRequest
		if err := json.Unmarshal(msg.Body, &req); err != nil {
			return err
		}
		resp, err := merchantapis.HandleGenerateToken(app, req)
		if err != nil {
			log.Printf("[MQ] Failed to generate token: %v", err)
			return err
		}
		respBytes, _ := json.Marshal(resp)
		return respond(app, msg, respBytes)

	case "auth.verify_token_ip":
		resp, err := merchantapis.HandleVerifyTokenAndIP(app, msg.Body)
		if err != nil {
			log.Printf("[MQ] Failed to verify token/IP: %v", err)
			return err
		}
		return respond(app, msg, resp)

	case "collection.collect":
		var req merchantapis.CollectRequest
		if err := json.Unmarshal(msg.Body, &req); err != nil {
			return err
		}
		_, err := merchantapis.HandleCollect(app, &req)
		if err != nil {
			log.Printf("[MQ] Failed to process collection: %v", err)
			return err
		}
		return nil

	case "disbursement.disburse":
		var req merchantapis.DisburseRequest
		if err := json.Unmarshal(msg.Body, &req); err != nil {
			return err
		}
		resp, err := merchantapis.HandleDisburse(app, &req)
		if err != nil {
			log.Printf("[MQ] Failed to process disbursement: %v", err)
			return err
		}
		respBytes, _ := json.Marshal(resp)
		return respond(app, msg, respBytes)
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
