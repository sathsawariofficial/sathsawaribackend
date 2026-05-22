package postgress

import "time"

type Admin struct {
	ID        string    `json:"id" gorm:"primary_key"`
	Username  string    `json:"user_name" gorm:"not null"`
	Password  string    `json:"password" gorm:"not null"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type SMSFCM struct {
	ID        string    `json:"id" gorm:"primary_key"`
	App       string    `json:"app"`
	FCM       string    `json:"fcm" gorm:"not null"`
	APPHash   string    `json:"app_hash"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type Driver struct {
	ID            string    `json:"id" gorm:"primary_key"`
	DriverMobile  string    `json:"driver_mobile" gorm:"unique;not null"`
	DriverName    string    `json:"driver_name" gorm:"not null"`
	Password      string    `json:"password" gorm:"not null"`
	Pin           string    `json:"pin" gorm:"not null"`
	Rating        string    `json:"rating"`
	NumberOfVotes string    `json:"number_of_votes"`
	Status        string    `json:"status"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`

	// forign key relation
	Vehicles []Vehicle `json:"vehicles" gorm:"foreignKey:DriverId;references:ID"`
}

type DriverDevice struct {
	DriverId string `json:"driver_id"`
	DeviceId string `json:"device_id"`
}

type DriverFCM struct {
	ID        string    `json:"id" gorm:"primary_key"`
	DriverId  string    `json:"driver_Id" gorm:"not null"`
	FCM       string    `json:"FCM"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type Vehicle struct {
	ID            string    `json:"id" gorm:"primary_key"`
	DriverId      string    `json:"driver_id" gorm:"not null"`
	VehicleNumber string    `json:"vehicle_number" gorm:"unique;not null"`
	VehicleInfo   string    `json:"vehicle_info" gorm:"not null"`
	Status        string    `json:"status"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

type Ride struct {
	ID                   string    `json:"id" gorm:"primary_key"`
	DriverID             string    `json:"driver_id" gorm:"not null"`
	VehicleID            string    `json:"vehicle_id" gorm:"not null"`
	StartDatetime        string    `json:"start_datetime" gorm:"not null"`
	EstimatedEndDatetime string    `json:"estimated_end_datetime" gorm:"not null"`
	NumberOfSeats        int       `json:"number_of_seats" gorm:"not null"`
	SeatsTaken           int       `json:"seats_taken" gorm:"not null;default:0"`
	StartLocation        string    `json:"start_location" gorm:"not null"`
	EndLocation          string    `json:"end_location" gorm:"not null"`
	Fare                 float64   `json:"fare" gorm:"not null"`
	RouteDetails         string    `json:"route_details" gorm:"not null"`
	IsActive             bool      `json:"is_active" gorm:"not null"`
	ParentRideId         string    `json:"parent_id"`
	CreatedAt            time.Time `json:"created_at"`
	UpdatedAt            time.Time `json:"updated_at"`
}

type RideTemplate struct {
	ID                   string    `json:"id" gorm:"primary_key"`
	RideID               string    `json:"ride_id" gorm:"not null"`
	DriverID             string    `json:"driver_id" gorm:"not null"`
	VehicleID            string    `json:"vehicle_id" gorm:"not null"`
	StartDatetime        string    `json:"start_datetime" gorm:"not null"`
	EstimatedEndDatetime string    `json:"estimated_end_datetime" gorm:"not null"`
	NumberOfSeats        int       `json:"number_of_seats" gorm:"not null"`
	SeatsTaken           int       `json:"seats_taken" gorm:"not null;default:0"`
	StartLocation        string    `json:"start_location" gorm:"not null"`
	EndLocation          string    `json:"end_location" gorm:"not null"`
	Fare                 float64   `json:"fare" gorm:"not null"`
	RouteDetails         string    `json:"route_details" gorm:"not null"`
	IsActive             bool      `json:"is_active"`
	CreatedAt            time.Time `json:"created_at"`
	UpdatedAt            time.Time `json:"updated_at"`
}

type RideDetails struct {
	ID                   string    `json:"id"`
	DriverID             string    `json:"driver_id"`
	DriverName           string    `json:"driver_name"`
	DriverMobile         string    `json:"driver_mobile"`
	Rating               string    `json:"rating"`
	VehicleNumber        string    `json:"vehicle_number"`
	VehicleInfo          string    `json:"vehicle_info"`
	StartDatetime        string    `json:"start_datetime"`
	EstimatedEndDatetime string    `json:"estimated_end_datetime"`
	NumberOfSeats        int       `json:"number_of_seats"`
	SeatsTaken           int       `json:"seats_taken"`
	StartLocation        string    `json:"start_location"`
	EndLocation          string    `json:"end_location"`
	Fare                 float64   `json:"fare"`
	RouteDetails         string    `json:"route_details"`
	IsActive             bool      `json:"is_active"`
	CreatedAt            time.Time `json:"created_at"`
	UpdatedAt            time.Time `json:"updated_at"`
}

type RidePassenger struct {
	ID           string    `json:"id" gorm:"primary_key"`
	RideId       string    `json:"ride_id"`
	PassengerId  string    `json:"passenger_id"`
	Name         string    `json:"name" gorm:"not null"`
	MobileNumber string    `json:"mobile_number"`
	IsActive     bool      `json:"is_active"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

type PassengerRating struct {
	ID                    string    `json:"id" gorm:"primary_key"`
	DriverID              string    `json:"driver_id"`
	RideID                string    `json:"ride_id"`
	PassengerMobileNumber string    `json:"passenger_mobile_number" gorm:"not null"`
	CreatedAt             time.Time `json:"created_at"`
}

type DriverDetails struct {
	DriverName    string `json:"driver_name"`
	DriverMobile  string `json:"driver_mobile"`
	VehicleNumber string `json:"vehicle_number"`
	VehicleInfo   string `json:"vehicle_info"`
	TotalRides    int    `json:"total_rides"`
	Rating        string `json:"rating"`
}

type DriverOldData struct {
	ID            string    `json:"id" gorm:"primary_key"`
	DriverID      string    `json:"driver_id"`
	DriverMobile  string    `json:"driver_mobile" gorm:"unique;not null"`
	DriverName    string    `json:"driver_name" gorm:"not null"`
	Password      string    `json:"password" gorm:"not null"`
	VehicleNumber string    `json:"vehicle_number" gorm:"unique;not null"`
	VehicleInfo   string    `json:"vehicle_info" gorm:"not null"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

type ApprochInfo struct {
	ID        string    `json:"id" gorm:"primary_key"`
	Name      string    `json:"name"`
	Number    string    `json:"number"`
	Email     string    `json:"email"`
	Message   string    `json:"message"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type BookSeatDemand struct {
	ID            string    `json:"id" gorm:"primary_key"`
	RideId        string    `json:"ride_id" gorm:"not null"`
	Name          string    `json:"name" gorm:"not null"`
	Mobilenumber  string    `json:"mobile_number"`
	NumberOfSeats int       `json:"number_of_seats"`
	Message       string    `json:"message"`
	IsHandled     bool      `json:"is_handled"`
	IsApproved    bool      `json:"is_approved"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

type NotificationRequest struct {
	ID               string    `json:"id" gorm:"primary_key"`
	UserId           string    `json:"user_id" gorm:"not null"`
	UserType         int       `json:"user_type" gorm:"not null"`
	Title            string    `json:"title"`
	Message          string    `json:"message"`
	NotificationType string    `json:"notification_request"`
	Data             string    `json:"data"`
	CreatedAt        time.Time `json:"created_at"`
	UpdatedAt        time.Time `json:"updated_at"`
}

type MissingLocations struct {
	DeviceId string  `json:"device_id"`
	Place    string  `json:"place"`
	UserLat  float64 `json:"lat"`
	UserLng  float64 `json:"long"`
}
