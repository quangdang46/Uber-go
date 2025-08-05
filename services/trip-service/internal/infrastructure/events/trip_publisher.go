package events

import (
	"context"
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

func (p *TripEventPublisher) PublishTripCreated(ctx context.Context) error {
	err := p.rabbitmq.PublishMessage(ctx, "hello", "trip created")
	if err != nil {
		return err
	}

	return nil
}
