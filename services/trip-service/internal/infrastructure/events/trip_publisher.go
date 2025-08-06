package events

import (
	"context"
	"encoding/json"
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
	payload := messages.TripEventData{
		Trip: trip.ToProto(),
	}

	tripEventJSON, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	return p.rabbitmq.PublishMessage(ctx, contracts.TripEventCreated, contracts.AmqpMessage{
		OwnerID: trip.UserID,
		Data:    tripEventJSON,
	})

}
