package pickdrop

import (
	"rideshare/pkgs/database/postgress"
	"rideshare/pkgs/utils"
	"time"
)

// EnableServiceRequest turns a driver into the owner of a Pick & Drop service. The
// vehicles are optional, a service can start with none and add them later.
type EnableServiceRequest struct {
	Name        string   `json:"name"`
	Description string   `json:"description"`
	VehicleIds  []string `json:"vehicleIds"`
}

type EnableServiceResponse struct {
	ServiceId string `json:"serviceId"`
}

// VehicleIdsRequest carries the vehicles an owner adds or a member driver offers.
type VehicleIdsRequest struct {
	VehicleIds []string `json:"vehicleIds"`
}

// DriverJoinRequest is a driver asking to join a service. joinType driver, the default,
// joins as somebody who can drive the service's shifts, optionally proposing vehicles.
// joinType vehicles joins only through the vehicles proposed, at least one: the driver
// is never put behind the wheel of a shift, their vehicles are driven by others.
// Nothing is granted until the owner decides.
type DriverJoinRequest struct {
	ServiceId  string   `json:"serviceId"`
	JoinType   string   `json:"joinType"`
	VehicleIds []string `json:"vehicleIds"`
}

type PassengerJoinRequest struct {
	ServiceId string `json:"serviceId"`
}

type JoinRequestResponse struct {
	RequestId string `json:"requestId"`
}

// RequestDecision is one line of the bulk decide api. requestId is the id of the
// request row itself, not of the driver, vehicle or passenger behind it.
type RequestDecision struct {
	Type      string `json:"type"`
	RequestId string `json:"requestId"`
	Action    string `json:"action"`
}

type DecideRequestsRequest struct {
	Decisions []RequestDecision `json:"decisions"`
}

// DecisionResult says what happened to one decision, and why when it did not take.
type DecisionResult struct {
	Type      string `json:"type"`
	RequestId string `json:"requestId"`
	Action    string `json:"action"`
	Applied   bool   `json:"applied"`
	Reason    string `json:"reason,omitempty"`
}

type DecideRequestsResponse struct {
	Applied []DecisionResult `json:"applied"`
	Skipped []DecisionResult `json:"skipped"`
}

type ServiceSummary struct {
	ID            string    `json:"id"`
	Name          string    `json:"name"`
	Description   string    `json:"description"`
	OwnerDriverId string    `json:"ownerDriverId"`
	OwnerName     string    `json:"ownerName"`
	OwnerMobile   string    `json:"ownerMobile"`
	Status        string    `json:"status"`
	CreatedAt     time.Time `json:"createdAt"`
}

// ServiceCounts is the owner's dashboard of their service. The advertisement limit
// is worked out from the approved vehicles every time, it is never stored.
type ServiceCounts struct {
	ApprovedDrivers int64 `json:"approvedDrivers"`
	// members who joined with vehicles only, they are not counted as drivers
	ApprovedVehicleOwners int64 `json:"approvedVehicleOwners"`
	ApprovedVehicles      int64 `json:"approvedVehicles"`
	ApprovedPassengers    int64 `json:"approvedPassengers"`
	PendingRequests       int64 `json:"pendingRequests"`
	ActiveShifts          int64 `json:"activeShifts"`
	Advertisements        int64 `json:"advertisements"`
	AdvertisementLimit    int64 `json:"advertisementLimit"`
}

type MembershipInfo struct {
	RequestId   string     `json:"requestId"`
	JoinType    string     `json:"joinType,omitempty"`
	Status      string     `json:"status"`
	RequestedAt time.Time  `json:"requestedAt"`
	DecidedAt   *time.Time `json:"decidedAt"`
}

type MemberVehicle struct {
	RequestId     string     `json:"requestId"`
	VehicleId     string     `json:"vehicleId"`
	VehicleNumber string     `json:"vehicleNumber"`
	VehicleInfo   string     `json:"vehicleInfo"`
	NumberOfSeats int        `json:"numberOfSeats"`
	HasValidSeats bool       `json:"hasValidSeats"`
	Status        string     `json:"status"`
	RequestedAt   time.Time  `json:"requestedAt"`
	DecidedAt     *time.Time `json:"decidedAt"`
}

// MyServiceResponse is what Pick & Drop looks like to the calling driver: the service
// they own, the one they belong to or asked to join, or nothing at all.
type MyServiceResponse struct {
	Role       string          `json:"role"`
	Service    *ServiceSummary `json:"service"`
	Counts     *ServiceCounts  `json:"counts,omitempty"`
	Membership *MembershipInfo `json:"membership,omitempty"`
	Vehicles   []MemberVehicle `json:"vehicles"`
}

