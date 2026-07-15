package mq

import (
	"encoding/json"
	"log"
	"strings"

	"github.com/Cork-Holdings/gp_payment_orchestration/internal/global"
	"github.com/Cork-Holdings/gp_payment_orchestration/internal/modules/feecalculator"
	"github.com/Cork-Holdings/gp_payment_orchestration/internal/modules/feeprofiles"
	"github.com/rabbitmq/amqp091-go"
)

func Reciever(app *global.App, msg amqp091.Delivery) error {
	log.Printf("[MQ Consumer] Received routing key: %s", msg.RoutingKey)

	switch msg.RoutingKey {
	case "fees.calculate":
		var req feeprofiles.CalculateFeeRequest
		if err := json.Unmarshal(msg.Body, &req); err != nil {
			log.Printf("[MQ] JSON Unmarshal error: %v", err)
			sendRPCResponse(app, msg, 400, "failed", "Invalid JSON body structure", nil)
			return nil
		}

		txnType := feecalculator.TransactionTypeDisbursement
		if strings.ToUpper(req.TransactionType) == "MNO_COLLECTION" || strings.ToUpper(req.TransactionType) == "COLLECTION" {
			txnType = feecalculator.TransactionTypeCollection
		}

		res, err := feecalculator.CalculateFees(req.MerchantID, req.PhoneNumber, req.Amount, txnType)
		if err != nil {
			log.Printf("[MQ] Failed to calculate fees: %v", err)
			sendRPCResponse(app, msg, 500, "failed", err.Error(), nil)
			return err
		}
		if res.Status == "error" {
			sendRPCResponse(app, msg, 400, "failed", res.Error, nil)
			return nil
		}

		sendRPCResponse(app, msg, 200, "success", "Fee calculated successfully", res)
		return nil

	}

	return nil
}

func sendRPCResponse(app *global.App, msg amqp091.Delivery, code int, status string, message string, data interface{}) {
	if msg.ReplyTo == "" {
		return
	}

	responseMap := map[string]interface{}{
		"code":    code,
		"status":  status,
		"message": message,
		"data":    data,
	}

	responseBytes, _ := json.Marshal(responseMap)

	err := app.MQ.Channel.Publish(
		"",
		msg.ReplyTo,
		false,
		false,
		amqp091.Publishing{
			ContentType:   "application/json",
			CorrelationId: msg.CorrelationId,
			Body:          responseBytes,
		},
	)
	if err != nil {
		log.Printf("[MQ RPC] Failed to publish RPC response to %s: %v", msg.ReplyTo, err)
	} else {
		log.Printf("[MQ RPC] Successfully sent RPC response back to %s (CorrelationId: %s)", msg.ReplyTo, msg.CorrelationId)
	}
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
