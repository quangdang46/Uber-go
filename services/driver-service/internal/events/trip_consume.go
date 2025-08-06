package events

import (
	"context"
	"encoding/json"
	"log"
	"ride-sharing/shared/contracts"
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
	return c.rabbitmq.ConsumeMessages(messages.FindAvailableDriversQueue, func(ctx context.Context, msg amqp.Delivery) error {
		var tripEvent contracts.AmqpMessage
		if err := json.Unmarshal(msg.Body, &tripEvent); err != nil {
			return err
		}
		log.Println("received message: ", tripEvent)
		var payload messages.TripEventData
		if err := json.Unmarshal(tripEvent.Data, &payload); err != nil {
			return err
		}
		log.Println("received message: ", payload)


		return nil
	})
}
