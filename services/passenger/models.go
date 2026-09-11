package passenger

type BookSeatRequest struct {
	RideId       string `json:"rideId"`
	Name         string `json:"name"`
	MobileNumber string `json:"mobileNumber"`
	Seats        int    `json:"seats"`
	Code         string `json:"code"`
}

type BookSeatResponse struct {
	BookingId string `json:"bookingId"`
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

type RideRequestResponse struct {
	OpenURL   string `json:"openUrl"`
	RequestId string `json:"requestId"`
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
	OpenURL              string `json:"openUrl"`
}

type RidesRequestsDetailsResponse struct {
	TotalPages int                  `json:"totalPages"`
	Rides      []RideRequestDetails `json:"rides"`
}

// NotificationSettingsRequest is the whole place notification setting of a passenger's
// device: while enabled, rides taking in one of the places are pushed to the fcm. What
// is sent replaces what was saved before.
type NotificationSettingsRequest struct {
	DeviceId string   `json:"deviceId"`
	FCM      string   `json:"fcm"`
	Enabled  bool     `json:"enabled"`
	Places   []string `json:"places"`
}

type NotificationSettingsResponse struct {
	DeviceId string   `json:"deviceId"`
	Enabled  bool     `json:"enabled"`
	Places   []string `json:"places"`
}
