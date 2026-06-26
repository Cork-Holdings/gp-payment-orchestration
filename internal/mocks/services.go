package mocks

import (
	"encoding/json"
	"log"
	"os"

	"github.com/Cork-Holdings/gp_payment_orchestration/internal/global"
	"github.com/Cork-Holdings/gp_payment_orchestration/utils"
	"github.com/rabbitmq/amqp091-go"
)

func StartFakeTransactionService(app *global.App) {

	ch, err := app.MQ.Conn.Channel()
	if err != nil {
		panic(err)
	}

	queue, err := ch.QueueDeclare(
		"transactions.check_status",
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		panic(err)
	}

	err = ch.QueueBind(
		queue.Name,
		"transactions.check_status",
		os.Getenv("EXCHANGE"),
		false,
		nil,
	)
	if err != nil {
		panic(err)
	}

	msgs, err := ch.Consume(
		queue.Name,
		"",
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		panic(err)
	}

	go func() {
		for msg := range msgs {

			log.Println("📩 Fake service received:", string(msg.Body))
			log.Println("ReplyTo:", msg.ReplyTo)
			log.Println("CorrelationID:", msg.CorrelationId)

			//time.Sleep(2 * time.Second)

			response := map[string]any{
				"code":    200,
				"status":  "success",
				"message": "Transaction status retrieved successfully",
				"data": map[string]any{
					"transaction_ref": "TXN123",
					"status":          "SUCCESS",
				},
			}

			body, _ := json.Marshal(response)

			ch.Publish(
				"",
				msg.ReplyTo,
				false,
				false,
				amqp091.Publishing{
					CorrelationId: msg.CorrelationId,
					Body:          body,
				},
			)
		}
	}()
}

func StartFakeMerchantService(app *global.App) {

	ch, err := app.MQ.Conn.Channel()
	if err != nil {
		panic(err)
	}

	queue, err := ch.QueueDeclare(
		"merchants.get_details",
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		panic(err)
	}

	err = ch.QueueBind(
		queue.Name,
		"merchants.get_details",
		os.Getenv("EXCHANGE"),
		false,
		nil,
	)
	if err != nil {
		panic(err)
	}

	msgs, err := ch.Consume(
		queue.Name,
		"",
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		panic(err)
	}

	go func() {
		for msg := range msgs {

			log.Println("📩 Fake Merchant Service received:", string(msg.Body))
			log.Println("ReplyTo:", msg.ReplyTo)
			log.Println("CorrelationID:", msg.CorrelationId)

			//time.Sleep(2 * time.Second)

			response := map[string]any{
				"code":    200,
				"status":  "success",
				"message": "Merchant details retrieved successfully",
				"data": map[string]any{
					"merchant_id": "MCH-90210",
					"name":        "Merchant Name",
					"email":       "[EMAIL_ADDRESS]",
					"phone":       "1234567890",
					"status":      "ACTIVE",
				},
			}

			body, _ := json.Marshal(response)

			ch.Publish(
				"",
				msg.ReplyTo,
				false,
				false,
				amqp091.Publishing{
					CorrelationId: msg.CorrelationId,
					Body:          body,
				},
			)
		}
	}()

}

func StartMerchantCollectionBalanceService(app *global.App) {

	ch, err := app.MQ.Conn.Channel()
	if err != nil {
		panic(err)
	}

	queue, err := ch.QueueDeclare(
		"merchant.accounts.collection.check_balance",
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		panic(err)
	}

	err = ch.QueueBind(
		queue.Name,
		"merchant.accounts.collection.check_balance",
		os.Getenv("EXCHANGE"),
		false,
		nil,
	)
	if err != nil {
		panic(err)
	}

	msgs, err := ch.Consume(
		queue.Name,
		"",
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		panic(err)
	}

	go func() {
		for msg := range msgs {

			log.Println("📩 Fake Merchant Accounts Service received:", string(msg.Body))
			log.Println("ReplyTo:", msg.ReplyTo)
			log.Println("CorrelationID:", msg.CorrelationId)

			//time.Sleep(2 * time.Second)

			response := map[string]any{
				"status":  "success",
				"message": "Collection balance fetched successfully.",
				"data": map[string]any{
					"merchant":     "INTERNAL TEST ACCOUNT",
					"balance":      "4.86",
					"currency":     "ZMW",
					"last_updated": "2026-06-24 11:56:17",
				},
			}
			body, _ := json.Marshal(response)

			ch.Publish(
				"",
				msg.ReplyTo,
				false,
				false,
				amqp091.Publishing{
					CorrelationId: msg.CorrelationId,
					Body:          body,
				},
			)
		}
	}()

}

