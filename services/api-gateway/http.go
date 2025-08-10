package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"ride-sharing/services/api-gateway/grpc_clients"
	"ride-sharing/shared/contracts"
	"ride-sharing/shared/env"
	"ride-sharing/shared/messages"
	"ride-sharing/shared/tracing"
	"github.com/stripe/stripe-go/v81"
	"github.com/stripe/stripe-go/v81/webhook"
)

var tracer = tracing.GetTracer("api-gateway")

func handleTripPreview(w http.ResponseWriter, r *http.Request) {

	ctx,span:=tracer.Start(r.Context(),"handleTripPreview")
	defer span.End()

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

	tripService, err := grpc_clients.NewTripServiceClient()

	if err != nil {
		log.Printf("Failed to create trip service client: %v", err)
		http.Error(w, "Failed to create trip service client", http.StatusInternalServerError)
		return
	}

	defer tripService.Close()

	tripPreview, err := tripService.Client.PreviewTrip(ctx, reqBody.toProto())

	if err != nil {
		log.Printf("Failed to preview a trip: %v", err)
		http.Error(w, "Failed to preview trip", http.StatusInternalServerError)
		return
	}

	response := contracts.APIResponse{Data: tripPreview}

	WriteJSON(w, http.StatusCreated, response)
}

func handleTripStart(w http.ResponseWriter, r *http.Request) {

	ctx,span:=tracer.Start(r.Context(),"handleTripStart")
	defer span.End()


	var startTrip startTripRequest

	if err := json.NewDecoder(r.Body).Decode(&startTrip); err != nil {
		http.Error(w, "failed to parse JSON data", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	tripService, err := grpc_clients.NewTripServiceClient()

	if err != nil {
		log.Printf("Failed to create trip service client: %v", err)
		http.Error(w, "Failed to create trip service client", http.StatusInternalServerError)
		return
	}

	defer tripService.Close()
	fmt.Println("start trip", startTrip)

	trip, err := tripService.Client.CreateTrip(ctx, startTrip.toProto())

	if err != nil {
		http.Error(w, "Failed to start trip", http.StatusInternalServerError)
		return
	}
  
	response := contracts.APIResponse{Data: trip}

	WriteJSON(w, http.StatusCreated, response)
}

func handleStripeWebhook(w http.ResponseWriter, r *http.Request, rabbitmq *messages.RabbitMQ) {
	
	ctx,span:=tracer.Start(r.Context(),"handleStripeWebhook")
	defer span.End()
	body, err := io.ReadAll(r.Body)

	if err != nil {
		http.Error(w, "failed to read request body", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	webhookKey := env.GetString("STRIPE_WEBHOOK_SECRET", "")

	if webhookKey == "" {
		http.Error(w, "stripe webhook secret is not set", http.StatusInternalServerError)
		return
	}

	event, err := webhook.ConstructEventWithOptions(body, r.Header.Get("Stripe-Signature"), webhookKey, webhook.ConstructEventOptions{
		IgnoreAPIVersionMismatch: true,
	})

	if err != nil {
		http.Error(w, "failed to construct stripe event", http.StatusInternalServerError)
		return
	}

	log.Printf("Stripe event: %+v", event)

	switch event.Type {
	case "checkout.session.completed":
		var session stripe.CheckoutSession

		err := json.Unmarshal(event.Data.Raw, &session)
		if err != nil {
			http.Error(w, "failed to unmarshal stripe event", http.StatusInternalServerError)
			return
		}

		payload := messages.PaymentStatusUpdateData{
			TripID:   session.Metadata["trip_id"],
			UserID:   session.Metadata["user_id"],
			DriverID: session.Metadata["driver_id"],
		}

		jsonPayload, err := json.Marshal(payload)
		if err != nil {
			http.Error(w, "failed to marshal stripe event", http.StatusInternalServerError)
			return
		}

		message := contracts.AmqpMessage{
			OwnerID: session.Metadata["user_id"],
			Data:    jsonPayload,
		}

		if err := rabbitmq.PublishMessage(ctx, contracts.PaymentEventSuccess, message); err != nil {
			log.Printf("failed to publish message: %v", err)
			http.Error(w, "failed to publish message", http.StatusInternalServerError)
			return
		}

	default:
		log.Printf("Unhandled stripe event: %s", event.Type)
	}
}
 