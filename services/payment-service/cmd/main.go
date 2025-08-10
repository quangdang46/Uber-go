package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"ride-sharing/services/payment-service/internal/infrastructure/events"
	"ride-sharing/services/payment-service/internal/infrastructure/stripe"
	"ride-sharing/services/payment-service/internal/service"
	"ride-sharing/services/payment-service/pkg/types"
	"ride-sharing/shared/env"
	"ride-sharing/shared/messages"
	"ride-sharing/shared/tracing"
)

var GrpcAddr = env.GetString("GRPC_ADDR", ":9004")

func main() {
	shutdownJaeger, err := tracing.InitTracer(tracing.Config{
		ServiceName: "payment-service",
		Environment: env.GetString("ENVIRONMENT", "development"),
		JaegerEndpoint: env.GetString("JAEGER_ENDPOINT", "http://jaeger:14268/api/traces"),
	})
	if err != nil {
		log.Fatalf("Failed to initialize tracer: %v", err)
	}
	defer shutdownJaeger(context.Background())
	rabbitMqURI := env.GetString("RABBITMQ_URI", "amqp://guest:guest@rabbitmq:5672/")

	// Setup graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() {
		sigCh := make(chan os.Signal, 1)
		signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)
		<-sigCh
		cancel()
	}()

	appURL := env.GetString("APP_URL", "http://localhost:3000")

	// Stripe config
	stripeCfg := &types.PaymentConfig{
		StripeSecretKey: env.GetString("STRIPE_SECRET_KEY", ""),
		SuccessURL:      env.GetString("STRIPE_SUCCESS_URL", appURL+"?payment=success"),
		CancelURL:       env.GetString("STRIPE_CANCEL_URL", appURL+"?payment=cancel"),
	}

	if stripeCfg.StripeSecretKey == "" {
		log.Fatalf("STRIPE_SECRET_KEY is not set")
		return
	}

	paymentProcessor := stripe.NewStripeClient(stripeCfg)

	svc := service.NewPaymentService(paymentProcessor)

	log.Println("Payment processor created", svc)

	// RabbitMQ connection
	rabbitmq, err := messages.NewRabbitMQ(rabbitMqURI)
	if err != nil {
		log.Fatal(err)
	}

	defer rabbitmq.Close()

	log.Println("Starting RabbitMQ connection")
	tripConsumer := events.NewTripConsumer(rabbitmq, svc)
	go tripConsumer.Listen()
	// Wait for shutdown signal
	<-ctx.Done()
	log.Println("Shutting down payment service...")
}