func StartMerchantDisbursementBalanceService(app *global.App) {

	ch, err := app.MQ.Conn.Channel()
	if err != nil {
		panic(err)
	}

	queue, err := ch.QueueDeclare(
		"merchant.accounts.disbursement.check_balance",
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		panic(err)
	}

	err = ch.QueueBind(
		queue.Name,
		"merchant.accounts.disbursement.check_balance",
		os.Getenv("EXCHANGE"),
		false,
		nil,
	)
	if err != nil {
		panic(err)
	}

	msgs, err := ch.Consume(
		queue.Name,
		"",
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		panic(err)
	}

	go func() {
		for msg := range msgs {

			log.Println("📩 Fake Merchant Disbursements Balance Service received:", string(msg.Body))
			log.Println("ReplyTo:", msg.ReplyTo)
			log.Println("CorrelationID:", msg.CorrelationId)

			//time.Sleep(2 * time.Second)

			response := map[string]any{
				"status":  "success",
				"message": "Disbursement balance fetched successfully.",
				"data": map[string]any{
					"merchant":     "INTERNAL TEST ACCOUNT",
					"balance":      "652.90",
					"currency":     "ZMW",
					"last_updated": "2026-06-21 16:09:24",
				},
			}
			body, _ := json.Marshal(response)

			ch.Publish(
				"",
				msg.ReplyTo,
				false,
				false,
				amqp091.Publishing{
					CorrelationId: msg.CorrelationId,
					Body:          body,
				},
			)
		}
	}()

}

func StartFakeCollectionsService(transactionRef string, app *global.App) {

	ch, err := app.MQ.Conn.Channel()
	if err != nil {
		panic(err)
	}

	queue, err := ch.QueueDeclare(
		"transaction.execute.collection",
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		panic(err)
	}

	err = ch.QueueBind(
		queue.Name,
		"transaction.execute.collection",
		os.Getenv("EXCHANGE"),
		false,
		nil,
	)
	if err != nil {
		panic(err)
	}

	msgs, err := ch.Consume(
		queue.Name,
		"",
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		panic(err)
	}

	go func() {
		for msg := range msgs {

			log.Println("📩 Fake Collections Service received:", string(msg.Body))
			log.Println("ReplyTo:", msg.ReplyTo)
			log.Println("CorrelationID:", msg.CorrelationId)

			//time.Sleep(2 * time.Second)

			response := map[string]any{
				"code":    200,
				"status":  "pending",
				"message": "Payment Request Sent successfully. Awaiting customer action",
				"data": map[string]any{
					"transaction_ref": transactionRef,
					"status":          "pending",
				},
			}

			body, _ := json.Marshal(response)

			ch.Publish(
				"",
				msg.ReplyTo,
				false,
				false,
				amqp091.Publishing{
					CorrelationId: msg.CorrelationId,
					Body:          body,
				},
			)
		}
	}()

}

func StartFakeDisbursementsService(transactionRef string, app *global.App) {

	ch, err := app.MQ.Conn.Channel()
	if err != nil {
		panic(err)
	}

	queue, err := ch.QueueDeclare(
		"transaction.execute.disbursement",
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		panic(err)
	}

	err = ch.QueueBind(
		queue.Name,
		"transaction.execute.disbursement",
		os.Getenv("EXCHANGE"),
		false,
		nil,
	)
	if err != nil {
		panic(err)
	}

	msgs, err := ch.Consume(
		queue.Name,
		"",
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		panic(err)
	}

	go func() {
		for msg := range msgs {

			log.Println("📩 Fake Disbursement Service received:", string(msg.Body))
			log.Println("ReplyTo:", msg.ReplyTo)
			log.Println("CorrelationID:", msg.CorrelationId)

			//time.Sleep(2 * time.Second)

			response := map[string]any{
				"code":    200,
				"status":  "success",
				"message": "Disbursement successful",
				"data": map[string]any{
					"transaction_ref": transactionRef,
					"external_ref":    utils.GenerateTenDigitCode(),
					"status":          "successful",
				},
			}

			body, _ := json.Marshal(response)

			ch.Publish(
				"",
				msg.ReplyTo,
				false,
				false,
				amqp091.Publishing{
					CorrelationId: msg.CorrelationId,
					Body:          body,
				},
			)
		}
	}()

}
