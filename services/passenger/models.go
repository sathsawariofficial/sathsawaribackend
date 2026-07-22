package passenger

type BookSeatRequest struct {
	RideId       string `json:"rideId"`
	Name         string `json:"name"`
	MobileNumber string `json:"mobileNumber"`
	Seats        int    `json:"seats"`
	Code         string `json:"code"`
}

type RideRequest struct {
	StartDatetime        string `json:"startDatetime"`
	EstimatedEndDatetime string `json:"estimatedEndDatetime"`
	NumberOfSeats        int    `json:"numberOfSeats"`
	StartLocation        string `json:"startLocation"`
	EndLocation          string `json:"endLocation"`
	RouteDetails         string `json:"routeDetails"`
	ContactNumber        string `json:"contactNumber"`
}

type GetRideRequest struct {
	StartDatetime        string `json:"startDatetime"`
	EstimatedEndDatetime string `json:"estimatedEndDatetime"`
	StartLocation        string `json:"startLocation"`
	EndLocation          string `json:"endLocation"`
}

type RideRequestDetails struct {
	ID                   string `json:"id"`
	ContactNumber        string `json:"contactNumber"`
	StartDatetime        string `json:"startDatetime"`
	EstimatedEndDatetime string `json:"estimatedEndDatetime"`
	NumberOfSeats        int    `json:"numberOfSeats"`
	StartLocation        string `json:"startLocation"`
	EndLocation          string `json:"endLocation"`
	RouteDetails         string `json:"routeDetails"`
}

type RidesRequestDetailsResponse struct {
	TotalPages int                  `json:"totalPages"`
	Rides      []RideRequestDetails `json:"rides"`
}
