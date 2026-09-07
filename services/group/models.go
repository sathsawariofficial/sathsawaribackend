package group

import "time"

// CreateGroupRequest starts a fleet. A group has to be born with at least one
// vehicle, an empty fleet has nothing to build a shift out of.
type CreateGroupRequest struct {
	Name        string   `json:"name"`
	Description string   `json:"description"`
	VehicleIds  []string `json:"vehicleIds"`
}

type CreateGroupResponse struct {
	GroupId string `json:"groupId"`
}

// DriverJoinRequest is a driver asking to be let into a fleet. A driver may come
// with vehicles, may come only to drive somebody else's vehicle, or may only lend
// vehicles without driving them.
type DriverJoinRequest struct {
	GroupId    string   `json:"groupId"`
	JoinType   string   `json:"joinType"`
	VehicleIds []string `json:"vehicleIds"`
}

type PassengerJoinRequest struct {
	GroupId string `json:"groupId"`
}

// MembershipDecision is one line of the bulk decide api. memberId is the id of the
// membership row itself, not of the driver, vehicle or passenger behind it.
type MembershipDecision struct {
	MemberType string `json:"memberType"`
	MemberId   string `json:"memberId"`
	Action     string `json:"action"`
}

type DecideMembershipRequest struct {
	GroupId   string               `json:"groupId"`
	Decisions []MembershipDecision `json:"decisions"`
}

// DecisionResult reports back what actually happened to one decision, the caller
// sends a batch and needs to know which lines did not take and why.
type DecisionResult struct {
	MemberType string `json:"memberType"`
	MemberId   string `json:"memberId"`
	Action     string `json:"action"`
	Applied    bool   `json:"applied"`
	Reason     string `json:"reason,omitempty"`
}

type DecideMembershipResponse struct {
	Applied []DecisionResult `json:"applied"`
	Skipped []DecisionResult `json:"skipped"`
}

// SetSubManagersRequest promotes and demotes sub managers in one call. Only a
// driver who is already an approved member of the group can be promoted.
type SetSubManagersRequest struct {
	GroupId string   `json:"groupId"`
	Promote []string `json:"promote"`
	Demote  []string `json:"demote"`
}

type SubManagerResult struct {
	DriverId string `json:"driverId"`
	Action   string `json:"action"`
	Applied  bool   `json:"applied"`
	Reason   string `json:"reason,omitempty"`
}

type SetSubManagersResponse struct {
	Applied []SubManagerResult `json:"applied"`
	Skipped []SubManagerResult `json:"skipped"`
}

type GroupDetails struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	OwnerId     string    `json:"ownerDriverId"`
	Status      string    `json:"status"`
	IsOwner     bool      `json:"isOwner"`
	CreatedAt   time.Time `json:"createdAt"`
}

type GroupsResponse struct {
	TotalPages int            `json:"totalPages"`
	Groups     []GroupDetails `json:"groups"`
}

type MemberDetails struct {
	ID           string    `json:"id"`
	DriverId     string    `json:"driverId"`
	DriverName   string    `json:"driverName"`
	DriverMobile string    `json:"driverMobile"`
	Rating       string    `json:"rating"`
	RoleName     string    `json:"roleName"`
	JoinType     string    `json:"joinType"`
	Status       string    `json:"status"`
	CreatedAt    time.Time `json:"createdAt"`
}

type VehicleDetails struct {
	ID            string    `json:"id"`
	VehicleId     string    `json:"vehicleId"`
	VehicleNumber string    `json:"vehicleNumber"`
	VehicleInfo   string    `json:"vehicleInfo"`
	NumberOfSeats int       `json:"numberOfSeats"`
	HasAC         bool      `json:"hasAC"`
	HasHeating    bool      `json:"hasHeating"`
	DriverId      string    `json:"driverId"`
	DriverName    string    `json:"driverName"`
	DriverMobile  string    `json:"driverMobile"`
	Status        string    `json:"status"`
	CreatedAt     time.Time `json:"createdAt"`
}

type PassengerDetails struct {
	ID              string    `json:"id"`
	PassengerId     string    `json:"passengerId"`
	PassengerName   string    `json:"passengerName"`
	PassengerMobile string    `json:"passengerMobile"`
	Gender          string    `json:"gender"`
	Status          string    `json:"status"`
	CreatedAt       time.Time `json:"createdAt"`
}

type GroupDetailsResponse struct {
	Group      GroupDetails       `json:"group"`
	Members    []MemberDetails    `json:"members"`
	Vehicles   []VehicleDetails   `json:"vehicles"`
	Passengers []PassengerDetails `json:"passengers"`
}

type PendingRequestsResponse struct {
	Members    []MemberDetails    `json:"members"`
	Vehicles   []VehicleDetails   `json:"vehicles"`
	Passengers []PassengerDetails `json:"passengers"`
}

// PassengerScheduleEntry is one leg of one passenger's standing travel form as the
// manager sees it while building a shift.
type PassengerScheduleEntry struct {
	PassengerId     string  `json:"passengerId"`
	PassengerName   string  `json:"passengerName"`
	PassengerMobile string  `json:"passengerMobile"`
	Gender          string  `json:"gender"`
	DayOfWeek       int     `json:"dayOfWeek"`
	Direction       string  `json:"direction"`
	IsEnabled       bool    `json:"isEnabled"`
	Location        string  `json:"location"`
	Lat             float64 `json:"lat"`
	Lng             float64 `json:"lng"`
	ScheduledTime   string  `json:"scheduledTime"`
}

type PassengerSchedulesResponse struct {
	TotalPages int                      `json:"totalPages"`
	Schedules  []PassengerScheduleEntry `json:"schedules"`
}

// GroupSummary is a fleet as it looks to somebody deciding whether to ask to join.
// MyStatus is empty when they have never asked.
type GroupSummary struct {
	ID             string    `json:"id"`
	Name           string    `json:"name"`
	Description    string    `json:"description"`
	OwnerId        string    `json:"ownerDriverId"`
	OwnerName      string    `json:"ownerName"`
	VehicleCount   int       `json:"vehicleCount"`
	PassengerCount int       `json:"passengerCount"`
	MyStatus       string    `json:"myStatus"`
	CreatedAt      time.Time `json:"createdAt"`
}

type GroupSearchResponse struct {
	TotalPages int            `json:"totalPages"`
	Groups     []GroupSummary `json:"groups"`
}
