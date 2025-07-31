package main

import (
	"log"
	"net/http"

	"ride-sharing/shared/env"
)

var (
	httpAddr = env.GetString("HTTP_ADDR", ":8081")
)

func main() {
	log.Println("Starting API Gateway")

	mux := http.NewServeMux()

	mux.HandleFunc("POST /trip/preview", HandleTripPreview)

	server := &http.Server{
		Handler: mux,
		Addr:    httpAddr,
	}
	if err := server.ListenAndServe(); err != nil {
		log.Printf("HTTP server error:%v", err)
	}
}
