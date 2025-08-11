package messages

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"ride-sharing/shared/contracts"
	"ride-sharing/shared/retry"
	"ride-sharing/shared/tracing"

	amqp "github.com/rabbitmq/amqp091-go"
)

const (
	TripExchange = "trips"

	DeadLetterExchange = "dlx"
	DeadLetterQueue    = "dlq"
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

// IsConnectionHealthy kiểm tra xem connection có sẵn sàng không
func (r *RabbitMQ) IsConnectionHealthy() bool {
	return r.conn != nil && !r.conn.IsClosed() && r.Channel != nil
}

func (r *RabbitMQ) ConsumeMessages(queueName string, handler MessageHandler) error {

	if !r.IsConnectionHealthy() {
		return fmt.Errorf("rabbitmq connection/channel is not available for consuming")
	}

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
			if err := tracing.TracedConsumer(msg, func(ctx context.Context, msg amqp.Delivery) error {
				cfg := retry.DefaultConfig()

				err := retry.WithBackoff(ctx, cfg, func() error {
					if err := handler(context.Background(), msg); err != nil {
						log.Printf("failed to handle message: %v", err)
						// NACK message với requeue=false để không requeue infinitely
						_ = msg.Nack(false, false)
					} else {
						_ = msg.Ack(false)
					}
					return nil
				})

				if err != nil {
					log.Printf("failed to handle message: %v", err)
					header := amqp.Table{}
					if msg.Headers != nil {
						header = msg.Headers
					}

					header["x-death-reason"] = err.Error()
					header["x-origin-exchange"] = msg.Exchange
					header["x-origin-routing-key"] = msg.RoutingKey
					header["x-retry-count"] = cfg.MaxRetries
					msg.Headers = header
					_ = msg.Reject(false)

					return err
				}

				if err := msg.Ack(false); err != nil {
					log.Printf("failed to ack message: %v", err)
				}

				return nil
			}); err != nil {
				log.Printf("failed to trace consumer: %v", err)
			}
		}
	}()

	return nil

}

func (r *RabbitMQ) PublishMessage(ctx context.Context, routingKey string, message contracts.AmqpMessage) error {

	jsonMsg, err := json.Marshal(message)
	if err != nil {
		return fmt.Errorf("failed to marshal message: %v", err)
	}

	// Kiểm tra connection trước khi publish
	if !r.IsConnectionHealthy() {
		return fmt.Errorf("rabbitmq connection/channel is not available")
	}

	msg := amqp.Publishing{
		ContentType:  "application/json",
		Body:         jsonMsg,
		DeliveryMode: amqp.Persistent,
	}

	log.Println("publishing message to ", routingKey)
	return tracing.TracedPublisher(ctx,
		TripExchange,
		routingKey,
		msg,
		r.publish,
	)
}

func (r *RabbitMQ) setupExchangesAndQueues() error {

	// dlq queue

	if err := r.setupDeadLetterQueue(); err != nil {
		return fmt.Errorf("failed to setup dead letter queue: %v", err)
	}

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

	err = r.declareAndBindQueue(
		DriverTripResponseQueue,
		[]string{
			contracts.DriverCmdTripAccept,
			contracts.DriverCmdTripDecline,
		},
		TripExchange,
	)

	if err != nil {
		return fmt.Errorf("failed to bind queue: %v", err)
	}

	err = r.declareAndBindQueue(
		NotifyDriverNoDriversFoundQueue,
		[]string{
			contracts.TripEventNoDriversFound,
		},
		TripExchange,
	)

	if err != nil {
		return fmt.Errorf("failed to bind queue: %v", err)
	}

	err = r.declareAndBindQueue(
		NotifyDriverAssignQueue,
		[]string{
			contracts.TripEventDriverAssigned,
		},
		TripExchange,
	)

	if err != nil {
		return fmt.Errorf("failed to bind queue: %v", err)
	}

	if err := r.declareAndBindQueue(
		PaymentTripResponseQueue,
		[]string{contracts.PaymentCmdCreateSession},
		TripExchange,
	); err != nil {
		return err
	}

	if err := r.declareAndBindQueue(
		NotifyPaymentSessionCreatedQueue,
		[]string{contracts.PaymentEventSessionCreated},
		TripExchange,
	); err != nil {
		return err
	}

	if err := r.declareAndBindQueue(
		NotifyPaymentSuccessQueue,
		[]string{contracts.PaymentEventSuccess},
		TripExchange,
	); err != nil {
		return err
	}

	return nil
}

func (r *RabbitMQ) declareAndBindQueue(queueName string, messageType []string, exchange string) error {

	args := amqp.Table{
		"x-dead-letter-exchange":    DeadLetterExchange,
		"x-dead-letter-routing-key": queueName,
	}

	q, err := r.Channel.QueueDeclare(
		queueName,
		true,
		false,
		false,
		false,
		args, // arguments
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

func (r *RabbitMQ) publish(ctx context.Context, exchange, routingKey string, message amqp.Publishing) error {

	return r.Channel.PublishWithContext(ctx, exchange, routingKey, false, false, message)

}

func (r *RabbitMQ) setupDeadLetterQueue() error {
	err := r.Channel.ExchangeDeclare(
		DeadLetterExchange,
		"topic",
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		return fmt.Errorf("failed to declare dead letter exchange: %v", err)
	}
	q, err := r.Channel.QueueDeclare(
		DeadLetterQueue,
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		return fmt.Errorf("failed to declare dead letter queue: %v", err)
	}

	err = r.Channel.QueueBind(
		q.Name,
		"#",
		DeadLetterExchange,
		false,
		nil,
	)
	if err != nil {
		return fmt.Errorf("failed to bind dead letter queue: %v", err)
	}

	return nil
}
