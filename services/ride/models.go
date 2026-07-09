package ride

type RideCreationRequest struct {
	VehicleId            string   `json:"vehicleId"`
	StartDatetime        string   `json:"startDatetime"`
	EstimatedEndDatetime string   `json:"estimatedEndDatetime"`
	NumberOfSeats        int      `json:"numberOfSeats"`
	StartLocation        string   `json:"startLocation"`
	EndLocation          string   `json:"endLocation"`
	RoutePoints          []string `json:"routePoints"`
	Fare                 float64  `json:"fare"`
	RouteDetails         string   `json:"routeDetails"`
	IsRecurring          bool     `json:"isRecurring"`
	MakeTemplate         bool     `json:"makeTemplate"`
	Frequency            int      `json:"frequency"`  // for how long will this happen
	Period               int      `json:"period"`     // whats the period like daily, weekly, monthly
	DaysOfWeek           []int    `json:"daysOfWeek"` // if monthly or weekly than for how many days of the week
	EXTDriverId          string   `json:"-"`
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

type RideDetailsWithChildrenResponse struct {
	Ride       RideDetails   `json:"ride"`
	ChildRides []RideDetails `json:"childRides"`
}

type RideDetails struct {
	ID                   string   `json:"id,omitempty"`
	DriverID             string   `json:"driverId,omitempty"`
	DriverName           string   `json:"driverName,omitempty"`
	DriverMobile         string   `json:"driverMobile,omitempty"`
	Rating               string   `json:"rating,omitempty"`
	VehicleId            string   `json:"vehicleId,omitempty"`
	VehicleNumber        string   `json:"vehicleNumber,omitempty"`
	VehicleInfo          string   `json:"vehicleInfo,omitempty"`
	StartDatetime        string   `json:"startDatetime,omitempty"`
	EstimatedEndDatetime string   `json:"estimatedEndDatetime,omitempty"`
	NumberOfSeats        int      `json:"numberOfSeats,omitempty"`
	SeatsTaken           int      `json:"seatsTaken,omitempty"`
	StartLocation        string   `json:"startLocation,omitempty"`
	EndLocation          string   `json:"endLocation,omitempty"`
	RoutePoints          []string `json:"routePoints,omitempty"`
	Code                 string   `json:"code,omitempty"`
	Fare                 float64  `json:"fare,omitempty"`
	RouteDetails         string   `json:"routeDetails,omitempty"`
	ParentRideId         string   `json:"parent_id,omitempty"`
	IsActive             bool     `json:"isActive,omitempty"`
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
}
