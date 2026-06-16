package main

import (
	"crypto/tls"
	"log"

	amqp "github.com/rabbitmq/amqp091-go"
)

func main() {
	var conn *amqp.Connection
	var err error

	// We will try AMQPS on 5671 (standard for secure brokers)
	secureURL := "amqps://L5sJpllqgI4Taxxw:ycVeIg7qQCroSYOeFpQgbSEcg1bHOIKu@rabbitmq.mygeepay.com:5671/"
	log.Printf("Attempting secure AMQPS connection to rabbitmq.mygeepay.com:5671...")
	
	// Skip verification to prevent certificate handshake failure on self-signed/internal certs
	tlsConfig := &tls.Config{InsecureSkipVerify: true}
	conn, err = amqp.DialTLS(secureURL, tlsConfig)
	if err != nil {
		log.Printf("DialTLS failed: %v. Retrying with standard AMQP on 5672...", err)
		
		// Fallback to standard AMQP on 5672
		fallbackURL := "amqp://L5sJpllqgI4Taxxw:ycVeIg7qQCroSYOeFpQgbSEcg1bHOIKu@rabbitmq.mygeepay.com:5672/"
		conn, err = amqp.Dial(fallbackURL)
		if err != nil {
			log.Fatalf("Both secure DialTLS and fallback Dial failed: %v", err)
		}
	}
	defer conn.Close()
	log.Println("Successfully connected to RabbitMQ broker!")

	ch, err := conn.Channel()
	if err != nil {
		log.Fatalf("Failed to open a channel: %v", err)
	}
	defer ch.Close()

	// 1. Declare Topic Exchange for core events
	exchangeName := "events.topic"
	log.Printf("Declaring topic exchange: %s...", exchangeName)
	err = ch.ExchangeDeclare(
		exchangeName,
		"topic",
		true,  // durable
		false, // auto-deleted
		false, // internal
		false, // no-wait
		nil,   // arguments
	)
	if err != nil {
		log.Fatalf("Failed to declare exchange %s: %v", exchangeName, err)
	}

	// 2. Declare Dead-Letter Exchange (DLX)
	dlxName := "events.dlx"
	log.Printf("Declaring dead-letter exchange: %s...", dlxName)
	err = ch.ExchangeDeclare(
		dlxName,
		"topic",
		true,  // durable
		false, // auto-deleted
		false, // internal
		false, // no-wait
		nil,   // arguments
	)
	if err != nil {
		log.Fatalf("Failed to declare DLX %s: %v", dlxName, err)
	}

	// 3. Declare Queues with DLX configuration
	queues := []struct {
		Name       string
		DLK        string
		RoutingKey string
		IsDLQ      bool
	}{
		// Main queues
		{Name: "reporting.transactions", DLK: "dlq.transaction.completed", RoutingKey: "transaction.*", IsDLQ: false},
		{Name: "webhook.transactions", DLK: "dlq.transaction.completed", RoutingKey: "transaction.completed", IsDLQ: false},
		{Name: "reporting.settlements", DLK: "dlq.settlement.completed", RoutingKey: "settlement.*", IsDLQ: false},
		
		// Dead Letter queues
		{Name: "dlq.transactions", RoutingKey: "dlq.transaction.*", IsDLQ: true},
		{Name: "dlq.settlements", RoutingKey: "dlq.settlement.*", IsDLQ: true},
	}

	for _, q := range queues {
		var args amqp.Table
		if !q.IsDLQ {
			// Attach DLX to redirect rejected messages
			args = amqp.Table{
				"x-dead-letter-exchange":    dlxName,
				"x-dead-letter-routing-key": q.DLK,
			}
		}

		log.Printf("Declaring queue: %s...", q.Name)
		_, err := ch.QueueDeclare(
			q.Name,
			true,  // durable
			false, // delete when unused
			false, // exclusive
			false, // no-wait
			args,  // arguments
		)
		if err != nil {
			log.Fatalf("Failed to declare queue %s: %v", q.Name, err)
		}

		// Bind queue
		targetExchange := exchangeName
		if q.IsDLQ {
			targetExchange = dlxName
		}
		
		log.Printf("Binding queue %s to exchange %s with routing key %s...", q.Name, targetExchange, q.RoutingKey)
		err = ch.QueueBind(
			q.Name,
			q.RoutingKey,
			targetExchange,
			false,
			nil,
		)
		if err != nil {
			log.Fatalf("Failed to bind queue %s: %v", q.Name, err)
		}
	}

	log.Println("RabbitMQ topology setup completed successfully!")
}
