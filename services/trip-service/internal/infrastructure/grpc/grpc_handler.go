package grpc

import (
	"context"
	"fmt"
	"log"
	"ride-sharing/services/trip-service/internal/domain"
	"ride-sharing/services/trip-service/internal/infrastructure/events"
	pb "ride-sharing/shared/proto/trip"
	"ride-sharing/shared/types"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type gRPCHandler struct {
	pb.UnimplementedTripServiceServer
	service   domain.TripService
	publisher *events.TripEventPublisher
}

func (h *gRPCHandler) PreviewTrip(ctx context.Context, req *pb.PreviewTripRequest) (*pb.PreviewTripResponse, error) {
	pickup := req.GetStartLocation()
	destination := req.GetEndLocation()

	pickupCoordinate := &types.Coordinate{
		Latitude:  pickup.Latitude,
		Longitude: pickup.Longitude,
	}
	destinationCoordinate := &types.Coordinate{
		Latitude:  destination.Latitude,
		Longitude: destination.Longitude,
	}

	route, err := h.service.GetRoute(ctx, pickupCoordinate, destinationCoordinate)

	if err != nil {
		log.Println(err)
		return nil, status.Errorf(codes.Internal, "failed to get routes %v", err)
	}

	userID := req.GetUserID()

	estimatedFares := h.service.EstimatePackagesPriceWithRoute(route)
	fares, err := h.service.GenerateTripFares(ctx, estimatedFares, userID, route)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to generate the ride fares %v", err)

	}


	return &pb.PreviewTripResponse{
		Route:     route.ToProto(),
		RideFares: domain.ToRideFaresProto(fares),
	}, nil

}

// ///////////////
func NewGRPCHandler(server *grpc.Server, service domain.TripService, publisher *events.TripEventPublisher) *gRPCHandler {
	handler := &gRPCHandler{
		service:   service,
		publisher: publisher,
	}

	pb.RegisterTripServiceServer(server, handler)

	return handler
}

/////////////////

func (h *gRPCHandler) CreateTrip(ctx context.Context, req *pb.CreateTripRequest) (*pb.CreateTripResponse, error) {

	fareID := req.GetRideFareID()
	userID := req.GetUserID()
	fmt.Println("fareID",fareID)
	fmt.Println("userID",userID)

	rideFare, err := h.service.GetAndValidateFare(ctx, fareID, userID)
	fmt.Println("rideFare",rideFare)

	if err != nil {
		fmt.Println("error in get and validate fare",err)
		return nil, status.Errorf(codes.Internal, "failed to created fare %v", err)

	}

	trip, err := h.service.CreateTrip(ctx, rideFare)

	if err != nil {
		fmt.Println("error in create trip",err)
		return nil, status.Errorf(codes.Internal, "failed to created trip %v", err)

	}

	if err := h.publisher.PublishTripCreated(ctx, trip); err != nil {
		fmt.Println("error in publish trip created",err)
		return nil, status.Errorf(codes.Internal, "failed to publish trip created %v", err)
	}

	return &pb.CreateTripResponse{
		TripID: string(trip.ID.Hex()),
	}, nil

}
