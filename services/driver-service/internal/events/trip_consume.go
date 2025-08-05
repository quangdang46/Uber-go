package events

import (
	"context"
	"log"
	"ride-sharing/shared/messages"

	amqp "github.com/rabbitmq/amqp091-go"
)

type TripConsumer struct {
	rabbitmq *messages.RabbitMQ
}

func NewTripConsumer(rabbitmq *messages.RabbitMQ) *TripConsumer {
	return &TripConsumer{
		rabbitmq: rabbitmq,
	}
}


func (c *TripConsumer) Listen() error {
	return c.rabbitmq.ConsumeMessages("hello", func(ctx context.Context, msg amqp.Delivery) error {
		log.Println("received message: ", string(msg.Body))
		return nil
	})
}