type PassengerMembershipResponse struct {
	Role       string          `json:"role"`
	Service    *ServiceSummary `json:"service"`
	Membership *MembershipInfo `json:"membership,omitempty"`
}

// ServiceSearchItem is a service as it looks to somebody deciding whether to ask to
// join. MyStatus is owner, pending, approved, or empty when they have never asked.
type ServiceSearchItem struct {
	ID             string    `json:"id"`
	Name           string    `json:"name"`
	Description    string    `json:"description"`
	OwnerName      string    `json:"ownerName"`
	DriverCount    int64     `json:"driverCount"`
	VehicleCount   int64     `json:"vehicleCount"`
	PassengerCount int64     `json:"passengerCount"`
	MyStatus       string    `json:"myStatus"`
	CreatedAt      time.Time `json:"createdAt"`
}

type ServiceSearchResponse struct {
	TotalPages int                 `json:"totalPages"`
	Services   []ServiceSearchItem `json:"services"`
}

type DriverRequestItem struct {
	RequestId    string          `json:"requestId"`
	DriverId     string          `json:"driverId"`
	DriverName   string          `json:"driverName"`
	DriverMobile string          `json:"driverMobile"`
	Rating       string          `json:"rating"`
	JoinType     string          `json:"joinType"`
	Status       string          `json:"status"`
	RequestedAt  time.Time       `json:"requestedAt"`
	DecidedAt    *time.Time      `json:"decidedAt"`
	Vehicles     []MemberVehicle `json:"vehicles"`
}

type VehicleRequestItem struct {
	RequestId     string     `json:"requestId"`
	VehicleId     string     `json:"vehicleId"`
	VehicleNumber string     `json:"vehicleNumber"`
	VehicleInfo   string     `json:"vehicleInfo"`
	NumberOfSeats int        `json:"numberOfSeats"`
	HasValidSeats bool       `json:"hasValidSeats"`
	HasAC         bool       `json:"hasAC"`
	HasHeating    bool       `json:"hasHeating"`
	DriverId      string     `json:"driverId"`
	DriverName    string     `json:"driverName"`
	DriverMobile  string     `json:"driverMobile"`
	OwnerJoinType string     `json:"ownerJoinType"`
	Status        string     `json:"status"`
	RequestedAt   time.Time  `json:"requestedAt"`
	DecidedAt     *time.Time `json:"decidedAt"`
}

type PassengerRequestItem struct {
	RequestId       string            `json:"requestId"`
	PassengerId     string            `json:"passengerId"`
	PassengerName   string            `json:"passengerName"`
	PassengerMobile string            `json:"passengerMobile"`
	Gender          string            `json:"gender"`
	Status          string            `json:"status"`
	RequestedAt     time.Time         `json:"requestedAt"`
	DecidedAt       *time.Time        `json:"decidedAt"`
	WeeklyDemand    []AvailabilityDay `json:"weeklyDemand,omitempty"`
}

// JoinRequestsResponse lists one kind of request, requests holds driver, vehicle or
// passenger items according to type.
type JoinRequestsResponse struct {
	TotalPages int    `json:"totalPages"`
	Type       string `json:"type"`
	Requests   any    `json:"requests"`
}

type JoinRequestDetailResponse struct {
	Type    string `json:"type"`
	Request any    `json:"request"`
}

// AvailabilityDay is one day of a passenger's weekly requirement.
type AvailabilityDay struct {
	DayOfWeek  int                   `json:"dayOfWeek"`
	IsRequired bool                  `json:"isRequired"`
	Locations  []utils.RouteLocation `json:"locations"`
}

// AvailabilityFilter narrows the available lists down to who is free for the shift
// the owner is about to build. Without it every approved member is listed and
// isAvailable is left out.
type AvailabilityFilter struct {
	DaysOfWeek     []int
	StartDate      string
	EndDate        string
	StartTime      string
	EndTime        string
	ExcludeShiftId string
}

func (filter AvailabilityFilter) Enabled() bool {
	return len(filter.DaysOfWeek) > 0
}

type AvailableDriverItem struct {
	DriverId     string     `json:"driverId"`
	DriverName   string     `json:"driverName"`
	DriverMobile string     `json:"driverMobile"`
	Rating       string     `json:"rating"`
	IsOwner      bool       `json:"isOwner"`
	JoinedAt     *time.Time `json:"joinedAt"`
	ActiveShifts int64      `json:"activeShifts"`
	IsAvailable  *bool      `json:"isAvailable,omitempty"`
	Conflict     string     `json:"conflict,omitempty"`
}

type AvailableDriversResponse struct {
	TotalPages int                   `json:"totalPages"`
	Drivers    []AvailableDriverItem `json:"drivers"`
}

