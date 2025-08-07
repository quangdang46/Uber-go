package messages

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"ride-sharing/shared/contracts"

	amqp "github.com/rabbitmq/amqp091-go"
)

const (
	TripExchange = "trips"
)

type RabbitMQ struct {
	conn    *amqp.Connection
	Channel *amqp.Channel
}

func NewRabbitMQ(uri string) (*RabbitMQ, error) {
	conn, err := amqp.Dial(uri)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to RabbitMQ: %v", err)
	}
	ch, err := conn.Channel()

	if err != nil {
		return nil, fmt.Errorf("failed to created channel: %v", err)
	}

	rmq := &RabbitMQ{
		conn:    conn,
		Channel: ch,
	}

	if err := rmq.setupExchangesAndQueues(); err != nil {
		rmq.Close()
		return nil, fmt.Errorf("failed to setup exchange: %v", err)

	}

	return rmq, nil

}

type MessageHandler func(ctx context.Context, msg amqp.Delivery) error

func (r *RabbitMQ) ConsumeMessages(queueName string, handler MessageHandler) error {

	err := r.Channel.Qos(1, 0, false)
	if err != nil {
		return fmt.Errorf("failed to consume messages: %v", err)
	}

	msgs, err := r.Channel.Consume(
		queueName,
		"",
		false,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		return fmt.Errorf("failed to consume messages: %v", err)
	}

	go func() {
		for msg := range msgs {
			if err := handler(context.Background(), msg); err != nil {
				log.Printf("failed to handle message: %v", err)
				continue
			}
			_ = msg.Ack(false)
		}
	}()

	return nil

}

func (r *RabbitMQ) PublishMessage(ctx context.Context, routingKey string, message contracts.AmqpMessage) error {

	jsonMsg, err := json.Marshal(message)
	if err != nil {
		return fmt.Errorf("failed to marshal message: %v", err)
	}

	log.Println("publishing message to ", routingKey)
	return r.Channel.PublishWithContext(ctx,
		TripExchange,
		routingKey,
		false,
		false,
		amqp.Publishing{
			ContentType:  "application/json",
			Body:         jsonMsg,
			DeliveryMode: amqp.Persistent,
		},
	)
}

func (r *RabbitMQ) setupExchangesAndQueues() error {
	err := r.Channel.ExchangeDeclare(
		TripExchange, // exchange name
		"topic",      // exchange type
		true,         // durable
		false,        // autoDelete
		false,        // internal
		false,        // noWait
		nil,          // arguments
	)

	if err != nil {
		return fmt.Errorf("failed to declare exchange: %v", err)
	}

	err = r.declareAndBindQueue(
		FindAvailableDriversQueue,
		[]string{
			contracts.TripEventCreated,
			contracts.TripEventDriverNotInterested,
		},
		TripExchange,
	)
	if err != nil {
		return fmt.Errorf("failed to bind queue: %v", err)
	}

	err = r.declareAndBindQueue(
		DriverCmdTripRequestQueue,
		[]string{
			contracts.DriverCmdTripRequest,
		},
		TripExchange,
	)

	if err != nil {
		return fmt.Errorf("failed to bind queue: %v", err)
	}

	return err
}

func (r *RabbitMQ) declareAndBindQueue(queueName string, messageType []string, exchange string) error {
	q, err := r.Channel.QueueDeclare(
		queueName,
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		return fmt.Errorf("failed to bind queue: %v", err)
	}

	for _, messageType := range messageType {
		err = r.Channel.QueueBind(
			q.Name,
			messageType,
			exchange,
			false,
			nil,
		)
		if err != nil {
			return fmt.Errorf("failed to bind queue: %v", err)
		}
	}

	return nil
}

func (r *RabbitMQ) Close() {
	if r.conn != nil {
		r.conn.Close()
	}
	if r.Channel != nil {
		r.Channel.Close()
	}
}
