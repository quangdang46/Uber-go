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
	log.Printf("Starting to listen on queue: %s", messages.DriverTripResponseQueue)

	// Kiểm tra connection health trước khi start consumer
	if !c.rabbitmq.IsConnectionHealthy() {
		return fmt.Errorf("rabbitmq connection is not healthy, cannot start driver consumer")
	}

	return c.rabbitmq.ConsumeMessages(messages.DriverTripResponseQueue, func(ctx context.Context, msg amqp.Delivery) error {
		log.Printf("DriverConsumer received message with routing key: %s", msg.RoutingKey)

		var tripEvent contracts.AmqpMessage
		if err := json.Unmarshal(msg.Body, &tripEvent); err != nil {
			log.Printf("Failed to unmarshal AmqpMessage: %v", err)
			return err
		}
		log.Printf("Parsed AmqpMessage: OwnerID=%s", tripEvent.OwnerID)

		var payload messages.DriverTripResponseData
		if err := json.Unmarshal(tripEvent.Data, &payload); err != nil {
			log.Printf("Failed to unmarshal DriverTripResponseData: %v", err)
			return err
		}
		log.Printf("Parsed payload: TripID=%s, RiderID=%s, DriverID=%s", payload.TripID, payload.RiderID, payload.Driver.Id)

		switch msg.RoutingKey {
		case contracts.DriverCmdTripAccept:
			log.Printf("Processing trip accept for TripID: %s", payload.TripID)
			if err := c.handleTripAccept(ctx, payload.TripID, &payload.Driver); err != nil {
				log.Printf("Failed to handle trip accept: %v", err)
				// ACK message để tránh infinite requeue, nhưng log error để debug
				log.Printf("ACKing failed trip accept message to prevent requeue loop")
				return nil // ACK the message instead of returning error
			}
			log.Printf("Successfully processed trip accept for TripID: %s", payload.TripID)
		case contracts.DriverCmdTripDecline:
			log.Printf("Processing trip decline for TripID: %s", payload.TripID)
			if err := c.handleTripDecline(ctx, payload.TripID, payload.RiderID); err != nil {
				log.Printf("Failed to handle trip decline: %v", err)
				// ACK message để tránh infinite requeue, nhưng log error để debug
				log.Printf("ACKing failed trip decline message to prevent requeue loop")
				return nil // ACK the message instead of returning error
			}
			log.Printf("Successfully processed trip decline for TripID: %s", payload.TripID)
		default:
			log.Printf("Unknown routing key: %s", msg.RoutingKey)
		}

		return nil
	})
}

func (c *DriverConsumer) handleTripAccept(ctx context.Context, tripID string, driver *pbDriver.Driver) error {
	log.Printf("handleTripAccept: Looking for trip with ID: %s", tripID)

	trip, err := c.service.GetTripByID(ctx, tripID)
	if err != nil {
		log.Printf("handleTripAccept: Error getting trip: %v", err)
		return err
	}

	if trip == nil {
		log.Printf("handleTripAccept: Trip not found with ID: %s", tripID)
		// Log tất cả trips hiện có để debug
		log.Printf("handleTripAccept: DEBUG - Trip not found, this might be a race condition or ID mismatch")
		return fmt.Errorf("trip does not exist with ID: %s", tripID)
	}

	log.Printf("handleTripAccept: Found trip with ID: %s, Status: %s", tripID, trip.Status)

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

	marshalledPayload, err := json.Marshal(messages.PaymentTripResponseData{
		TripID:   tripID,
		UserID:   trip.UserID,
		DriverID: driver.Id,
		Amount:   trip.RideFare.TotalPriceInCents,
		Currency: "USD",
	})

	fmt.Println("====>marshalledPayload", marshalledPayload)
	if err := c.rabbitmq.PublishMessage(ctx, contracts.PaymentCmdCreateSession,
		contracts.AmqpMessage{
			OwnerID: trip.UserID,
			Data:    marshalledPayload,
		},
	); err != nil {
		fmt.Println("====>err", err)
		return err
	}

	fmt.Println("====>published message")

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
