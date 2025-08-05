package main

import (
	"context"
	"log"
	"net"
	"os"
	"os/signal"
	grpcHandler "ride-sharing/services/driver-service/internal/grpc"
	"ride-sharing/services/driver-service/internal/service"
	"ride-sharing/shared/env"
	"ride-sharing/shared/messages"
	"syscall"

	grpcserver "google.golang.org/grpc"
)

var GrpcAddr = ":9093"

func main() {

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() {
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
		<-sigChan
		cancel()
	}()

	lis, err := net.Listen("tcp", GrpcAddr)

	if err != nil {
		log.Fatal("failed to listen %v", err)
	}

	rabbit, err := messages.NewRabbitMQ(env.GetString("RABBITMQ_URI", "amqp://guest:guest@localhost:5672/"))

	if err != nil {
		log.Fatal("failed to connect rabbitmq")
	}
	defer rabbit.Close()

	_service := service.NewService()
	// starting the gRpc server
	grpcServer := grpcserver.NewServer()
	grpcHandler.NewGRPCHandler(grpcServer, _service)

	go func() {
		if err := grpcServer.Serve(lis); err != nil {
			log.Printf("failed to server %v", err)
			cancel()
		}
	}()

	<-ctx.Done()

	log.Printf("shutting down the server")

	grpcServer.GracefulStop()

}
