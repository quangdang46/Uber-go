package main

import (
	"context"
	"log"
	"ride-sharing/services/trip-service/internal/domain"
	"ride-sharing/services/trip-service/internal/infrastructure/repository"
	"ride-sharing/services/trip-service/internal/service"
	"time"
)

func main() {

	ctx := context.Background()

	immemRepo := repository.NewInMemRepository()

	svc := service.NewService(immemRepo)

	fare := &domain.RideFareModel{
		UserID: "11",
	}

	t, err := svc.CreateTrip(ctx, fare)

	if err != nil {
		log.Println(err)
	}

	log.Println(t)

	for {
		time.Sleep(time.Second)
 	}

}
