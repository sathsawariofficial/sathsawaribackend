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

// PassengerLocationPreference is the standing travel form a passenger fills in:
// one row per (passenger, day of week, direction), so a passenger can travel in
// the morning but not in the evening, or be picked from one place and dropped at
// another. It is reference data the manager reads while building a shift, it does
// not create any shift by itself.
// DELPassengerLocationPreference keeps the travel form of a deleted passenger, the
// same way a deleted driver and their vehicles are kept. The form is the record of
// what that person had actually asked the fleet for.
type DELPassengerLocationPreference struct {
	ID            string    `json:"id" gorm:"primary_key"`
	PassengerID   string    `json:"passenger_id" gorm:"index;not null"`
	DayOfWeek     int       `json:"day_of_week"`
	Direction     string    `json:"direction"`
	IsEnabled     bool      `json:"is_enabled"`
	Location      string    `json:"location"`
	Lat           float64   `json:"lat"`
	Lng           float64   `json:"lng"`
	ScheduledTime string    `json:"scheduled_time"`
	UpdateBy      string    `json:"update_by"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

// DELRole keeps a role an admin deleted, together with the permission codes it held
// at that moment. Who was allowed to do what is worth being able to answer later,
// and the codes are kept inline so the record stands on its own even after the
// permissions themselves change.
type DELRole struct {
	ID              string         `json:"id" gorm:"primary_key"`
	Name            string         `json:"name" gorm:"not null"`
	Description     string         `json:"description"`
	PermissionCodes pq.StringArray `json:"permission_codes" gorm:"type:text[]"`
	UpdateBy        string         `json:"update_by"`
	CreatedAt       time.Time      `json:"created_at"`
	UpdatedAt       time.Time      `json:"updated_at"`
}

type PassengerLocationPreference struct {
	ID            string    `json:"id" gorm:"primary_key"`
	PassengerID   string    `json:"passenger_id" gorm:"index:idx_passenger_day_direction,unique;not null"`
	DayOfWeek     int       `json:"day_of_week" gorm:"index:idx_passenger_day_direction,unique;not null"`
	Direction     string    `json:"direction" gorm:"index:idx_passenger_day_direction,unique;not null"`
	IsEnabled     bool      `json:"is_enabled" gorm:"not null;default:true"`
	Location      string    `json:"location"`
	Lat           float64   `json:"lat"`
	Lng           float64   `json:"lng"`
	ScheduledTime string    `json:"scheduled_time"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

type Role struct {
	ID          string    `json:"id" gorm:"primary_key"`
	Name        string    `json:"name" gorm:"unique;not null"`
	Description string    `json:"description"`
	IsSystem    bool      `json:"is_system" gorm:"not null;default:false"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type Permission struct {
	ID          string    `json:"id" gorm:"primary_key"`
	Code        string    `json:"code" gorm:"unique;not null"`
	Description string    `json:"description"`
	IsSystem    bool      `json:"is_system" gorm:"not null;default:false"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type RolePermission struct {
	ID           string    `json:"id" gorm:"primary_key"`
	RoleID       string    `json:"role_id" gorm:"index:idx_role_permission,unique;not null"`
	PermissionID string    `json:"permission_id" gorm:"index:idx_role_permission,unique;not null"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

type Group struct {
	ID            string    `json:"id" gorm:"primary_key"`
	Name          string    `json:"name" gorm:"not null"`
	Description   string    `json:"description"`
	OwnerDriverID string    `json:"owner_driver_id" gorm:"index;not null"`
	Status        string    `json:"status" gorm:"index;not null"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

// GroupMember is a driver's membership of a group. RoleID is only filled for the
// owner and for appointed sub managers, a plain joined driver carries no role.
type GroupMember struct {
	ID        string     `json:"id" gorm:"primary_key"`
	GroupID   string     `json:"group_id" gorm:"index:idx_group_member,unique;not null"`
	DriverID  string     `json:"driver_id" gorm:"index:idx_group_member,unique;not null"`
	RoleID    string     `json:"role_id" gorm:"index"`
	JoinType  string     `json:"join_type" gorm:"not null"`
	Status    string     `json:"status" gorm:"index;not null"`
	DecidedBy string     `json:"decided_by"`
	DecidedAt *time.Time `json:"decided_at"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
}

type GroupVehicle struct {
	ID        string     `json:"id" gorm:"primary_key"`
	GroupID   string     `json:"group_id" gorm:"index:idx_group_vehicle,unique;not null"`
	VehicleID string     `json:"vehicle_id" gorm:"index:idx_group_vehicle,unique;not null"`
	DriverID  string     `json:"driver_id" gorm:"index;not null"`
	Status    string     `json:"status" gorm:"index;not null"`
	DecidedBy string     `json:"decided_by"`
	DecidedAt *time.Time `json:"decided_at"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
}

type GroupPassenger struct {
	ID          string     `json:"id" gorm:"primary_key"`
	GroupID     string     `json:"group_id" gorm:"index:idx_group_passenger,unique;not null"`
	PassengerID string     `json:"passenger_id" gorm:"index:idx_group_passenger,unique;not null"`
	Status      string     `json:"status" gorm:"index;not null"`
	DecidedBy   string     `json:"decided_by"`
	DecidedAt   *time.Time `json:"decided_at"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
}

// Shift is one directional trip of one vehicle on one day. A pickup and a drop off
// are always two separate shifts even when the same passengers travel in both.
type Shift struct {
	ID                   string    `json:"id" gorm:"primary_key"`
	GroupID              string    `json:"group_id" gorm:"index;not null"`
	VehicleID            string    `json:"vehicle_id" gorm:"index;not null"`
	DriverID             string    `json:"driver_id" gorm:"index;not null"`
	Direction            string    `json:"direction" gorm:"index;not null"`
	StartDatetime        string    `json:"start_datetime" gorm:"index;not null"`
	EstimatedEndDatetime string    `json:"estimated_end_datetime" gorm:"not null"`
	StartLocation        string    `json:"start_location" gorm:"not null"`
	EndLocation          string    `json:"end_location" gorm:"not null"`
	NumberOfSeats        int       `json:"number_of_seats" gorm:"not null"`
	SeatsTaken           int       `json:"seats_taken" gorm:"not null;default:0"`
	RouteDetails         string    `json:"route_details"`
	TemplateID           string    `json:"template_id"`
	CreatedByDriverID    string    `json:"created_by_driver_id" gorm:"not null"`
	IsActive             bool      `json:"is_active" gorm:"index;not null"`
	CreatedAt            time.Time `json:"created_at"`
	UpdatedAt            time.Time `json:"updated_at"`

	// forign key relation
	Stops []ShiftStop `json:"stops" gorm:"foreignKey:ShiftID;references:ID"`
	Seats []ShiftSeat `json:"seats" gorm:"foreignKey:ShiftID;references:ID"`
}

// ShiftStop is one waypoint of a shift's route, in the order the manager or sub
// manager arranged them.
type ShiftStop struct {
	ID             string    `json:"id" gorm:"primary_key"`
	ShiftID        string    `json:"shift_id" gorm:"index;not null"`
	SequenceNumber int       `json:"sequence_number" gorm:"not null"`
	Location       string    `json:"location" gorm:"not null"`
	Lat            float64   `json:"lat"`
	Lng            float64   `json:"lng"`
	ScheduledTime  string    `json:"scheduled_time"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

// ShiftSeat is one physical seat of the vehicle for one shift. Gender is locked on
// the seat, a seat reserved for one gender can never be given to the other.
type ShiftSeat struct {
	ID          string    `json:"id" gorm:"primary_key"`
	ShiftID     string    `json:"shift_id" gorm:"index:idx_shift_seat,unique;not null"`
	SeatNumber  int       `json:"seat_number" gorm:"index:idx_shift_seat,unique;not null"`
	Gender      string    `json:"gender"`
	PassengerID string    `json:"passenger_id" gorm:"index"`
	StopID      string    `json:"stop_id" gorm:"index"`
	Status      string    `json:"status" gorm:"not null"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type ShiftTemplate struct {
	ID                string        `json:"id" gorm:"primary_key"`
	ShiftID           string        `json:"shift_id"`
	GroupID           string        `json:"group_id" gorm:"index;not null"`
	Name              string        `json:"name"`
	VehicleID         string        `json:"vehicle_id" gorm:"not null"`
	DriverID          string        `json:"driver_id" gorm:"not null"`
	Direction         string        `json:"direction" gorm:"not null"`
	StartDatetime     string        `json:"start_datetime" gorm:"not null"`
	EstimatedEndTime  string        `json:"estimated_end_datetime" gorm:"not null"`
	StartLocation     string        `json:"start_location"`
	EndLocation       string        `json:"end_location"`
	NumberOfSeats     int           `json:"number_of_seats"`
	RouteDetails      string        `json:"route_details"`
	DaysOfWeek        pq.Int64Array `json:"days_of_week" gorm:"type:integer[]"`
	CreatedByDriverID string        `json:"created_by_driver_id"`
	CreatedAt         time.Time     `json:"created_at"`
	UpdatedAt         time.Time     `json:"updated_at"`

	// forign key relation
	Vehicle Vehicle             `json:"vehicle" gorm:"foreignKey:VehicleID;references:ID"`
	Stops   []ShiftTemplateStop `json:"stops" gorm:"foreignKey:ShiftTemplateID;references:ID"`
	Seats   []ShiftTemplateSeat `json:"seats" gorm:"foreignKey:ShiftTemplateID;references:ID"`
}

type ShiftTemplateStop struct {
	ID              string    `json:"id" gorm:"primary_key"`
	ShiftTemplateID string    `json:"shift_template_id" gorm:"index;not null"`
	SequenceNumber  int       `json:"sequence_number" gorm:"not null"`
	Location        string    `json:"location" gorm:"not null"`
	Lat             float64   `json:"lat"`
	Lng             float64   `json:"lng"`
	ScheduledTime   string    `json:"scheduled_time"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}

// ShiftTemplateSeat points at its stop by sequence number rather than by id, so a
// template stays valid when it is used to build a brand new shift.
type ShiftTemplateSeat struct {
	ID              string    `json:"id" gorm:"primary_key"`
	ShiftTemplateID string    `json:"shift_template_id" gorm:"index;not null"`
	SeatNumber      int       `json:"seat_number" gorm:"not null"`
	Gender          string    `json:"gender"`
	PassengerID     string    `json:"passenger_id"`
	StopSequence    int       `json:"stop_sequence"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}

// ShiftDetails is the joined read projection of a shift, it carries the driver and
// the creating manager's contact details so they can be shown and notified.
type ShiftDetails struct {
	ID                   string    `json:"id"`
	GroupID              string    `json:"group_id"`
	GroupName            string    `json:"group_name"`
	VehicleID            string    `json:"vehicle_id"`
	VehicleNumber        string    `json:"vehicle_number"`
	VehicleInfo          string    `json:"vehicle_info"`
	HasAC                bool      `json:"has_ac"`
	HasHeating           bool      `json:"has_heating"`
	DriverID             string    `json:"driver_id"`
	DriverName           string    `json:"driver_name"`
	DriverMobile         string    `json:"driver_mobile"`
	Rating               string    `json:"rating"`
	Direction            string    `json:"direction"`
	StartDatetime        string    `json:"start_datetime"`
	EstimatedEndDatetime string    `json:"estimated_end_datetime"`
	StartLocation        string    `json:"start_location"`
	EndLocation          string    `json:"end_location"`
	NumberOfSeats        int       `json:"number_of_seats"`
	SeatsTaken           int       `json:"seats_taken"`
	RouteDetails         string    `json:"route_details"`
	CreatedByDriverID    string    `json:"created_by_driver_id"`
	CreatedByName        string    `json:"created_by_name"`
	CreatedByMobile      string    `json:"created_by_mobile"`
	IsActive             bool      `json:"is_active"`
	CreatedAt            time.Time `json:"created_at"`
	UpdatedAt            time.Time `json:"updated_at"`
}

// ShiftSeatDetails is the joined read projection of a seat with its passenger and
// the stop that seat is picked from or dropped at.
type ShiftSeatDetails struct {
	ID              string  `json:"id"`
	ShiftID         string  `json:"shift_id"`
	SeatNumber      int     `json:"seat_number"`
	Gender          string  `json:"gender"`
	Status          string  `json:"status"`
	PassengerID     string  `json:"passenger_id"`
	PassengerName   string  `json:"passenger_name"`
	PassengerMobile string  `json:"passenger_mobile"`
	StopID          string  `json:"stop_id"`
	SequenceNumber  int     `json:"sequence_number"`
	Location        string  `json:"location"`
	Lat             float64 `json:"lat"`
	Lng             float64 `json:"lng"`
	ScheduledTime   string  `json:"scheduled_time"`
}

// GroupMemberDetails is the joined read projection of a driver's membership.
type GroupMemberDetails struct {
	ID           string    `json:"id"`
	GroupID      string    `json:"group_id"`
	DriverID     string    `json:"driver_id"`
	DriverName   string    `json:"driver_name"`
	DriverMobile string    `json:"driver_mobile"`
	Rating       string    `json:"rating"`
	RoleID       string    `json:"role_id"`
	RoleName     string    `json:"role_name"`
	JoinType     string    `json:"join_type"`
	Status       string    `json:"status"`
	CreatedAt    time.Time `json:"created_at"`
}

// GroupVehicleDetails is the joined read projection of a vehicle in a group.
type GroupVehicleDetails struct {
	ID            string    `json:"id"`
	GroupID       string    `json:"group_id"`
	VehicleID     string    `json:"vehicle_id"`
	VehicleNumber string    `json:"vehicle_number"`
	VehicleInfo   string    `json:"vehicle_info"`
	NumberOfSeats int       `json:"number_of_seats"`
	HasAC         bool      `json:"has_ac"`
	HasHeating    bool      `json:"has_heating"`
	DriverID      string    `json:"driver_id"`
	DriverName    string    `json:"driver_name"`
	DriverMobile  string    `json:"driver_mobile"`
	Status        string    `json:"status"`
	CreatedAt     time.Time `json:"created_at"`
}

// PlatformOverview is the whole product counted in one query, the numbers an admin
// wants on the first screen before drilling into anything.
type PlatformOverview struct {
	TotalDrivers     int64 `json:"total_drivers"`
	ActiveDrivers    int64 `json:"active_drivers"`
	TotalPassengers  int64 `json:"total_passengers"`
	ActivePassengers int64 `json:"active_passengers"`
	TotalVehicles    int64 `json:"total_vehicles"`
	SeatedVehicles   int64 `json:"seated_vehicles"`
	TotalGroups      int64 `json:"total_groups"`
	ActiveGroups     int64 `json:"active_groups"`
	TotalShifts      int64 `json:"total_shifts"`
	UpcomingShifts   int64 `json:"upcoming_shifts"`
	TotalRides       int64 `json:"total_rides"`
	ActiveRides      int64 `json:"active_rides"`
	PendingRequests  int64 `json:"pending_requests"`
}

// AdminGroupOverview is a fleet as an admin sees it in a list: who runs it, how big
// it is, and how much work it is carrying.
type AdminGroupOverview struct {
	ID             string    `json:"id"`
	Name           string    `json:"name"`
	Description    string    `json:"description"`
	Status         string    `json:"status"`
	OwnerDriverID  string    `json:"owner_driver_id"`
	OwnerName      string    `json:"owner_name"`
	OwnerMobile    string    `json:"owner_mobile"`
	MemberCount    int       `json:"member_count"`
	VehicleCount   int       `json:"vehicle_count"`
	PassengerCount int       `json:"passenger_count"`
	ShiftCount     int       `json:"shift_count"`
	CreatedAt      time.Time `json:"created_at"`
}

// GroupSearchDetails is how a fleet looks to somebody who is not in it yet: enough
// to decide whether to ask to join, plus where their own request currently stands.
type GroupSearchDetails struct {
	ID             string    `json:"id"`
	Name           string    `json:"name"`
	Description    string    `json:"description"`
	OwnerDriverID  string    `json:"owner_driver_id"`
	OwnerName      string    `json:"owner_name"`
	VehicleCount   int       `json:"vehicle_count"`
	PassengerCount int       `json:"passenger_count"`
	MyStatus       string    `json:"my_status"`
	CreatedAt      time.Time `json:"created_at"`
}

// GroupPassengerScheduleDetails is one leg of one passenger's standing travel form,
// read by the manager while deciding who to seat on which shift. A passenger shows
// up once per day and direction they asked to travel on.
type GroupPassengerScheduleDetails struct {
	PassengerID     string  `json:"passenger_id"`
	PassengerName   string  `json:"passenger_name"`
	PassengerMobile string  `json:"passenger_mobile"`
	Gender          string  `json:"gender"`
	DayOfWeek       int     `json:"day_of_week"`
	Direction       string  `json:"direction"`
	IsEnabled       bool    `json:"is_enabled"`
	Location        string  `json:"location"`
	Lat             float64 `json:"lat"`
	Lng             float64 `json:"lng"`
	ScheduledTime   string  `json:"scheduled_time"`
}

// GroupPassengerDetails is the joined read projection of a passenger in a group.
type GroupPassengerDetails struct {
	ID              string    `json:"id"`
	GroupID         string    `json:"group_id"`
	PassengerID     string    `json:"passenger_id"`
	PassengerName   string    `json:"passenger_name"`
	PassengerMobile string    `json:"passenger_mobile"`
	Gender          string    `json:"gender"`
	Status          string    `json:"status"`
	CreatedAt       time.Time `json:"created_at"`
}
