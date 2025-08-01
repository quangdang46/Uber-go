package main

import (
	"log"
	"net/http"
	httpHandler "ride-sharing/services/trip-service/internal/infrastructure/http"
	"ride-sharing/services/trip-service/internal/infrastructure/repository"
	"ride-sharing/services/trip-service/internal/service"
)

func main() {

	immemRepo := repository.NewInMemRepository()
	svc := service.NewService(immemRepo)

	mux := http.NewServeMux()

	httphandler := &httpHandler.HttpHandler{Service: svc}

	mux.HandleFunc("POST /preview", httphandler.HandleTripPreview)

	server := &http.Server{
		Handler: mux,
		Addr:    ":8083",
	}
	log.Println("Trip service is running in port 8083")

	if err := server.ListenAndServe(); err != nil {
		log.Printf("HTTP server error:%v", err)
	}
}
