package grpc_clients

import (
	"os"
	pb "ride-sharing/shared/proto/trip"
	"ride-sharing/shared/tracing"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type tripServiceClient struct {
	Client pb.TripServiceClient
	conn   *grpc.ClientConn
}

func NewTripServiceClient() (*tripServiceClient, error) {

	tripServiceUrl := os.Getenv("TRIP_SERVICE_URL")
	if tripServiceUrl == "" {
		tripServiceUrl = "trip-service:9093"
	}

	dialOptions := append(tracing.DialOptionWithTracing(), grpc.WithTransportCredentials(insecure.NewCredentials()))
	conn, err := grpc.NewClient(tripServiceUrl, dialOptions...)

	if err != nil {
		return nil, err
	}

	client := pb.NewTripServiceClient(conn)

	return &tripServiceClient{
		conn:   conn,
		Client: client,
	}, nil

}

func (c *tripServiceClient) Close() {
	if c.conn != nil {
		if err := c.conn.Close(); err != nil {
			return
		}
	}
}
