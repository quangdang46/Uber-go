package events

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"ride-sharing/services/driver-service/internal/service"
	"ride-sharing/shared/contracts"
	"ride-sharing/shared/messages"

	amqp "github.com/rabbitmq/amqp091-go"
)

type TripConsumer struct {
	rabbitmq *messages.RabbitMQ
	service  *service.Service
}

func NewTripConsumer(rabbitmq *messages.RabbitMQ, service *service.Service) *TripConsumer {
	return &TripConsumer{
		rabbitmq: rabbitmq,
		service:  service,
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

		switch msg.RoutingKey {
		case contracts.TripEventCreated, contracts.TripEventDriverNotInterested:
			return c.handleFindAndNotifyDriver(ctx, payload)
		}

		log.Println("received message: ", msg.RoutingKey)

		return nil
	})
}

func (c *TripConsumer) handleFindAndNotifyDriver(ctx context.Context, payload messages.TripEventData) error {
	suitableIDs := c.service.FindAvailableDrivers(payload.Trip.SelectedFare.PackageSlug)
	fmt.Println("suitableIDs: ", suitableIDs)
	if len(suitableIDs) == 0 {
		if err := c.rabbitmq.PublishMessage(ctx, contracts.TripEventNoDriversFound, contracts.AmqpMessage{
			OwnerID: payload.Trip.UserID,
		}); err != nil {
			return err
		}
		return nil
	}

	suitableDriverID := suitableIDs[0]

	marshalledEvent, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	if err := c.rabbitmq.PublishMessage(ctx, contracts.DriverCmdTripRequest, contracts.AmqpMessage{
		OwnerID: suitableDriverID,
		Data:    marshalledEvent,
	}); err != nil {
		return err
	}

	return nil
}
