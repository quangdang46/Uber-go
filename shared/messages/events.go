package messages

import pb "ride-sharing/shared/proto/trip"

const (
	FindAvailableDriversQueue = "find_available_drivers"
	NotifyNewTripQueue        = "notify_new_trip"
)

type TripEventData struct {
	Trip *pb.Trip `json:"trip"`
}
