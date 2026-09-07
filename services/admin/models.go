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

// AdminRoleRequest is the body of the admin create role api. A role created by an
// admin is never a system role.
type AdminRoleRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

// AdminRoleUpdateRequest is the body of the admin update role api. Both fields are
// pointers so that a caller can send only the one it wants changed.
type AdminRoleUpdateRequest struct {
	Name        *string `json:"name"`
	Description *string `json:"description"`
}

// AdminPermissionRequest is the body of the admin create permission api.
type AdminPermissionRequest struct {
	Code        string `json:"code"`
	Description string `json:"description"`
}

// AdminRolePermissionsRequest is the body of the admin bulk set permissions api, it
// carries the complete permission set the role should hold after the call.
type AdminRolePermissionsRequest struct {
	RoleId        string   `json:"roleId"`
	PermissionIds []string `json:"permissionIds"`
}

type AdminPermissionDetail struct {
	ID          string    `json:"id"`
	Code        string    `json:"code"`
	Description string    `json:"description"`
	IsSystem    bool      `json:"isSystem"`
	CreatedAt   time.Time `json:"createdAt"`
}

type AdminRoleDetail struct {
	ID          string                  `json:"id"`
	Name        string                  `json:"name"`
	Description string                  `json:"description"`
	IsSystem    bool                    `json:"isSystem"`
	CreatedAt   time.Time               `json:"createdAt"`
	Permissions []AdminPermissionDetail `json:"permissions"`
}

type AdminRoleListResponse struct {
	TotalPages int               `json:"totalPages"`
	Roles      []AdminRoleDetail `json:"roles"`
}

type AdminPermissionListResponse struct {
	TotalPages  int                     `json:"totalPages"`
	Permissions []AdminPermissionDetail `json:"permissions"`
}

type AdminRoleCreatedResponse struct {
	Id string `json:"id"`
}

type AdminPermissionCreatedResponse struct {
	Id string `json:"id"`
}

// adminRolePermissionRow is the scan projection of the single joined query that
// loads the permissions of a whole page of roles at once, the role id is carried
// on every row so the rows can be stitched back onto their roles in memory.
type adminRolePermissionRow struct {
	RoleID      string    `json:"role_id"`
	ID          string    `json:"id"`
	Code        string    `json:"code"`
	Description string    `json:"description"`
	IsSystem    bool      `json:"is_system"`
	CreatedAt   time.Time `json:"created_at"`
}

