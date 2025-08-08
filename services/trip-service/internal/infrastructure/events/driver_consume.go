package events

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"ride-sharing/services/trip-service/internal/service"
	"ride-sharing/shared/contracts"
	"ride-sharing/shared/messages"

	amqp "github.com/rabbitmq/amqp091-go"

	pbDriver "ride-sharing/shared/proto/driver"
)

type DriverConsumer struct {
	rabbitmq *messages.RabbitMQ
	service  *service.TripService
}

func NewDriverConsumer(rabbitmq *messages.RabbitMQ, service *service.TripService) *DriverConsumer {
	return &DriverConsumer{
		rabbitmq: rabbitmq,
		service:  service,
	}
}

func (c *DriverConsumer) Listen() error {
	return c.rabbitmq.ConsumeMessages(messages.DriverTripResponseQueue, func(ctx context.Context, msg amqp.Delivery) error {
		var tripEvent contracts.AmqpMessage
		if err := json.Unmarshal(msg.Body, &tripEvent); err != nil {
			return err
		}
		log.Println("received message: ", tripEvent)
		var payload messages.DriverTripResponseData
		if err := json.Unmarshal(tripEvent.Data, &payload); err != nil {
			return err
		}
		log.Println("received message: ", payload)

		switch msg.RoutingKey {
		case contracts.DriverCmdTripAccept:
			if err := c.handleTripAccept(ctx, payload.TripID, &payload.Driver); err != nil {
				return err
			}
		case contracts.DriverCmdTripDecline:
			if err := c.handleTripDecline(ctx, payload.TripID, payload.RiderID); err != nil {
				return err
			}
		}

		return nil
	})
}

func (c *DriverConsumer) handleTripAccept(ctx context.Context, tripID string, driver *pbDriver.Driver) error {

	trip, err := c.service.GetTripByID(ctx, tripID)
	if err != nil {
		return err
	}

	if trip == nil {
		return fmt.Errorf("trip does not exist")
	}

	if err := c.service.UpdateTrip(ctx, tripID, "accepted", driver); err != nil {
		return err
	}

	marshallTrip, err := json.Marshal(trip)
	if err != nil {
		return err
	}

	amqpMessage := contracts.AmqpMessage{
		OwnerID: trip.UserID,
		Data:    marshallTrip,
	}

	if err := c.rabbitmq.PublishMessage(ctx, contracts.TripEventDriverAssigned, amqpMessage); err != nil {
		return err
	}

	// payment service

	return nil
}

func (c *DriverConsumer) handleTripDecline(ctx context.Context, tripID, riderID string) error {

	trip, err := c.service.GetTripByID(ctx, tripID)
	if err != nil {
		return err
	}

	if trip == nil {
		return fmt.Errorf("trip does not exist")
	}

	if err := c.service.UpdateTrip(ctx, tripID, "declined", nil); err != nil {
		return err
	}

	marshallPayload, err := json.Marshal(trip)

	if err != nil {
		return err
	}

	amqpMessage := contracts.AmqpMessage{
		OwnerID: riderID,
		Data:    marshallPayload,
	}

	if err := c.rabbitmq.PublishMessage(ctx, contracts.TripEventDriverNotInterested, amqpMessage); err != nil {
		return err
	}

	return nil
}
