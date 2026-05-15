package passenger

type BookSeatDemandRequest struct {
	RideId        string `json:"rideId"`
	NumberOfSeats int    `json:"numberOfSeats"`
	Name          string `json:"name"`
	Mobilenumber  string `json:"mobileNumber"`
	Message       string `json:"message"`
}
