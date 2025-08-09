package events

import (
	"context"
	"encoding/json"
	"fmt"
	"ride-sharing/services/trip-service/internal/domain"
	"ride-sharing/shared/contracts"
	"ride-sharing/shared/messages"
)

type TripEventPublisher struct {
	rabbitmq *messages.RabbitMQ
}

func NewTripEventPublisher(rabbitmq *messages.RabbitMQ) *TripEventPublisher {
	return &TripEventPublisher{
		rabbitmq: rabbitmq,
	}
}

func (p *TripEventPublisher) PublishTripCreated(ctx context.Context, trip *domain.TripModel) error {

	// Kiểm tra connection health trước khi publish
	if !p.rabbitmq.IsConnectionHealthy() {
		fmt.Println("WARNING: RabbitMQ connection is not healthy, cannot publish trip created event")
		return fmt.Errorf("rabbitmq connection is not healthy")
	}

	fmt.Printf("TripEventPublisher: Publishing trip with ID: %s (ObjectID: %s)\n", trip.ID.Hex(), trip.ID.String())

	protoTrip := trip.ToProto()
	fmt.Printf("TripEventPublisher: Proto trip ID: %s\n", protoTrip.Id)

	payload := messages.TripEventData{
		Trip: protoTrip,
	}

	tripEventJSON, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	err = p.rabbitmq.PublishMessage(ctx, contracts.TripEventCreated, contracts.AmqpMessage{
		OwnerID: trip.UserID,
		Data:    tripEventJSON,
	})

	if err != nil {
		fmt.Printf("Failed to publish trip created event: %v\n", err)
		return err
	}

	fmt.Printf("Successfully published trip created event with ID: %s\n", protoTrip.Id)
	return nil
}
