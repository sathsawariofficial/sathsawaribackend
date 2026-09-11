package admin

import (
	"time"
)

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

type AdminBroadcastRequest struct {
	UserType         int    `json:"userType"`
	Title            string `json:"title"`
	Message          string `json:"message"`
	NotificationType string `json:"notificationType"`
}

type ApprochInfo struct {
	Name      string    `json:"name"`
	Number    string    `json:"number"`
	Email     string    `json:"email"`
	Message   string    `json:"message"`
	Type      string    `json:"type"`
	CreatedAt time.Time `json:"created_at"`
}

type ApprochInfoResponse struct {
	TotalPages int           `json:"totalPages"`
	Approches  []ApprochInfo `json:"approches"`
}

type AnnouncementRequest struct {
	Title   string `json:"title"`
	Message string `json:"message"`
	Type    string `json:"type"`
	Link    string `json:"link"`
}

// AdminOverviewResponse is the platform at a glance, the first screen of an admin
// console before drilling into anything.
type AdminOverviewResponse struct {
	Drivers             AdminCountPair `json:"drivers"`
	Passengers          AdminCountPair `json:"passengers"`
	Vehicles            AdminCountPair `json:"vehicles"`
	Rides               AdminCountPair `json:"rides"`
	Services            AdminCountPair `json:"services"`
	Shifts              AdminCountPair `json:"shifts"`
	Advertisements      int64          `json:"advertisements"`
	OpenShiftRequests   int64          `json:"openShiftRequests"`
	PendingJoinRequests int64          `json:"pendingJoinRequests"`
}

// AdminCountPair is a total with the slice of it that is currently live, so the
// admin sees scale and activity in the same row.
type AdminCountPair struct {
	Total int64  `json:"total"`
	Live  int64  `json:"live"`
	Label string `json:"liveLabel"`
}

type AdminStatusRequest struct {
	Status string `json:"status"`
}

type AdminPassengerDetail struct {
	ID              string    `json:"id"`
	PassengerName   string    `json:"passengerName"`
	PassengerMobile string    `json:"passengerMobile"`
	Gender          string    `json:"gender"`
	Status          string    `json:"status"`
	CreatedAt       time.Time `json:"createdAt"`
}

type AdminPassengerListResponse struct {
	TotalPages int                    `json:"totalPages"`
	Passengers []AdminPassengerDetail `json:"passengers"`
}

type AdminPassengerProfileResponse struct {
	Passenger AdminPassengerDetail `json:"passenger"`
}

////////////////////////////// PICK & DROP OVERSIGHT //////////////////////////////
// Read only: Pick & Drop itself stays run by each service's owner, the admin console
// only ever looks at it.

// adminShiftFilter narrows a platform-wide shift or shift-request list the same way
// the owner's own search does: a day of the week, a place or name, a clock window.
type adminShiftFilter struct {
	Status    string
	DayOfWeek int
	Search    string
	StartTime string
	EndTime   string
}

type AdminServiceSummary struct {
	ID             string    `json:"id"`
	Name           string    `json:"name"`
	Description    string    `json:"description"`
	OwnerDriverId  string    `json:"ownerDriverId"`
	OwnerName      string    `json:"ownerName"`
	OwnerMobile    string    `json:"ownerMobile"`
	Status         string    `json:"status"`
	DriverCount    int64     `json:"driverCount"`
	VehicleCount   int64     `json:"vehicleCount"`
	PassengerCount int64     `json:"passengerCount"`
	ShiftCount     int64     `json:"shiftCount"`
	CreatedAt      time.Time `json:"createdAt"`
}

type AdminServicesResponse struct {
	TotalPages int                   `json:"totalPages"`
	Services   []AdminServiceSummary `json:"services"`
}

// AdminServiceCounts is the same roster breakdown the owner's own dashboard sees:
// approved members by kind, vehicles-only members counted apart from driving ones,
// and everything still waiting on a decision.
type AdminServiceCounts struct {
	ApprovedDrivers       int64 `json:"approvedDrivers"`
	ApprovedVehicleOwners int64 `json:"approvedVehicleOwners"`
	ApprovedVehicles      int64 `json:"approvedVehicles"`
	ApprovedPassengers    int64 `json:"approvedPassengers"`
	PendingRequests       int64 `json:"pendingRequests"`
	ActiveShifts          int64 `json:"activeShifts"`
}

type AdminServiceDetailResponse struct {
	Service AdminServiceSummary `json:"service"`
	Counts  AdminServiceCounts  `json:"counts"`
	Shifts  []AdminShiftSummary `json:"shifts"`
}

type AdminShiftSummary struct {
	ID               string    `json:"id"`
	Name             string    `json:"name"`
	ServiceId        string    `json:"serviceId"`
	ServiceName      string    `json:"serviceName"`
	ServiceOwnerId   string    `json:"serviceOwnerId"`
	DriverId         string    `json:"driverId"`
	DriverName       string    `json:"driverName"`
	DriverMobile     string    `json:"driverMobile"`
	VehicleId        string    `json:"vehicleId"`
	VehicleNumber    string    `json:"vehicleNumber"`
	VehicleOwnerId   string    `json:"vehicleOwnerId"`
	VehicleOwnerName string    `json:"vehicleOwnerName"`
	DaysOfWeek       []int     `json:"daysOfWeek"`
	StartDate        string    `json:"startDate"`
	EndDate          string    `json:"endDate"`
	StartTime        string    `json:"startTime"`
	EndTime          string    `json:"endTime"`
	SeatCapacity     int       `json:"seatCapacity"`
	OccupiedSeats    int       `json:"occupiedSeats"`
	Status           string    `json:"status"`
	CreatedAt        time.Time `json:"createdAt"`
}

type AdminShiftsResponse struct {
	TotalPages int                 `json:"totalPages"`
	Shifts     []AdminShiftSummary `json:"shifts"`
}

type AdminShiftRequestItem struct {
	ID            string    `json:"id"`
	PassengerId   string    `json:"passengerId"`
	PassengerName string    `json:"passengerName"`
	ServiceId     string    `json:"serviceId"`
	ServiceName   string    `json:"serviceName"`
	ContactNumber string    `json:"contactNumber"`
	Note          string    `json:"note"`
	DaysOfWeek    []int     `json:"daysOfWeek"`
	StartTime     string    `json:"startTime"`
	EndTime       string    `json:"endTime"`
	CreatedAt     time.Time `json:"createdAt"`
}

type AdminShiftRequestsResponse struct {
	TotalPages int                     `json:"totalPages"`
	Requests   []AdminShiftRequestItem `json:"requests"`
}

type AdminAdvertisementItem struct {
	ID            string    `json:"id"`
	ServiceId     string    `json:"serviceId"`
	ServiceName   string    `json:"serviceName"`
	OwnerDriverId string    `json:"ownerDriverId"`
	OwnerName     string    `json:"ownerName"`
	OwnerMobile   string    `json:"ownerMobile"`
	Title         string    `json:"title"`
	Description   string    `json:"description"`
	Fare          float64   `json:"fare"`
	DaysOfWeek    []int     `json:"daysOfWeek"`
	StartTime     string    `json:"startTime"`
	EndTime       string    `json:"endTime"`
	VehicleCount  int64     `json:"vehicleCount"`
	CreatedAt     time.Time `json:"createdAt"`
}

type AdminAdvertisementsResponse struct {
	TotalPages     int                      `json:"totalPages"`
	Advertisements []AdminAdvertisementItem `json:"advertisements"`
}
