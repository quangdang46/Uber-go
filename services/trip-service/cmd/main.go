package main

import (
	"context"
	"log"
	"net"
	"os"
	"os/signal"
	"ride-sharing/services/trip-service/internal/infrastructure/events"
	"ride-sharing/services/trip-service/internal/infrastructure/grpc"
	"ride-sharing/services/trip-service/internal/infrastructure/repository"
	"ride-sharing/services/trip-service/internal/service"
	"ride-sharing/shared/env"
	"ride-sharing/shared/messages"
	"syscall"

	grpcserver "google.golang.org/grpc"
)

var GrpcAddr = ":9093"

func main() {
	log.Println("Starting trip-service...")

	immemRepo := repository.NewInMemRepository()
	svc := service.NewTripService(immemRepo)

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

	// starting the gRpc server
	publisher := events.NewTripEventPublisher(rabbitmq)
	driverConsumer := events.NewDriverConsumer(rabbitmq, svc)
	paymentConsumer := events.NewPaymentConsumer(rabbitmq, svc)
	grpcServer := grpcserver.NewServer()
	grpc.NewGRPCHandler(grpcServer, svc, publisher)

	go func() {
		log.Printf("Starting gRPC server on %s", GrpcAddr)
		if err := grpcServer.Serve(lis); err != nil {
			log.Printf("Failed to serve gRPC: %v", err)
			cancel()
		}
	}()

	go func() {
		log.Println("Starting driver consumer...")
		if err := driverConsumer.Listen(); err != nil {
			log.Printf("Failed to start driver consumer: %v", err)
			cancel()
		}
		log.Println("Driver consumer started successfully")
	}()

	go func() {
		log.Println("Starting payment consumer")
		if err := paymentConsumer.Listen(); err != nil {
			log.Printf("Failed to start payment consumer: %v", err)
			cancel()
		}
	}()

	<-ctx.Done()

	log.Printf("Shutting down the server")
	grpcServer.GracefulStop()
}
