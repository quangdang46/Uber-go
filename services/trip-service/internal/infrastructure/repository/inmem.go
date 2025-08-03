package repository

import (
	"context"
	"ride-sharing/services/trip-service/internal/domain"
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
