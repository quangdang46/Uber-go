package main

import (
	"context"
	"log"
	"net"
	"os"
	"os/signal"
	grpcHandler "ride-sharing/services/driver-service/internal/grpc"
	"ride-sharing/services/driver-service/internal/service"
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
