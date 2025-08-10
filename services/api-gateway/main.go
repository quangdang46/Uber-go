package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"ride-sharing/shared/env"
	"ride-sharing/shared/messages"
	"ride-sharing/shared/tracing"
)

var (
	httpAddr = env.GetString("HTTP_ADDR", ":8081")

	rabbitmqURI = env.GetString("RABBITMQ_URI", "amqp://guest:guest@rabbitmq:5672/")
)

func main() {

	shutdownJaeger, err := tracing.InitTracer(tracing.Config{
		ServiceName:    "api-gateway",
		Environment:    env.GetString("ENVIRONMENT", "development"),
		JaegerEndpoint: env.GetString("JAEGER_ENDPOINT", "http://jaeger:14268/api/traces"),
	})
	if err != nil {
		log.Fatalf("Failed to initialize tracer: %v", err)
	}
	defer shutdownJaeger(context.Background())

	log.Println("Starting API Gateway")

	mux := http.NewServeMux()
	log.Printf("Connecting to RabbitMQ at %s", rabbitmqURI)

	rabbitmq, err := messages.NewRabbitMQ(rabbitmqURI)
	if err != nil {
		log.Fatalf("Failed to connect to RabbitMQ: %v", err)
	}
	defer func() {
		log.Println("Closing RabbitMQ connection")
		rabbitmq.Close()
	}()

	mux.HandleFunc("POST /trip/preview", tracing.WrapHandlerFunc(enableCORS(handleTripPreview), "/trip/preview"))
	mux.HandleFunc("POST /trip/start", tracing.WrapHandlerFunc(enableCORS(handleTripStart), "/trip/start"))

	mux.HandleFunc("/webhook/stripe", tracing.WrapHandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		handleStripeWebhook(w, r, rabbitmq)
	}, "/webhook/stripe"))
	mux.HandleFunc("/ws/drivers", tracing.WrapHandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		handleDriversWebSocket(w, r, rabbitmq)
	}, "/ws/drivers"))
	mux.HandleFunc("/ws/riders", tracing.WrapHandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		handleRidersWebSocket(w, r, rabbitmq)
	}, "/ws/riders"))

	server := &http.Server{
		Handler: mux,
		Addr:    httpAddr,
	}

	serverError := make(chan error, 1)
	go func() {
		log.Printf("Server listening on %s", httpAddr)
		serverError <- server.ListenAndServe()
	}()

	shutdown := make(chan os.Signal, 1)

	signal.Notify(shutdown, os.Interrupt, syscall.SIGTERM)

	select {

	case err := <-serverError:
		log.Printf("Error starting the server :%v", err)
	case sig := <-shutdown:
		log.Printf("Server is shutting down due to %v signal", sig)

		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if err := server.Shutdown(ctx); err != nil {
			log.Printf("Could not stop the server %v", err)
			server.Close()
		}

	}

}
