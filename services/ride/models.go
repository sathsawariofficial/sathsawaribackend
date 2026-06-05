package ride

type RideCreationRequest struct {
	VehicleId            string  `json:"vehicleId"`
	StartDatetime        string  `json:"startDatetime"`
	EstimatedEndDatetime string  `json:"estimatedEndDatetime"`
	NumberOfSeats        int     `json:"numberOfSeats"`
	StartLocation        string  `json:"startLocation"`
	EndLocation          string  `json:"endLocation"`
	Fare                 float64 `json:"fare"`
	RouteDetails         string  `json:"routeDetails"`
	IsRecurring          bool    `json:"isRecurring"`
	MakeTemplate         bool    `json:"makeTemplate"`
	Frequency            int     `json:"frequency"`  // for how long will this happen
	Period               int     `json:"period"`     // whats the period like daily, weekly, monthly
	DaysOfWeek           []int   `json:"daysOfWeek"` // if monthly or weekly than for how many days of the week
	EXTDriverId          string  `json:"-"`
}

type RideResponse struct {
	RideId string `json:"rideId"`
}

type RidesDetailsResponse struct {
	TotalPages int           `json:"totalPages"`
	Rides      []RideDetails `json:"rides"`
}

type DriverRideResponse struct {
	DriverId   string        `json:"driverId"`
	Message    string        `json:"message"`
	TotalPages int           `json:"totalPages"`
	Rides      []RideDetails `json:"rides"`
}

type RideDetails struct {
	ID                   string  `json:"id"`
	DriverID             string  `json:"driverId"`
	DriverName           string  `json:"driverName"`
	DriverMobile         string  `json:"driverMobile"`
	Rating               string  `json:"rating"`
	VehicleNumber        string  `json:"vehicleNumber"`
	VehicleInfo          string  `json:"vehicleInfo"`
	StartDatetime        string  `json:"startDatetime"`
	EstimatedEndDatetime string  `json:"estimatedEndDatetime"`
	NumberOfSeats        int     `json:"numberOfSeats"`
	SeatsTaken           int     `json:"seatsTaken"`
	StartLocation        string  `json:"startLocation"`
	EndLocation          string  `json:"endLocation"`
	Fare                 float64 `json:"fare"`
	RouteDetails         string  `json:"routeDetails"`
	IsActive             bool    `json:"isActive"`
}

type UpdateRideRequest struct {
	Status        *string `json:"status"`
	NumberOfSeats int     `json:"numberOfSeats"`
}

type BookSeatRequest struct {
	RideId        string `json:"rideId"`
	PassengerId   string `json:"passengerId"`
	Name          string `json:"name"`
	MobileNumber  string `json:"mobileNumber"`
	NumberOfSeats int    `json:"numberOfSeats"`
}

type BookSeats struct {
	RideId       string `json:"rideId"`
	PassengerId  string `json:"passengerId,omitempty"`
	Name         string `json:"name,omitempty"`
	MobileNumber string `json:"mobileNumber,omitempty"`
}

type RideTemplate struct {
	ID                   string  `json:"id"`
	RideID               string  `json:"rideId"`
	DriverID             string  `json:"driverId"`
	VehicleID            string  `json:"vehicleId"`
	StartDatetime        string  `json:"startDatetime"`
	EstimatedEndDatetime string  `json:"estimatedEndDatetime"`
	NumberOfSeats        int     `json:"numberOfSeats"`
	SeatsTaken           int     `json:"seatsTaken"`
	StartLocation        string  `json:"startLocation"`
	EndLocation          string  `json:"endLocation"`
	Fare                 float64 `json:"fare"`
	RouteDetails         string  `json:"routeDetails"`
}
