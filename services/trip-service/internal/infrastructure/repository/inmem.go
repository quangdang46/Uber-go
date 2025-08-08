package repository

import (
	"context"
	"fmt"
	"ride-sharing/services/trip-service/internal/domain"

	pbDriver "ride-sharing/shared/proto/driver"
	pb "ride-sharing/shared/proto/trip"
)

type inmemRepository struct {
	trips map[string]*domain.TripModel

	rideFares map[string]*domain.RideFareModel
}

func NewInMemRepository() *inmemRepository {

	return &inmemRepository{
		trips:     make(map[string]*domain.TripModel),
		rideFares: make(map[string]*domain.RideFareModel),
	}

}

func (r *inmemRepository) CreateTrip(ctx context.Context, trip *domain.TripModel) (*domain.TripModel, error) {

	r.trips[trip.ID.Hex()] = trip

	return trip, nil
}

func (r *inmemRepository) SaveRideFare(ctx context.Context, f *domain.RideFareModel) error {
	r.rideFares[f.ID.Hex()] = f
	return nil
}

func (r *inmemRepository) GetRideFareByID(ctx context.Context, id string) (*domain.RideFareModel, error) {
	fare, exist := r.rideFares[id]
	if !exist {
		return nil, fmt.Errorf("fare does not exist with ID: %s", id)
	}

	return fare, nil
}

func (r *inmemRepository) GetTripByID(ctx context.Context, id string) (*domain.TripModel, error) {
	trip, exist := r.trips[id]
	if !exist {
		return nil, fmt.Errorf("trip does not exist with ID: %s", id)
	}

	return trip, nil
}

func (r *inmemRepository) UpdateTrip(ctx context.Context, tripID, status string, driver *pbDriver.Driver) error {
	trip, exist := r.trips[tripID]
	if !exist {
		return fmt.Errorf("trip does not exist with ID: %s", tripID)
	}

	trip.Status = status
	trip.Driver = &pb.TripDriver{
		Id:             driver.Id,
		Name:           driver.Name,
		ProfilePicture: driver.ProfilePicture,
		CarPlate:       driver.CarPlate,
	}

	return nil
}
