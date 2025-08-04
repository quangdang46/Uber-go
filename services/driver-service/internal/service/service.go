package service

import (
	math "math/rand/v2"
	"ride-sharing/services/driver-service/pkg/utils"
	pb "ride-sharing/shared/proto/driver"
	"sync"

	"github.com/mmcloughlin/geohash"
)

type driverInMap struct {
	Driver *pb.Driver
}

type Service struct {
	drivers []*driverInMap
	mu      sync.RWMutex
}

func NewService() *Service {
	return &Service{
		drivers: make([]*driverInMap, 0),
	}
}

func (s *Service) RegisterDriver(driverId string, packageSlug string) (*pb.Driver, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	randomIndex := math.IntN(len(utils.PredefinedRoutes))
	randomRoute := utils.PredefinedRoutes[randomIndex]

	// we can ignore this property for now, but it must be sent to the frontend.
	geohash := geohash.Encode(randomRoute[0][0], randomRoute[0][1])

	driver := &pb.Driver{
		Geohash:  geohash,
		Location: &pb.Location{Latitude: randomRoute[0][0], Longitude: randomRoute[0][1]},
		Name:     "Lando Norris",
		// Id: ...,
		// PackageSlug:    packageSlug,
		// ProfilePicture: randomAvatar,
		// CarPlate:       randomPlate,
	}

	// TODO: Add driver to list

	return driver, nil
}

func (s *Service) UnregisterDriver(driverId string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	// TODO: Filter driver from list
}
