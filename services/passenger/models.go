package passenger

type BookSeatRequest struct {
	RideId string `json:"rideId"`
	Code   string `json:"code"`
}
