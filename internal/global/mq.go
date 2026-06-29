package global

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"sync"
	"time"

	"github.com/google/uuid"
	amqp "github.com/rabbitmq/amqp091-go"
)

var (
	rabbitOnce sync.Once
)

type rmq struct {
	Conn         *amqp.Connection
	Channel      *amqp.Channel
	ReplyToQueue string

	pending sync.Map
}

var mq *rmq

func GetMQ() *rmq {
	rabbitOnce.Do(func() {
		if mq == nil {
			mq = &rmq{}
		}
		conn, err := amqp.Dial(os.Getenv("MESSAGE_BROKER_URL"))
		if err != nil {
			log.Fatalf("Failed to connect to RabbitMQ: %s", err)
		}
		mq.Conn = conn
	})
	if mq.Channel == nil {
		ch, err := mq.Conn.Channel()
		if err != nil {
			log.Fatalf("Failed to open a RabbitMQ channel: %s", err)
		}
		mq.Channel = ch
		ch.ExchangeDeclare(os.Getenv("EXCHANGE"), "topic", true, false, false, false, nil)
		_, _ = ch.QueueDeclare(os.Getenv("QUEUE_NAME"), true, false, false, false, nil)
		// ch.QueueBind(q.Name, "transactions.*", os.Getenv("EXCHANGE"), false, nil)
		// ch.QueueBind(q.Name, "auth.*", os.Getenv("EXCHANGE"), false, nil)
		// ch.QueueBind(q.Name, "collection.*", os.Getenv("EXCHANGE"), false, nil)
		// ch.QueueBind(q.Name, "disbursement.*", os.Getenv("EXCHANGE"), false, nil)
		
		// Declare exclusive, auto-delete queue for responses unique to this worker connection instance
		qResp, err := ch.QueueDeclare(
			"",    // Let RabbitMQ generate a completely unique queue name
			false, // durable
			true,  // delete when unused (auto-delete)
			true,  // exclusive to this connection channel
			false, // no-wait
			nil,   // arguments
		)
		if err != nil {
			log.Fatal(err)
		}

		mq.ReplyToQueue = qResp.Name

		if err := mq.StartResponseListener(qResp.Name); err != nil {
			log.Fatalf("failed to start response listener: %v", err)
		}
	}
	return mq
}

// Emit publishes a message to the specified RabbitMQ queue
func (r *rmq) Emit(Event string, data interface{}) error {
	body, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("failed to marshal message: %v", err)
	}
	err = r.Channel.Publish(
		os.Getenv("EXCHANGE"),
		Event,
		false,
		false,
		amqp.Publishing{
			ContentType: "application/json",
			Body:        body,
		})
	if err != nil {
		return fmt.Errorf("failed to publish message: %v", err)
	}
	return nil
}
func (r *rmq) Request(event string, data any) ([]byte, error) {

	body, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}

	corrID := uuid.NewString()

	ch := make(chan []byte, 1)

	r.pending.Store(corrID, ch)
	defer r.pending.Delete(corrID)

	err = r.Channel.Publish(
		os.Getenv("EXCHANGE"),
		event,
		false,
		false,
		amqp.Publishing{
			ContentType:   "application/json",
			CorrelationId: corrID,
			ReplyTo:       r.ReplyToQueue,
			Body:          body,
		},
	)
	if err != nil {
		return nil, err
	}

	select {
	case res := <-ch:
		return res, nil

	case <-time.After(10 * time.Second):
		return nil, fmt.Errorf("timeout")
	}
}

func (r *rmq) StartResponseListener(queue string) error {
	log.Printf("🚀 RPC response listener started on exclusive queue: %s", queue)

	msgs, err := r.Channel.Consume(
		queue,
		"",
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		return err
	}

	go func() {
		for msg := range msgs {
			log.Printf("[MQ RPC Listener] Received response message (CorrelationId: %s)", msg.CorrelationId)
			if ch, ok := r.pending.Load(msg.CorrelationId); ok {
				select {
				case ch.(chan []byte) <- msg.Body:
					log.Printf("[MQ RPC Listener] Forwarded body to channel (CorrelationId: %s)", msg.CorrelationId)
				default:
					log.Printf("[MQ RPC Listener] Channel full or reader gone (CorrelationId: %s)", msg.CorrelationId)
				}
			} else {
				log.Printf("[MQ RPC Listener] No pending channel found for CorrelationId: %s", msg.CorrelationId)
			}
		}
	}()

	return nil
}

func (r *rmq) Consume(app *App, queueName string, reciever func(*App, amqp.Delivery) error) error {
	msgs, err := r.Channel.Consume(
		queueName,
		"",    // consumer tag
		true,  // auto-acknowledge
		false, // exclusive
		false, // no-local
		false, // no-wait
		nil,   // arguments
	)
	if err != nil {
		return fmt.Errorf("failed to consume messages: %v", err)
	}
	go func() {
		for msg := range msgs {
			if err := reciever(app, msg); err != nil {
				continue
			}
		}
	}()
	return nil
}
