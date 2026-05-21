package admin

type AdminLoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type AdminLoginResponse struct {
	Token string     `json:"sessionId"`
	Admin AdminLogin `json:"admin"`
}

type AdminLogin struct {
	ID string `json:"id"`
}

type DriverWithVehicle struct {
	ID            string    `json:"id"`
	DriverMobile  string    `json:"driverMobile"`
	DriverName    string    `json:"driverName"`
	Rating        string    `json:"rating"`
	NumberOfVotes string    `json:"numberOfVotes"`
	Status        string    `json:"status"`
	Vehicles      []Vehicle `json:"vehicles"`
}

type Vehicle struct {
	ID            string `json:"id"`
	DriverId      string `json:"driverId"`
	VehicleNumber string `json:"vehicleNumber"`
	VehicleInfo   string `json:"vehicleInfo"`
	Status        string `json:"status"`
}

type RideDetail struct {
	ID                   string  `json:"id"`
	DriverID             string  `json:"driver_id"`
	DriverName           string  `json:"driver_name"`
	DriverMobile         string  `json:"driver_mobile"`
	Rating               string  `json:"rating"`
	VehicleNumber        string  `json:"vehicle_number"`
	VehicleInfo          string  `json:"vehicle_info"`
	StartDatetime        string  `json:"start_datetime"`
	EstimatedEndDatetime string  `json:"estimated_end_datetime"`
	NumberOfSeats        int     `json:"number_of_seats"`
	SeatsTaken           int     `json:"seats_taken"`
	StartLocation        string  `json:"start_location"`
	EndLocation          string  `json:"end_location"`
	Fare                 float64 `json:"fare"`
	RouteDetails         string  `json:"route_details"`
	IsActive             bool    `json:"is_active"`
}

type DriverDetailsResponse struct {
	TotalPages int                 `json:"totalPages"`
	Details    []DriverWithVehicle `json:"details"`
}

type VehicleDetailsResponse struct {
	TotalPages int       `json:"totalPages"`
	Vehicles   []Vehicle `json:"vehicle"`
}

type RideDetailsResponse struct {
	TotalPages int          `json:"totalPages"`
	Rides      []RideDetail `json:"rides"`
}
