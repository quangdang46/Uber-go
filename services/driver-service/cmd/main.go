package main

import (
	"context"
	"log"
	"net"
	"os"
	"os/signal"
	"ride-sharing/services/driver-service/internal/infrastructure/events"
	grpcHandler "ride-sharing/services/driver-service/internal/grpc"
	"ride-sharing/services/driver-service/internal/service"
	"ride-sharing/shared/env"
	"ride-sharing/shared/messages"
	"ride-sharing/shared/tracing"
	"syscall"

	grpcserver "google.golang.org/grpc"
)

var GrpcAddr = ":9092"

func main() {

	shutdownJaeger, err := tracing.InitTracer(tracing.Config{
		ServiceName: "driver-service",
		Environment: env.GetString("ENVIRONMENT", "development"),
		JaegerEndpoint: env.GetString("JAEGER_ENDPOINT", "http://jaeger:14268/api/traces"),
	})
	if err != nil {
		log.Fatalf("Failed to initialize tracer: %v", err)
	}
	defer shutdownJaeger(context.Background())

	log.Println("Starting driver-service...")

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() {
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
		<-sigChan
		log.Println("Received shutdown signal")
		cancel()
	}()

	lis, err := net.Listen("tcp", GrpcAddr)
	if err != nil {
		log.Fatalf("Failed to listen on %s: %v", GrpcAddr, err)
	}

	rabbitmqURI := env.GetString("RABBITMQ_URI", "amqp://guest:guest@rabbitmq:5672/")
	log.Printf("Connecting to RabbitMQ at %s", rabbitmqURI)

	rabbitmq, err := messages.NewRabbitMQ(rabbitmqURI)
	if err != nil {
		log.Fatalf("Failed to connect to RabbitMQ: %v", err)
	}
	defer func() {
		log.Println("Closing RabbitMQ connection")
		rabbitmq.Close()
	}()

	log.Println("Successfully connected to RabbitMQ")

	_service := service.NewService()
	// starting the gRpc server
	grpcServer := grpcserver.NewServer(tracing.WithTracingInterceptor()...)
	grpcHandler.NewGRPCHandler(grpcServer, _service)

	go func() {
		log.Println("Starting trip consumer...")
		if err := events.NewTripConsumer(rabbitmq, _service).Listen(); err != nil {
			log.Printf("Failed to listen to trip created: %v", err)
			cancel()
		}
	}()

	go func() {
		log.Printf("Starting gRPC server on %s", GrpcAddr)
		if err := grpcServer.Serve(lis); err != nil {
			log.Printf("Failed to serve gRPC: %v", err)
			cancel()
		}
	}()

	<-ctx.Done()

	log.Printf("Shutting down the server")
	grpcServer.GracefulStop()
}
