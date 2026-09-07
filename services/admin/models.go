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
