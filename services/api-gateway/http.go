package main

import (
	"encoding/json"
	"net/http"
	"ride-sharing/shared/contracts"
)

func HandleTripPreview(w http.ResponseWriter, r *http.Request) {

	var reqBody previewTripRequest

	if err := json.NewDecoder(r.Body).Decode(&reqBody); err != nil {
		http.Error(w, "failed to parse JSON data", http.StatusBadRequest)
		return
	}

	defer r.Body.Close()

	// validation

	if reqBody.UserID == "" {
		http.Error(w, "user id is required", http.StatusBadRequest)
		return
	}

	// TODO: call grpc to trip service

	response := contracts.APIResponse{Data: "ok"}

	WriteJSON(w, http.StatusCreated, response)
}
