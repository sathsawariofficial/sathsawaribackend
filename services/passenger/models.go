package passenger

type BookSeatRequest struct {
	RideId       string `json:"rideId"`
	Name         string `json:"name"`
	MobileNumber string `json:"mobileNumber"`
	Seats        int    `json:"seats"`
	Code         string `json:"code"`
}