type AvailableVehicleItem struct {
	VehicleId     string `json:"vehicleId"`
	VehicleNumber string `json:"vehicleNumber"`
	VehicleInfo   string `json:"vehicleInfo"`
	NumberOfSeats int    `json:"numberOfSeats"`
	HasValidSeats bool   `json:"hasValidSeats"`
	HasAC         bool   `json:"hasAC"`
	HasHeating    bool   `json:"hasHeating"`
	DriverId      string `json:"driverId"`
	DriverName    string `json:"driverName"`
	// owner, driver or vehicles: a vehicle whose owner joined with vehicles only needs
	// somebody else behind the wheel
	OwnerJoinType string     `json:"ownerJoinType"`
	JoinedAt      *time.Time `json:"joinedAt"`
	ActiveShifts  int64      `json:"activeShifts"`
	IsAvailable   *bool      `json:"isAvailable,omitempty"`
	Conflict      string     `json:"conflict,omitempty"`
}

type AvailableVehiclesResponse struct {
	TotalPages int                    `json:"totalPages"`
	Vehicles   []AvailableVehicleItem `json:"vehicles"`
}

type AvailablePassengerItem struct {
	PassengerId     string            `json:"passengerId"`
	PassengerName   string            `json:"passengerName"`
	PassengerMobile string            `json:"passengerMobile"`
	Gender          string            `json:"gender"`
	JoinedAt        *time.Time        `json:"joinedAt"`
	ActiveShifts    int64             `json:"activeShifts"`
	WeeklyDemand    []AvailabilityDay `json:"weeklyDemand"`
	IsAvailable     *bool             `json:"isAvailable,omitempty"`
	Conflict        string            `json:"conflict,omitempty"`
}

type AvailablePassengersResponse struct {
	TotalPages int                      `json:"totalPages"`
	Passengers []AvailablePassengerItem `json:"passengers"`
}

// AdvertisementRequest is a route an owner puts out for passengers to find. Every
// location has to be reached between the start and the end time.
type AdvertisementRequest struct {
	Title       string                `json:"title"`
	Description string                `json:"description"`
	Fare        float64               `json:"fare"`
	DaysOfWeek  []int                 `json:"daysOfWeek"`
	StartTime   string                `json:"startTime"`
	EndTime     string                `json:"endTime"`
	Locations   []utils.RouteLocation `json:"locations"`
}

type AdvertisementResponse struct {
	AdvertisementId string `json:"advertisementId"`
}

// AdvertisementSearch is what a passenger narrows the advertisements down by. The
// time window works like the ride search: the advertised route has to fit inside it.
type AdvertisementSearch struct {
	Search    string
	DayOfWeek int
	StartTime string
	EndTime   string
}

type AdvertisementItem struct {
	ID           string                `json:"id"`
	ServiceId    string                `json:"serviceId"`
	ServiceName  string                `json:"serviceName"`
	OwnerName    string                `json:"ownerName"`
	OwnerMobile  string                `json:"ownerMobile"`
	Title        string                `json:"title"`
	Description  string                `json:"description"`
	Fare         float64               `json:"fare"`
	DaysOfWeek   []int                 `json:"daysOfWeek"`
	StartTime    string                `json:"startTime"`
	EndTime      string                `json:"endTime"`
	VehicleCount int64                 `json:"vehicleCount"`
	Locations    []utils.RouteLocation `json:"locations"`
	CreatedAt    time.Time             `json:"createdAt"`
}

type MyAdvertisementsResponse struct {
	TotalPages         int                 `json:"totalPages"`
	AdvertisementCount int64               `json:"advertisementCount"`
	AdvertisementLimit int64               `json:"advertisementLimit"`
	Advertisements     []AdvertisementItem `json:"advertisements"`
}

type AdvertisementSearchResponse struct {
	TotalPages     int                 `json:"totalPages"`
	Advertisements []AdvertisementItem `json:"advertisements"`
}

// notice is a notification worked out inside a transaction, sent only once that
// transaction has committed.
type notice struct {
	UserId   string
	UserType int
	Type     string
	Title    string
	Message  string
}

// serviceCountsRow is the scan target of the single counting query behind an
// owner's view of their service.
type serviceCountsRow struct {
	ApprovedDrivers       int64
	ApprovedVehicleOwners int64
	ApprovedVehicles      int64
	ApprovedPassengers    int64
	PendingRequests       int64
	ActiveShifts          int64
	Advertisements        int64
}

type myServiceData struct {
	Role       string
	Service    *postgress.PickDropServiceDetails
	Counts     *serviceCountsRow
	Membership *postgress.PickDropDriver
	Vehicles   []postgress.PickDropVehicleDetails
}

type passengerMembershipData struct {
	Role       string
	Service    *postgress.PickDropServiceDetails
	Membership *postgress.PickDropPassenger
}
