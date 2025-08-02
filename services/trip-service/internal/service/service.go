package service

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"ride-sharing/services/trip-service/internal/domain"
	tripTypes "ride-sharing/services/trip-service/pkg/types"
	"ride-sharing/shared/types"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type service struct {
	repo domain.TripRepository
}

func NewService(repo domain.TripRepository) *service {
	return &service{
		repo: repo,
	}
}

func (s *service) CreateTrip(ctx context.Context, fare *domain.RideFareModel) (*domain.TripModel, error) {
	t := &domain.TripModel{
		ID:       primitive.NewObjectID(),
		UserId:   fare.UserID,
		Status:   "pending",
		RideFare: fare,
	}

	return s.repo.CreateTrip(ctx, t)
}

func (s *service) GetRoute(ctx context.Context, pickup, destination *types.Coordinate) (*tripTypes.OsrmApiResponse, error) {
	baseURL := "http://router.project-osrm.org"

	url := fmt.Sprintf(
		"%s/route/v1/driving/%f,%f;%f,%f?overview=full&geometries=geojson",
		baseURL,
		pickup.Longitude, pickup.Latitude,
		destination.Longitude, destination.Latitude,
	)
	res, err := http.Get(url)

	if err != nil {
		return nil, fmt.Errorf("failed to fetch %v", err)

	}

	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)

	if err != nil {
		return nil, fmt.Errorf("failed to read the response %v", err)
	}

	var routeRes tripTypes.OsrmApiResponse

	if err := json.Unmarshal(body, &routeRes); err != nil {
		return nil, fmt.Errorf("failed parse res %v", err)
	}

	return &routeRes, nil

}