// AdminOverviewResponse is the platform at a glance, the first screen of an admin
// console before drilling into anything.
type AdminOverviewResponse struct {
	Drivers    AdminCountPair `json:"drivers"`
	Passengers AdminCountPair `json:"passengers"`
	Vehicles   AdminCountPair `json:"vehicles"`
	Groups     AdminCountPair `json:"groups"`
	Shifts     AdminCountPair `json:"shifts"`
	Rides      AdminCountPair `json:"rides"`
	// requests still waiting on a fleet manager, across every group
	PendingGroupRequests int64 `json:"pendingGroupRequests"`
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

type AdminPassengerGroup struct {
	GroupId   string    `json:"groupId"`
	GroupName string    `json:"groupName"`
	Status    string    `json:"status"`
	JoinedAt  time.Time `json:"joinedAt"`
}

type AdminSchedulePreference struct {
	DayOfWeek     int     `json:"dayOfWeek"`
	Direction     string  `json:"direction"`
	IsEnabled     bool    `json:"isEnabled"`
	Location      string  `json:"location"`
	Lat           float64 `json:"lat"`
	Lng           float64 `json:"lng"`
	ScheduledTime string  `json:"scheduledTime"`
}

// AdminPassengerProfileResponse explains what one account is actually doing: who
// they ride with and what they asked for.
type AdminPassengerProfileResponse struct {
	Passenger   AdminPassengerDetail      `json:"passenger"`
	Groups      []AdminPassengerGroup     `json:"groups"`
	TravelForm  []AdminSchedulePreference `json:"travelForm"`
	TotalGroups int                       `json:"totalGroups"`
}

type AdminGroupDetail struct {
	ID             string    `json:"id"`
	Name           string    `json:"name"`
	Description    string    `json:"description"`
	Status         string    `json:"status"`
	OwnerId        string    `json:"ownerDriverId"`
	OwnerName      string    `json:"ownerName"`
	OwnerMobile    string    `json:"ownerMobile"`
	MemberCount    int       `json:"memberCount"`
	VehicleCount   int       `json:"vehicleCount"`
	PassengerCount int       `json:"passengerCount"`
	ShiftCount     int       `json:"shiftCount"`
	CreatedAt      time.Time `json:"createdAt"`
}

type AdminGroupListResponse struct {
	TotalPages int                `json:"totalPages"`
	Groups     []AdminGroupDetail `json:"groups"`
}

type AdminGroupMember struct {
	ID           string    `json:"id"`
	DriverId     string    `json:"driverId"`
	DriverName   string    `json:"driverName"`
	DriverMobile string    `json:"driverMobile"`
	RoleName     string    `json:"roleName"`
	JoinType     string    `json:"joinType"`
	Status       string    `json:"status"`
	CreatedAt    time.Time `json:"createdAt"`
}

type AdminGroupVehicle struct {
	ID            string    `json:"id"`
	VehicleId     string    `json:"vehicleId"`
	VehicleNumber string    `json:"vehicleNumber"`
	VehicleInfo   string    `json:"vehicleInfo"`
	NumberOfSeats int       `json:"numberOfSeats"`
	HasAC         bool      `json:"hasAC"`
	HasHeating    bool      `json:"hasHeating"`
	DriverName    string    `json:"driverName"`
	Status        string    `json:"status"`
	CreatedAt     time.Time `json:"createdAt"`
}

type AdminGroupPassenger struct {
	ID              string    `json:"id"`
	PassengerId     string    `json:"passengerId"`
	PassengerName   string    `json:"passengerName"`
	PassengerMobile string    `json:"passengerMobile"`
	Gender          string    `json:"gender"`
	Status          string    `json:"status"`
	CreatedAt       time.Time `json:"createdAt"`
}

type AdminGroupDetailsResponse struct {
	Group      AdminGroupDetail      `json:"group"`
	Members    []AdminGroupMember    `json:"members"`
	Vehicles   []AdminGroupVehicle   `json:"vehicles"`
	Passengers []AdminGroupPassenger `json:"passengers"`
}

type AdminShiftDetail struct {
	ID                   string    `json:"id"`
	GroupId              string    `json:"groupId"`
	GroupName            string    `json:"groupName"`
	VehicleNumber        string    `json:"vehicleNumber"`
	VehicleInfo          string    `json:"vehicleInfo"`
	DriverId             string    `json:"driverId"`
	DriverName           string    `json:"driverName"`
	DriverMobile         string    `json:"driverMobile"`
	Direction            string    `json:"direction"`
	StartDatetime        string    `json:"startDatetime"`
	EstimatedEndDatetime string    `json:"estimatedEndDatetime"`
	StartLocation        string    `json:"startLocation"`
	EndLocation          string    `json:"endLocation"`
	NumberOfSeats        int       `json:"numberOfSeats"`
	SeatsTaken           int       `json:"seatsTaken"`
	CreatedByName        string    `json:"createdByName"`
	CreatedByMobile      string    `json:"createdByMobile"`
	IsActive             bool      `json:"isActive"`
	CreatedAt            time.Time `json:"createdAt"`
}

type AdminShiftListResponse struct {
	TotalPages int                `json:"totalPages"`
	Shifts     []AdminShiftDetail `json:"shifts"`
}

type AdminShiftStop struct {
	SequenceNumber int     `json:"sequenceNumber"`
	Location       string  `json:"location"`
	Lat            float64 `json:"lat"`
	Lng            float64 `json:"lng"`
	ScheduledTime  string  `json:"scheduledTime"`
}

type AdminShiftSeat struct {
	SeatNumber      int    `json:"seatNumber"`
	Gender          string `json:"gender"`
	Status          string `json:"status"`
	PassengerName   string `json:"passengerName,omitempty"`
	PassengerMobile string `json:"passengerMobile,omitempty"`
	StopSequence    int    `json:"stopSequence,omitempty"`
	Location        string `json:"location,omitempty"`
	ScheduledTime   string `json:"scheduledTime,omitempty"`
}

type AdminShiftDetailsResponse struct {
	Shift AdminShiftDetail `json:"shift"`
	Stops []AdminShiftStop `json:"stops"`
	Seats []AdminShiftSeat `json:"seats"`
}
