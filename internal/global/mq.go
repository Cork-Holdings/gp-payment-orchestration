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
	Conn    *amqp.Connection
	Channel *amqp.Channel
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
		ch.ExchangeDeclare(os.Getenv("EXCHANGE"), "topic", true, false, false, false, nil)
		q, _ := ch.QueueDeclare(os.Getenv("QUEUE_NAME"), true, false, false, false, nil)
		ch.QueueBind(q.Name, "gateway.*", os.Getenv("EXCHANGE"), false, nil)

		mq.Channel = ch
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
		return nil, fmt.Errorf("failed to marshal message: %w", err)
	}

	// Temporary reply queue
	replyQueue, err := r.Channel.QueueDeclare(
		"",    // generated name
		false, // durable
		true,  // auto delete
		true,  // exclusive
		false,
		nil,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create reply queue: %w", err)
	}

	msgs, err := r.Channel.Consume(
		replyQueue.Name,
		"",
		true,
		true,
		false,
		false,
		nil,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to consume reply queue: %w", err)
	}

	correlationID := uuid.NewString()

	err = r.Channel.Publish(
		os.Getenv("EXCHANGE"),
		event,
		false,
		false,
		amqp.Publishing{
			ContentType:   "application/json",
			CorrelationId: correlationID,
			ReplyTo:       replyQueue.Name,
			Body:          body,
		},
	)
	if err != nil {
		return nil, fmt.Errorf("failed to publish request: %w", err)
	}

	timeout := time.After(10 * time.Second)

	for {
		select {
		case msg := <-msgs:
			if msg.CorrelationId == correlationID {
				return msg.Body, nil
			}

		case <-timeout:
			return nil, fmt.Errorf("request timed out after 10 seconds")
		}
	}
}
func (r *rmq) Consume(app *App, queueName string, reciever func(*App, amqp.Delivery) error) error {
	msgs, err := r.Channel.Consume(
		queueName,
		"",    // consumer tag
		false, // auto-acknowledge (manual ack)
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
			go func(m amqp.Delivery) {
				if err := reciever(app, m); err != nil {
					// On receiver error, send to DLQ (requeue=false)
					m.Nack(false, false)
					return
				}
				m.Ack(false)
			}(msg)
		}
	}()
	return nil
}
