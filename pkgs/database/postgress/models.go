package postgress

import (
	"time"

	"github.com/lib/pq"
)

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
	UpdateBy      string    `json:"update_by"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`

	// forign key relation
	Vehicles []Vehicle `json:"vehicles" gorm:"foreignKey:DriverId;references:ID"`
}

type DELDriver struct {
	ID            string    `json:"id" gorm:"primary_key"`
	DriverMobile  string    `json:"driver_mobile" gorm:"unique;not null"`
	DriverName    string    `json:"driver_name" gorm:"not null"`
	Password      string    `json:"password" gorm:"not null"`
	Pin           string    `json:"pin" gorm:"not null"`
	Rating        string    `json:"rating"`
	NumberOfVotes string    `json:"number_of_votes"`
	Status        string    `json:"status"`
	UpdateBy      string    `json:"update_by"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

type DriverDevice struct {
	DriverId string `json:"driver_id"`
	DeviceId string `json:"device_id"`
}

type UserFCM struct {
	ID        string    `json:"id" gorm:"primary_key"`
	UserId    string    `gorm:"index:idx_user_fcm,unique;not null"`
	FCM       string    `json:"FCM"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// Vehicle carries the number of passenger seats it offers, the driver's own seat is
// not counted. A vehicle registered before seats were recorded holds 0, which means
// unknown: ride share keeps working on it, but it can never be put on a shift.
type Vehicle struct {
	ID            string    `json:"id" gorm:"primary_key"`
	DriverId      string    `json:"driver_id" gorm:"not null"`
	VehicleNumber string    `json:"vehicle_number" gorm:"unique;not null"`
	VehicleInfo   string    `json:"vehicle_info" gorm:"not null"`
	Status        string    `json:"status"`
	NumberOfSeats int       `json:"number_of_seats" gorm:"not null;default:0"`
	HasAC         bool      `json:"has_ac" gorm:"not null;default:false"`
	HasHeating    bool      `json:"has_heating" gorm:"not null;default:false"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

type DELVehicle struct {
	ID            string    `json:"id" gorm:"primary_key"`
	DriverId      string    `json:"driver_id" gorm:"not null"`
	VehicleNumber string    `json:"vehicle_number" gorm:"unique;not null"`
	VehicleInfo   string    `json:"vehicle_info" gorm:"not null"`
	Status        string    `json:"status"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

type Ride struct {
	ID                   string         `json:"id" gorm:"primary_key"`
	DriverID             string         `json:"driver_id" gorm:"not null"`
	VehicleID            string         `json:"vehicle_id" gorm:"not null"`
	StartDatetime        string         `json:"start_datetime" gorm:"not null"`
	EstimatedEndDatetime string         `json:"estimated_end_datetime" gorm:"not null"`
	NumberOfSeats        int            `json:"number_of_seats" gorm:"not null"`
	SeatsTaken           int            `json:"seats_taken" gorm:"not null;default:0"`
	StartLocation        string         `json:"start_location" gorm:"not null"`
	EndLocation          string         `json:"end_location" gorm:"not null"`
	RoutePoints          pq.StringArray `json:"route_points" gorm:"type:text[]"`
	Fare                 float64        `json:"fare" gorm:"not null"`
	RouteDetails         string         `json:"route_details" gorm:"not null"`
	IsActive             bool           `json:"is_active" gorm:"not null"`
	ParentRideId         string         `json:"parent_id"`
	Code                 string         `json:"code"`
	CreatedAt            time.Time      `json:"created_at"`
	UpdatedAt            time.Time      `json:"updated_at"`
}

type RideSearchReplica struct {
	RideID         string         `gorm:"column:ride_id"`
	StartLocation  string         `gorm:"column:start_location"`
	EndLocation    string         `gorm:"column:end_location"`
	RoutePoints    pq.StringArray `gorm:"column:route_points;type:text[]"`
	StartDatetime  string         `gorm:"column:start_datetime"`
	AvailableSeats int            `gorm:"column:available_seats"`
	IsActive       bool           `gorm:"column:is_active"`
}

// TableName tells GORM to point to your new search replica table
func (RideSearchReplica) TableName() string {
	return "ride_searches"
}

type DELRide struct {
	ID                   string         `json:"id" gorm:"primary_key"`
	DriverID             string         `json:"driver_id" gorm:"not null"`
	VehicleID            string         `json:"vehicle_id" gorm:"not null"`
	StartDatetime        string         `json:"start_datetime" gorm:"not null"`
	EstimatedEndDatetime string         `json:"estimated_end_datetime" gorm:"not null"`
	NumberOfSeats        int            `json:"number_of_seats" gorm:"not null"`
	SeatsTaken           int            `json:"seats_taken" gorm:"not null;default:0"`
	StartLocation        string         `json:"start_location" gorm:"not null"`
	EndLocation          string         `json:"end_location" gorm:"not null"`
	RoutePoints          pq.StringArray `json:"route_points" gorm:"type:text[]"`
	Fare                 float64        `json:"fare" gorm:"not null"`
	RouteDetails         string         `json:"route_details" gorm:"not null"`
	IsActive             bool           `json:"is_active" gorm:"not null"`
	ParentRideId         string         `json:"parent_id"`
	Code                 string         `json:"code"`
	CreatedAt            time.Time      `json:"created_at"`
	UpdatedAt            time.Time      `json:"updated_at"`
}

type RideBooking struct {
	ID           string    `json:"id" gorm:"primary_key"`
	RideID       string    `json:"ride_id" gorm:"not null"`
	BookingID    string    `json:"booking_id"`
	Name         string    `json:"name" gorm:"not null"`
	MobileNumber string    `json:"mobileNumber" gorm:"not null"`
	Seats        int       `json:"seats" gorm:"not null"`
	Reserved     bool      `json:"reserved" gorm:"not null;default:false"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
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
	CreatedAt            time.Time `json:"created_at"`
	UpdatedAt            time.Time `json:"updated_at"`

	// forign key relation
	Vehicle Vehicle `json:"vehicles" gorm:"foreignKey:VehicleID;references:ID"`
}

type RideDetails struct {
	ID                   string         `json:"id"`
	DriverID             string         `json:"driver_id"`
	DriverName           string         `json:"driver_name"`
	DriverMobile         string         `json:"driver_mobile"`
	Rating               string         `json:"rating"`
	VehicleNumber        string         `json:"vehicle_number"`
	VehicleInfo          string         `json:"vehicle_info"`
	StartDatetime        string         `json:"start_datetime"`
	EstimatedEndDatetime string         `json:"estimated_end_datetime"`
	NumberOfSeats        int            `json:"number_of_seats"`
	SeatsTaken           int            `json:"seats_taken"`
	StartLocation        string         `json:"start_location"`
	EndLocation          string         `json:"end_location"`
	RoutePoints          pq.StringArray `json:"route_points" gorm:"column:route_points"`
	Fare                 float64        `json:"fare"`
	VehicleId            string         `json:"vehicle_id"`
	Code                 string         `json:"code"`
	RouteDetails         string         `json:"route_details"`
	ParentRideId         string         `json:"parent_id"`
	IsActive             bool           `json:"is_active"`
	CreatedAt            time.Time      `json:"created_at"`
	UpdatedAt            time.Time      `json:"updated_at"`
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
	ID           string `json:"id"`
	VehicleID    string `json:"vehicleId"`
	DriverMobile string `json:"driverMobile"`
	DriverName   string `json:"driverName"`
	Rating       string `json:"rating"`
	HasPin       bool   `json:"hasPin"`
}

type ApprochInfo struct {
	ID        string    `json:"id" gorm:"primary_key"`
	Name      string    `json:"name"`
	Number    string    `json:"number"`
	Email     string    `json:"email"`
	Message   string    `json:"message"`
	Type      string    `json:"type"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
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

type BroadcastNotificationRequests struct {
	ID               string    `json:"id" gorm:"primary_key"`
	UserType         int       `json:"user_type" gorm:"not null"`
	Title            string    `json:"title"`
	Message          string    `json:"message"`
	NotificationType string    `json:"notification_request"`
	Processed        bool      `json:"processed"`
	Data             string    `json:"data"`
	CreatedAt        time.Time `json:"created_at"`
	UpdatedAt        time.Time `json:"updated_at"`
}

type AnnouncementRequests struct {
	ID        string    `json:"id" gorm:"primary_key"`
	Title     string    `json:"title"`
	Message   string    `json:"message"`
	Type      string    `json:"type"`
	Link      string    `json:"link"`
	Processed bool      `json:"processed"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type RideRequest struct {
	ID                   string    `json:"id" gorm:"primary_key"`
	StartDatetime        string    `json:"start_datetime" gorm:"not null"`
	EstimatedEndDatetime string    `json:"estimated_end_datetime" gorm:"not null"`
	NumberOfSeats        int       `json:"number_of_seats" gorm:"not null"`
	StartLocation        string    `json:"start_location" gorm:"not null"`
	EndLocation          string    `json:"end_location" gorm:"not null"`
	RouteDetails         string    `json:"route_details"`
	ContactNumber        string    `json:"contact_number" gorm:"not null;default:'N/A'"`
	IsActive             bool      `json:"is_active" gorm:"not null"`
	CreatedAt            time.Time `json:"created_at"`
	UpdatedAt            time.Time `json:"updated_at"`
}

type Passenger struct {
	ID              string    `json:"id" gorm:"primary_key"`
	PassengerMobile string    `json:"passenger_mobile" gorm:"unique;not null"`
	PassengerName   string    `json:"passenger_name" gorm:"not null"`
	Password        string    `json:"password" gorm:"not null"`
	Gender          string    `json:"gender" gorm:"not null"`
	Status          string    `json:"status"`
	UpdateBy        string    `json:"update_by"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}

type DELPassenger struct {
	ID              string    `json:"id" gorm:"primary_key"`
	PassengerMobile string    `json:"passenger_mobile" gorm:"unique;not null"`
	PassengerName   string    `json:"passenger_name" gorm:"not null"`
	Password        string    `json:"password" gorm:"not null"`
	Gender          string    `json:"gender" gorm:"not null"`
	Status          string    `json:"status"`
	UpdateBy        string    `json:"update_by"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}

// PlatformOverview is the whole product counted in one query, the numbers an admin
// wants on the first screen before drilling into anything.
type PlatformOverview struct {
	TotalDrivers        int64 `json:"total_drivers"`
	ActiveDrivers       int64 `json:"active_drivers"`
	TotalPassengers     int64 `json:"total_passengers"`
	ActivePassengers    int64 `json:"active_passengers"`
	TotalVehicles       int64 `json:"total_vehicles"`
	SeatedVehicles      int64 `json:"seated_vehicles"`
	TotalRides          int64 `json:"total_rides"`
	ActiveRides         int64 `json:"active_rides"`
	TotalServices       int64 `json:"total_services"`
	ActiveServices      int64 `json:"active_services"`
	TotalShifts         int64 `json:"total_shifts"`
	ActiveShifts        int64 `json:"active_shifts"`
	TotalAdvertisements int64 `json:"total_advertisements"`
	OpenShiftRequests   int64 `json:"open_shift_requests"`
	PendingJoinRequests int64 `json:"pending_join_requests"`
}

////////////////////////////// PICK & DROP //////////////////////////////

// PickDropService is the organisation a driver runs once they enable Pick & Drop.
// It has exactly one owner, the driver who enabled it, and nobody else manages it.
// It is never deleted, disabling it keeps its members, vehicles and shift history.
type PickDropService struct {
	ID            string     `json:"id" gorm:"primary_key"`
	Name          string     `json:"name" gorm:"not null"`
	Description   string     `json:"description"`
	OwnerDriverID string     `json:"owner_driver_id" gorm:"index;not null"`
	Status        string     `json:"status" gorm:"index;not null"`
	DisabledAt    *time.Time `json:"disabled_at"`
	CreatedAt     time.Time  `json:"created_at"`
	UpdatedAt     time.Time  `json:"updated_at"`
}

// PickDropDriver is one request of a driver to join a service. A row is never
// reused: asking again after leaving writes a new row, so when somebody joined and
// left stays answerable. A partial unique index allows one open row per driver.
type PickDropDriver struct {
	ID        string `json:"id" gorm:"primary_key"`
	ServiceID string `json:"service_id" gorm:"index;not null"`
	DriverID  string `json:"driver_id" gorm:"index;not null"`
	// driver can drive the service's shifts, vehicles belongs only through the
	// vehicles they brought and is never put behind the wheel of a shift
	JoinType    string     `json:"join_type" gorm:"not null;default:'driver'"`
	Status      string     `json:"status" gorm:"index;not null"`
	RequestedAt time.Time  `json:"requested_at" gorm:"not null"`
	DecidedBy   string     `json:"decided_by"`
	DecidedAt   *time.Time `json:"decided_at"`
	EndedBy     string     `json:"ended_by"`
	EndedAt     *time.Time `json:"ended_at"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
}

// PickDropVehicle is a vehicle offered to a service. The owner's own vehicles are
// approved as they are added, a joining driver's vehicles wait for the owner.
type PickDropVehicle struct {
	ID          string     `json:"id" gorm:"primary_key"`
	ServiceID   string     `json:"service_id" gorm:"index;not null"`
	VehicleID   string     `json:"vehicle_id" gorm:"index;not null"`
	DriverID    string     `json:"driver_id" gorm:"index;not null"`
	Status      string     `json:"status" gorm:"index;not null"`
	RequestedAt time.Time  `json:"requested_at" gorm:"not null"`
	DecidedBy   string     `json:"decided_by"`
	DecidedAt   *time.Time `json:"decided_at"`
	EndedBy     string     `json:"ended_by"`
	EndedAt     *time.Time `json:"ended_at"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
}

// PickDropPassenger is one request of a passenger to join a service, kept after it
// is decided for the same reason as PickDropDriver.
type PickDropPassenger struct {
	ID          string     `json:"id" gorm:"primary_key"`
	ServiceID   string     `json:"service_id" gorm:"index;not null"`
	PassengerID string     `json:"passenger_id" gorm:"index;not null"`
	Status      string     `json:"status" gorm:"index;not null"`
	RequestedAt time.Time  `json:"requested_at" gorm:"not null"`
	DecidedBy   string     `json:"decided_by"`
	DecidedAt   *time.Time `json:"decided_at"`
	EndedBy     string     `json:"ended_by"`
	EndedAt     *time.Time `json:"ended_at"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
}

// PassengerAvailability is one day of a passenger's standing weekly requirement.
type PassengerAvailability struct {
	ID          string    `json:"id" gorm:"primary_key"`
	PassengerID string    `json:"passenger_id" gorm:"uniqueIndex:idx_passenger_availability_day;not null"`
	DayOfWeek   int       `json:"day_of_week" gorm:"uniqueIndex:idx_passenger_availability_day;not null"`
	IsRequired  bool      `json:"is_required" gorm:"not null"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// PassengerAvailabilityLocation is one place and time of a required day, in order.
type PassengerAvailabilityLocation struct {
	ID          string    `json:"id" gorm:"primary_key"`
	PassengerID string    `json:"passenger_id" gorm:"index;not null"`
	DayOfWeek   int       `json:"day_of_week" gorm:"not null"`
	Sequence    int       `json:"sequence" gorm:"not null"`
	Location    string    `json:"location" gorm:"not null"`
	Lat         float64   `json:"lat"`
	Lng         float64   `json:"lng"`
	Time        string    `json:"time" gorm:"not null"`
	CreatedAt   time.Time `json:"created_at"`
}

type DELPassengerAvailability struct {
	ID          string    `json:"id" gorm:"primary_key"`
	PassengerID string    `json:"passenger_id" gorm:"index;not null"`
	DayOfWeek   int       `json:"day_of_week"`
	IsRequired  bool      `json:"is_required"`
	DeletedBy   string    `json:"deleted_by"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type DELPassengerAvailabilityLocation struct {
	ID          string    `json:"id" gorm:"primary_key"`
	PassengerID string    `json:"passenger_id" gorm:"index;not null"`
	DayOfWeek   int       `json:"day_of_week"`
	Sequence    int       `json:"sequence"`
	Location    string    `json:"location"`
	Lat         float64   `json:"lat"`
	Lng         float64   `json:"lng"`
	Time        string    `json:"time"`
	CreatedAt   time.Time `json:"created_at"`
}

// PickDropAdvertisement is a public offer of a route by a service. The start, end and
// route points are kept lower cased exactly like a ride's, which is what lets the
// same trigram and array searches find them.
type PickDropAdvertisement struct {
	ID            string         `json:"id" gorm:"primary_key"`
	ServiceID     string         `json:"service_id" gorm:"index;not null"`
	Title         string         `json:"title" gorm:"not null"`
	Description   string         `json:"description"`
	Fare          float64        `json:"fare" gorm:"not null;default:0"`
	DaysOfWeek    pq.Int64Array  `json:"days_of_week" gorm:"type:integer[];not null"`
	StartTime     string         `json:"start_time" gorm:"not null"`
	EndTime       string         `json:"end_time" gorm:"not null"`
	StartLocation string         `json:"start_location" gorm:"not null"`
	EndLocation   string         `json:"end_location" gorm:"not null"`
	RoutePoints   pq.StringArray `json:"route_points" gorm:"type:text[]"`
	CreatedBy     string         `json:"created_by"`
	CreatedAt     time.Time      `json:"created_at"`
	UpdatedAt     time.Time      `json:"updated_at"`
}

type PickDropAdvertisementLocation struct {
	ID              string    `json:"id" gorm:"primary_key"`
	AdvertisementID string    `json:"advertisement_id" gorm:"index;not null"`
	Sequence        int       `json:"sequence" gorm:"not null"`
	Location        string    `json:"location" gorm:"not null"`
	Lat             float64   `json:"lat"`
	Lng             float64   `json:"lng"`
	Time            string    `json:"time" gorm:"not null"`
	CreatedAt       time.Time `json:"created_at"`
}

type DELPickDropAdvertisement struct {
	ID            string         `json:"id" gorm:"primary_key"`
	ServiceID     string         `json:"service_id" gorm:"index"`
	Title         string         `json:"title"`
	Description   string         `json:"description"`
	Fare          float64        `json:"fare"`
	DaysOfWeek    pq.Int64Array  `json:"days_of_week" gorm:"type:integer[]"`
	StartTime     string         `json:"start_time"`
	EndTime       string         `json:"end_time"`
	StartLocation string         `json:"start_location"`
	EndLocation   string         `json:"end_location"`
	RoutePoints   pq.StringArray `json:"route_points" gorm:"type:text[]"`
	CreatedBy     string         `json:"created_by"`
	DeletedBy     string         `json:"deleted_by"`
	CreatedAt     time.Time      `json:"created_at"`
	UpdatedAt     time.Time      `json:"updated_at"`
}

type DELPickDropAdvertisementLocation struct {
	ID              string    `json:"id" gorm:"primary_key"`
	AdvertisementID string    `json:"advertisement_id" gorm:"index"`
	Sequence        int       `json:"sequence"`
	Location        string    `json:"location"`
	Lat             float64   `json:"lat"`
	Lng             float64   `json:"lng"`
	Time            string    `json:"time"`
	CreatedAt       time.Time `json:"created_at"`
}

// PickDropServiceDetails is a service as a list shows it, with its size counted by
// subselects and the caller's own standing in it.
type PickDropServiceDetails struct {
	ID             string `json:"id"`
	Name           string `json:"name"`
	Description    string `json:"description"`
	OwnerDriverID  string `json:"owner_driver_id"`
	OwnerName      string `json:"owner_name"`
	OwnerMobile    string `json:"owner_mobile"`
	Status         string `json:"status"`
	DriverCount    int64  `json:"driver_count"`
	VehicleCount   int64  `json:"vehicle_count"`
	PassengerCount int64  `json:"passenger_count"`
	// set only by the admin service listing, zero everywhere else
	ShiftCount int64     `json:"shift_count"`
	MyStatus   string    `json:"my_status"`
	CreatedAt  time.Time `json:"created_at"`
}

type PickDropDriverDetails struct {
	ID           string     `json:"id"`
	ServiceID    string     `json:"service_id"`
	DriverID     string     `json:"driver_id"`
	DriverName   string     `json:"driver_name"`
	DriverMobile string     `json:"driver_mobile"`
	Rating       string     `json:"rating"`
	JoinType     string     `json:"join_type"`
	Status       string     `json:"status"`
	RequestedAt  time.Time  `json:"requested_at"`
	DecidedAt    *time.Time `json:"decided_at"`
	EndedAt      *time.Time `json:"ended_at"`
}

type PickDropVehicleDetails struct {
	ID            string     `json:"id"`
	ServiceID     string     `json:"service_id"`
	VehicleID     string     `json:"vehicle_id"`
	VehicleNumber string     `json:"vehicle_number"`
	VehicleInfo   string     `json:"vehicle_info"`
	NumberOfSeats int        `json:"number_of_seats"`
	HasAC         bool       `json:"has_ac"`
	HasHeating    bool       `json:"has_heating"`
	DriverID      string     `json:"driver_id"`
	DriverName    string     `json:"driver_name"`
	DriverMobile  string     `json:"driver_mobile"`
	OwnerJoinType string     `json:"owner_join_type"`
	Status        string     `json:"status"`
	RequestedAt   time.Time  `json:"requested_at"`
	DecidedAt     *time.Time `json:"decided_at"`
	EndedAt       *time.Time `json:"ended_at"`
	ActiveShifts  int64      `json:"active_shifts"`
}

type PickDropPassengerDetails struct {
	ID              string     `json:"id"`
	ServiceID       string     `json:"service_id"`
	PassengerID     string     `json:"passenger_id"`
	PassengerName   string     `json:"passenger_name"`
	PassengerMobile string     `json:"passenger_mobile"`
	Gender          string     `json:"gender"`
	Status          string     `json:"status"`
	RequestedAt     time.Time  `json:"requested_at"`
	DecidedAt       *time.Time `json:"decided_at"`
	EndedAt         *time.Time `json:"ended_at"`
	ActiveShifts    int64      `json:"active_shifts"`
}

// AvailableDriverDetails is somebody the owner can put behind the wheel of a shift:
// the owner themselves or an approved driver of the service.
type AvailableDriverDetails struct {
	DriverID     string     `json:"driver_id"`
	DriverName   string     `json:"driver_name"`
	DriverMobile string     `json:"driver_mobile"`
	Rating       string     `json:"rating"`
	IsOwner      bool       `json:"is_owner"`
	JoinedAt     *time.Time `json:"joined_at"`
	ActiveShifts int64      `json:"active_shifts"`
}

type AdvertisementDetails struct {
	ID            string        `json:"id"`
	ServiceID     string        `json:"service_id"`
	ServiceName   string        `json:"service_name"`
	OwnerDriverID string        `json:"owner_driver_id"`
	OwnerName     string        `json:"owner_name"`
	OwnerMobile   string        `json:"owner_mobile"`
	Title         string        `json:"title"`
	Description   string        `json:"description"`
	Fare          float64       `json:"fare"`
	DaysOfWeek    pq.Int64Array `json:"days_of_week" gorm:"type:integer[]"`
	StartTime     string        `json:"start_time"`
	EndTime       string        `json:"end_time"`
	VehicleCount  int64         `json:"vehicle_count"`
	CreatedAt     time.Time     `json:"created_at"`
}

////////////////////////////// SHIFTS //////////////////////////////

// ShiftRequest is a passenger's recurring requirement put out for owners to find,
// the pick & drop counterpart of a ride request. A request addressed to one
// service is only shown to that service's owner.
type ShiftRequest struct {
	ID            string         `json:"id" gorm:"primary_key"`
	PassengerID   string         `json:"passenger_id" gorm:"index;not null"`
	ServiceID     string         `json:"service_id" gorm:"index"`
	ContactNumber string         `json:"contact_number" gorm:"not null"`
	Note          string         `json:"note"`
	DaysOfWeek    pq.Int64Array  `json:"days_of_week" gorm:"type:integer[];not null"`
	StartTime     string         `json:"start_time" gorm:"not null"`
	EndTime       string         `json:"end_time" gorm:"not null"`
	StartLocation string         `json:"start_location" gorm:"not null"`
	EndLocation   string         `json:"end_location" gorm:"not null"`
	RoutePoints   pq.StringArray `json:"route_points" gorm:"type:text[]"`
	CreatedAt     time.Time      `json:"created_at"`
	UpdatedAt     time.Time      `json:"updated_at"`
}

type ShiftRequestLocation struct {
	ID             string    `json:"id" gorm:"primary_key"`
	ShiftRequestID string    `json:"shift_request_id" gorm:"index;not null"`
	Sequence       int       `json:"sequence" gorm:"not null"`
	Location       string    `json:"location" gorm:"not null"`
	Lat            float64   `json:"lat"`
	Lng            float64   `json:"lng"`
	Time           string    `json:"time" gorm:"not null"`
	CreatedAt      time.Time `json:"created_at"`
}

type DELShiftRequest struct {
	ID            string         `json:"id" gorm:"primary_key"`
	PassengerID   string         `json:"passenger_id" gorm:"index"`
	ServiceID     string         `json:"service_id"`
	ContactNumber string         `json:"contact_number"`
	Note          string         `json:"note"`
	DaysOfWeek    pq.Int64Array  `json:"days_of_week" gorm:"type:integer[]"`
	StartTime     string         `json:"start_time"`
	EndTime       string         `json:"end_time"`
	StartLocation string         `json:"start_location"`
	EndLocation   string         `json:"end_location"`
	RoutePoints   pq.StringArray `json:"route_points" gorm:"type:text[]"`
	DeletedBy     string         `json:"deleted_by"`
	CreatedAt     time.Time      `json:"created_at"`
	UpdatedAt     time.Time      `json:"updated_at"`
}

type DELShiftRequestLocation struct {
	ID             string    `json:"id" gorm:"primary_key"`
	ShiftRequestID string    `json:"shift_request_id" gorm:"index"`
	Sequence       int       `json:"sequence"`
	Location       string    `json:"location"`
	Lat            float64   `json:"lat"`
	Lng            float64   `json:"lng"`
	Time           string    `json:"time"`
	CreatedAt      time.Time `json:"created_at"`
}

// Shift is the transportation assignment itself. It is ONE row however many days
// it runs: it recurs on its days of the week between its start date and its
// optional end date, from the time of its first location to the time of its last.
// OccupiedSeats is kept in step with the active shift passengers inside the same
// transaction and a CHECK constraint refuses it ever passing SeatCapacity.
type Shift struct {
	ID            string        `json:"id" gorm:"primary_key"`
	ServiceID     string        `json:"service_id" gorm:"index;not null"`
	Name          string        `json:"name"`
	DriverID      string        `json:"driver_id" gorm:"index;not null"`
	VehicleID     string        `json:"vehicle_id" gorm:"index;not null"`
	DaysOfWeek    pq.Int64Array `json:"days_of_week" gorm:"type:integer[];not null"`
	StartDate     string        `json:"start_date" gorm:"not null"`
	EndDate       string        `json:"end_date" gorm:"not null;default:''"`
	StartTime     string        `json:"start_time" gorm:"not null"`
	EndTime       string        `json:"end_time" gorm:"not null"`
	SeatCapacity  int           `json:"seat_capacity" gorm:"not null"`
	OccupiedSeats int           `json:"occupied_seats" gorm:"not null;default:0"`
	Status        string        `json:"status" gorm:"index;not null"`
	CreatedBy     string        `json:"created_by"`
	CreatedAt     time.Time     `json:"created_at"`
	UpdatedAt     time.Time     `json:"updated_at"`
}

// ShiftLocation is one stop of a shift's route, in the order it is driven.
type ShiftLocation struct {
	ID        string    `json:"id" gorm:"primary_key"`
	ShiftID   string    `json:"shift_id" gorm:"index;not null"`
	Sequence  int       `json:"sequence" gorm:"not null"`
	Location  string    `json:"location" gorm:"not null"`
	Lat       float64   `json:"lat"`
	Lng       float64   `json:"lng"`
	Time      string    `json:"time" gorm:"not null"`
	CreatedAt time.Time `json:"created_at"`
}

// ShiftPassenger is a passenger's place on a shift. Removing somebody marks the row
// removed rather than deleting it, so the row itself is the record of when they
// were added and taken off, and a replacement passenger always gets a row of their
// own that no earlier attendance can ever be attached to.
type ShiftPassenger struct {
	ID               string     `json:"id" gorm:"primary_key"`
	ShiftID          string     `json:"shift_id" gorm:"index;not null"`
	ServiceID        string     `json:"service_id" gorm:"index;not null"`
	PassengerID      string     `json:"passenger_id" gorm:"index;not null"`
	LocationSequence int        `json:"location_sequence" gorm:"not null"`
	Status           string     `json:"status" gorm:"index;not null"`
	AddedBy          string     `json:"added_by"`
	AddedAt          time.Time  `json:"added_at" gorm:"not null"`
	RemovedBy        string     `json:"removed_by"`
	RemovedAt        *time.Time `json:"removed_at"`
	RemovalReason    string     `json:"removal_reason"`
	CreatedAt        time.Time  `json:"created_at"`
	UpdatedAt        time.Time  `json:"updated_at"`
}

// ShiftOccurrence is one day a shift actually runs. It is written ahead of the day
// by the scheduler, and it freezes who drove, in which vehicle, along which route,
// so a later edit of the shift never rewrites a trip that already happened. The
// reminder claim lives on the row, which is what makes the reminder safe to retry.
type ShiftOccurrence struct {
	ID               string     `json:"id" gorm:"primary_key"`
	ShiftID          string     `json:"shift_id" gorm:"uniqueIndex:idx_shift_occurrence_date;not null"`
	ServiceID        string     `json:"service_id" gorm:"index;not null"`
	OccurrenceDate   string     `json:"occurrence_date" gorm:"uniqueIndex:idx_shift_occurrence_date;not null"`
	StartTime        string     `json:"start_time" gorm:"not null"`
	EndTime          string     `json:"end_time" gorm:"not null"`
	StartsAt         string     `json:"starts_at" gorm:"index;not null"`
	EndsAt           string     `json:"ends_at" gorm:"not null"`
	DriverID         string     `json:"driver_id" gorm:"index;not null"`
	VehicleID        string     `json:"vehicle_id" gorm:"not null"`
	ShiftName        string     `json:"shift_name"`
	ServiceName      string     `json:"service_name"`
	DriverName       string     `json:"driver_name"`
	DriverMobile     string     `json:"driver_mobile"`
	VehicleNumber    string     `json:"vehicle_number"`
	Route            string     `json:"route"`
	Status           string     `json:"status" gorm:"index;not null"`
	DriverReminderAt *time.Time `json:"driver_reminder_at"`
	CreatedAt        time.Time  `json:"created_at"`
	UpdatedAt        time.Time  `json:"updated_at"`
}

// ShiftAttendance is one passenger on one occurrence. Present is the default, so
// the row starts present and a passenger marks it absent. It belongs to the shift,
// the passenger and the date, never to the seat, so nobody inherits it.
type ShiftAttendance struct {
	ID               string     `json:"id" gorm:"primary_key"`
	OccurrenceID     string     `json:"occurrence_id" gorm:"uniqueIndex:idx_shift_attendance_passenger;not null"`
	ShiftID          string     `json:"shift_id" gorm:"index;not null"`
	PassengerID      string     `json:"passenger_id" gorm:"uniqueIndex:idx_shift_attendance_passenger;index;not null"`
	OccurrenceDate   string     `json:"occurrence_date" gorm:"not null"`
	LocationSequence int        `json:"location_sequence"`
	Location         string     `json:"location"`
	LocationTime     string     `json:"location_time"`
	Status           string     `json:"status" gorm:"not null"`
	MarkedAt         *time.Time `json:"marked_at"`
	ReminderAt       *time.Time `json:"reminder_at"`
	CreatedAt        time.Time  `json:"created_at"`
	UpdatedAt        time.Time  `json:"updated_at"`
}

// ShiftDriverUpdate is a location or status message a driver sent to the
// passengers of their shift, kept so what was said and to how many people can be
// answered later.
type ShiftDriverUpdate struct {
	ID           string    `json:"id" gorm:"primary_key"`
	ShiftID      string    `json:"shift_id" gorm:"index;not null"`
	OccurrenceID string    `json:"occurrence_id" gorm:"index"`
	DriverID     string    `json:"driver_id" gorm:"index;not null"`
	Message      string    `json:"message"`
	Location     string    `json:"location"`
	Lat          float64   `json:"lat"`
	Lng          float64   `json:"lng"`
	Recipients   int       `json:"recipients"`
	CreatedAt    time.Time `json:"created_at"`
}

// DELShift keeps a deleted shift. The original row is hard deleted, the same
// archive then delete pattern rides already follow.
type DELShift struct {
	ID            string        `json:"id" gorm:"primary_key"`
	ServiceID     string        `json:"service_id" gorm:"index"`
	Name          string        `json:"name"`
	DriverID      string        `json:"driver_id"`
	VehicleID     string        `json:"vehicle_id"`
	DaysOfWeek    pq.Int64Array `json:"days_of_week" gorm:"type:integer[]"`
	StartDate     string        `json:"start_date"`
	EndDate       string        `json:"end_date"`
	StartTime     string        `json:"start_time"`
	EndTime       string        `json:"end_time"`
	SeatCapacity  int           `json:"seat_capacity"`
	OccupiedSeats int           `json:"occupied_seats"`
	Status        string        `json:"status"`
	CreatedBy     string        `json:"created_by"`
	DeletedBy     string        `json:"deleted_by"`
	DeletedAt     time.Time     `json:"deleted_at"`
	CreatedAt     time.Time     `json:"created_at"`
	UpdatedAt     time.Time     `json:"updated_at"`
}

type DELShiftLocation struct {
	ID        string    `json:"id" gorm:"primary_key"`
	ShiftID   string    `json:"shift_id" gorm:"index"`
	Sequence  int       `json:"sequence"`
	Location  string    `json:"location"`
	Lat       float64   `json:"lat"`
	Lng       float64   `json:"lng"`
	Time      string    `json:"time"`
	CreatedAt time.Time `json:"created_at"`
}

// ShiftDetails is the joined read projection of a shift with its service, driver
// and vehicle.
type ShiftDetails struct {
	ID            string `json:"id"`
	Name          string `json:"name"`
	ServiceID     string `json:"service_id"`
	ServiceName   string `json:"service_name"`
	OwnerDriverID string `json:"owner_driver_id"`
	DriverID      string `json:"driver_id"`
	DriverName    string `json:"driver_name"`
	DriverMobile  string `json:"driver_mobile"`
	VehicleID     string `json:"vehicle_id"`
	VehicleNumber string `json:"vehicle_number"`
	VehicleInfo   string `json:"vehicle_info"`
	// the driver who brought the vehicle, who is not always the one driving it
	VehicleOwnerID   string        `json:"vehicle_owner_id"`
	VehicleOwnerName string        `json:"vehicle_owner_name"`
	DaysOfWeek       pq.Int64Array `json:"days_of_week" gorm:"type:integer[]"`
	StartDate        string        `json:"start_date"`
	EndDate          string        `json:"end_date"`
	StartTime        string        `json:"start_time"`
	EndTime          string        `json:"end_time"`
	SeatCapacity     int           `json:"seat_capacity"`
	OccupiedSeats    int           `json:"occupied_seats"`
	Status           string        `json:"status"`
	CreatedAt        time.Time     `json:"created_at"`
	UpdatedAt        time.Time     `json:"updated_at"`
}

type ShiftPassengerDetails struct {
	ID               string    `json:"id"`
	ShiftID          string    `json:"shift_id"`
	PassengerID      string    `json:"passenger_id"`
	PassengerName    string    `json:"passenger_name"`
	PassengerMobile  string    `json:"passenger_mobile"`
	LocationSequence int       `json:"location_sequence"`
	Status           string    `json:"status"`
	AddedAt          time.Time `json:"added_at"`
}

// ShiftHistoryDetails is a shift as the owner's history shows it, live or deleted.
type ShiftHistoryDetails struct {
	ID            string        `json:"id"`
	Name          string        `json:"name"`
	DriverName    string        `json:"driver_name"`
	VehicleNumber string        `json:"vehicle_number"`
	DaysOfWeek    pq.Int64Array `json:"days_of_week" gorm:"type:integer[]"`
	StartDate     string        `json:"start_date"`
	EndDate       string        `json:"end_date"`
	StartTime     string        `json:"start_time"`
	EndTime       string        `json:"end_time"`
	Status        string        `json:"status"`
	CreatedAt     time.Time     `json:"created_at"`
	DeletedAt     *time.Time    `json:"deleted_at"`
}

// ShiftParticipationDetails is one stint of one passenger on one shift, the unit of
// the owner's passenger history.
type ShiftParticipationDetails struct {
	ID               string     `json:"id"`
	ShiftID          string     `json:"shift_id"`
	ShiftName        string     `json:"shift_name"`
	ShiftStatus      string     `json:"shift_status"`
	PassengerID      string     `json:"passenger_id"`
	PassengerName    string     `json:"passenger_name"`
	PassengerMobile  string     `json:"passenger_mobile"`
	LocationSequence int        `json:"location_sequence"`
	Status           string     `json:"status"`
	RemovalReason    string     `json:"removal_reason"`
	JoinedServiceAt  *time.Time `json:"joined_service_at"`
	AddedAt          time.Time  `json:"added_at"`
	RemovedAt        *time.Time `json:"removed_at"`
	TripsPresent     int64      `json:"trips_present"`
	TripsAbsent      int64      `json:"trips_absent"`
}

type ShiftOccurrenceDetails struct {
	ID             string `json:"id"`
	ShiftID        string `json:"shift_id"`
	ShiftName      string `json:"shift_name"`
	ServiceName    string `json:"service_name"`
	OccurrenceDate string `json:"occurrence_date"`
	StartTime      string `json:"start_time"`
	EndTime        string `json:"end_time"`
	DriverName     string `json:"driver_name"`
	VehicleNumber  string `json:"vehicle_number"`
	Status         string `json:"status"`
	PresentCount   int64  `json:"present_count"`
	AbsentCount    int64  `json:"absent_count"`
}

type ShiftAttendanceDetails struct {
	PassengerID      string     `json:"passenger_id"`
	PassengerName    string     `json:"passenger_name"`
	PassengerMobile  string     `json:"passenger_mobile"`
	LocationSequence int        `json:"location_sequence"`
	Location         string     `json:"location"`
	LocationTime     string     `json:"location_time"`
	Status           string     `json:"status"`
	MarkedAt         *time.Time `json:"marked_at"`
}

// TravelHistoryDetails is one trip a passenger was on, read entirely from the frozen
// occurrence so it stays true after the shift changes, or is deleted.
type TravelHistoryDetails struct {
	OccurrenceID     string `json:"occurrence_id"`
	ShiftID          string `json:"shift_id"`
	ShiftName        string `json:"shift_name"`
	ServiceName      string `json:"service_name"`
	OccurrenceDate   string `json:"occurrence_date"`
	StartTime        string `json:"start_time"`
	EndTime          string `json:"end_time"`
	DriverName       string `json:"driver_name"`
	DriverMobile     string `json:"driver_mobile"`
	VehicleNumber    string `json:"vehicle_number"`
	Route            string `json:"route"`
	LocationSequence int    `json:"location_sequence"`
	Location         string `json:"location"`
	LocationTime     string `json:"location_time"`
	Status           string `json:"status"`
}

type ShiftRequestDetails struct {
	ID            string        `json:"id"`
	PassengerID   string        `json:"passenger_id"`
	PassengerName string        `json:"passenger_name"`
	ServiceID     string        `json:"service_id"`
	ServiceName   string        `json:"service_name"`
	ContactNumber string        `json:"contact_number"`
	Note          string        `json:"note"`
	DaysOfWeek    pq.Int64Array `json:"days_of_week" gorm:"type:integer[]"`
	StartTime     string        `json:"start_time"`
	EndTime       string        `json:"end_time"`
	CreatedAt     time.Time     `json:"created_at"`
}
