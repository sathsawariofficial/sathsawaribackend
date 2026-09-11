package shift

import (
	"rideshare/pkgs/database/postgress"
	"rideshare/pkgs/utils"
	"time"
)

// ShiftPassengerInput places one passenger on a shift, waiting at the stop of the
// route with this sequence number, counting from 1.
type ShiftPassengerInput struct {
	PassengerId      string `json:"passengerId"`
	LocationSequence int    `json:"locationSequence"`
}

// CreateShiftRequest builds one recurring shift. It runs on its days of the week
// from the start date, until the end date when one is given, from the time of its
// first location to the time of its last. It is one shift however many days it runs,
// and it is named after its vehicle's number.
type CreateShiftRequest struct {
	DriverId   string                `json:"driverId"`
	VehicleId  string                `json:"vehicleId"`
	DaysOfWeek []int                 `json:"daysOfWeek"`
	StartDate  string                `json:"startDate"`
	EndDate    string                `json:"endDate"`
	Locations  []utils.RouteLocation `json:"locations"`
	Passengers []ShiftPassengerInput `json:"passengers"`
}

type CreateShiftResponse struct {
	ShiftId string   `json:"shiftId"`
	Seats   SeatInfo `json:"seats"`
}

// UpdateShiftRequest edits a shift. Every field but the shift is optional and a
// field left out stays as it is. An empty endDate makes the shift open ended again,
// and locations replace the whole route. A new vehicle renames the shift after it.
type UpdateShiftRequest struct {
	ShiftId    string                `json:"shiftId"`
	DriverId   *string               `json:"driverId"`
	VehicleId  *string               `json:"vehicleId"`
	DaysOfWeek []int                 `json:"daysOfWeek"`
	StartDate  *string               `json:"startDate"`
	EndDate    *string               `json:"endDate"`
	Locations  []utils.RouteLocation `json:"locations"`
}

// UpdateShiftPassengersRequest adds, moves and removes passengers in one call, all
// judged against the seats the shift will have once every change lands.
type UpdateShiftPassengersRequest struct {
	ShiftId string                `json:"shiftId"`
	Add     []ShiftPassengerInput `json:"add"`
	Move    []ShiftPassengerInput `json:"move"`
	Remove  []string              `json:"remove"`
}

type UpdateShiftPassengersResponse struct {
	Added   []string `json:"added"`
	Moved   []string `json:"moved"`
	Removed []string `json:"removed"`
	Seats   SeatInfo `json:"seats"`
}

type SeatInfo struct {
	TotalSeats     int `json:"totalSeats"`
	OccupiedSeats  int `json:"occupiedSeats"`
	RemainingSeats int `json:"remainingSeats"`
}

// ShiftLocationDetail is one stop of a route. The same shape is frozen into every
// occurrence, which is why travel history can still show a route that later changed.
type ShiftLocationDetail struct {
	Sequence int     `json:"sequence"`
	Location string  `json:"location"`
	Lat      float64 `json:"lat"`
	Lng      float64 `json:"lng"`
	Time     string  `json:"time"`
}

type ShiftPassengerDetail struct {
	PassengerId      string     `json:"passengerId"`
	PassengerName    string     `json:"passengerName"`
	PassengerMobile  string     `json:"passengerMobile,omitempty"`
	LocationSequence int        `json:"locationSequence"`
	Location         string     `json:"location"`
	Time             string     `json:"time"`
	AddedAt          *time.Time `json:"addedAt,omitempty"`
}

type ShiftSummary struct {
	ID            string `json:"id"`
	Name          string `json:"name"`
	ServiceId     string `json:"serviceId"`
	ServiceName   string `json:"serviceName"`
	DriverId      string `json:"driverId"`
	DriverName    string `json:"driverName"`
	DriverMobile  string `json:"driverMobile"`
	VehicleId     string `json:"vehicleId"`
	VehicleNumber string `json:"vehicleNumber"`
	VehicleInfo   string `json:"vehicleInfo"`
	// the driver who brought the vehicle, a member who joined with vehicles only is
	// never the driver of their own vehicle's shift
	VehicleOwnerId   string    `json:"vehicleOwnerId"`
	VehicleOwnerName string    `json:"vehicleOwnerName"`
	DaysOfWeek       []int     `json:"daysOfWeek"`
	StartDate        string    `json:"startDate"`
	EndDate          string    `json:"endDate"`
	StartTime        string    `json:"startTime"`
	EndTime          string    `json:"endTime"`
	Seats            SeatInfo  `json:"seats"`
	Status           string    `json:"status"`
	NextOccurrence   string    `json:"nextOccurrence"`
	CreatedAt        time.Time `json:"createdAt"`
}

// ShiftDetailResponse is a shift with its route and passengers. A passenger reading
// it gets myStop and nobody else's mobile number.
type ShiftDetailResponse struct {
	Shift      ShiftSummary           `json:"shift"`
	Route      []ShiftLocationDetail  `json:"route"`
	Passengers []ShiftPassengerDetail `json:"passengers"`
	MyStop     *ShiftPassengerDetail  `json:"myStop,omitempty"`
}

type ShiftsResponse struct {
	TotalPages int            `json:"totalPages"`
	Shifts     []ShiftSummary `json:"shifts"`
}

type ShiftHistoryItem struct {
	ID            string     `json:"id"`
	Name          string     `json:"name"`
	DriverName    string     `json:"driverName"`
	VehicleNumber string     `json:"vehicleNumber"`
	DaysOfWeek    []int      `json:"daysOfWeek"`
	StartDate     string     `json:"startDate"`
	EndDate       string     `json:"endDate"`
	StartTime     string     `json:"startTime"`
	EndTime       string     `json:"endTime"`
	Status        string     `json:"status"`
	CreatedAt     time.Time  `json:"createdAt"`
	DeletedAt     *time.Time `json:"deletedAt"`
}

type ShiftHistoryResponse struct {
	TotalPages int                `json:"totalPages"`
	Shifts     []ShiftHistoryItem `json:"shifts"`
}

// PassengerHistoryItem is one stint of one passenger on one shift, read from rows that
// are never deleted, so it outlives the passenger leaving and the shift being deleted.
type PassengerHistoryItem struct {
	ShiftPassengerId string     `json:"shiftPassengerId"`
	ShiftId          string     `json:"shiftId"`
	ShiftName        string     `json:"shiftName"`
	ShiftStatus      string     `json:"shiftStatus"`
	PassengerId      string     `json:"passengerId"`
	PassengerName    string     `json:"passengerName"`
	PassengerMobile  string     `json:"passengerMobile"`
	LocationSequence int        `json:"locationSequence"`
	Status           string     `json:"status"`
	RemovalReason    string     `json:"removalReason,omitempty"`
	JoinedServiceAt  *time.Time `json:"joinedServiceAt"`
	AddedAt          time.Time  `json:"addedAt"`
	RemovedAt        *time.Time `json:"removedAt"`
	TripsPresent     int64      `json:"tripsPresent"`
	TripsAbsent      int64      `json:"tripsAbsent"`
}

type PassengerHistoryResponse struct {
	TotalPages int                    `json:"totalPages"`
	History    []PassengerHistoryItem `json:"history"`
}

type OccurrenceItem struct {
	OccurrenceId  string `json:"occurrenceId"`
	ShiftId       string `json:"shiftId"`
	ShiftName     string `json:"shiftName"`
	ServiceName   string `json:"serviceName"`
	Date          string `json:"date"`
	StartTime     string `json:"startTime"`
	EndTime       string `json:"endTime"`
	DriverName    string `json:"driverName"`
	VehicleNumber string `json:"vehicleNumber"`
	Status        string `json:"status"`
	PresentCount  int64  `json:"presentCount"`
	AbsentCount   int64  `json:"absentCount"`
}

type OccurrencesResponse struct {
	TotalPages  int              `json:"totalPages"`
	Occurrences []OccurrenceItem `json:"occurrences"`
}

type AttendanceEntry struct {
	PassengerId      string     `json:"passengerId"`
	PassengerName    string     `json:"passengerName"`
	PassengerMobile  string     `json:"passengerMobile,omitempty"`
	LocationSequence int        `json:"locationSequence"`
	Location         string     `json:"location"`
	Time             string     `json:"time"`
	Status           string     `json:"status"`
	MarkedAt         *time.Time `json:"markedAt"`
	IsMe             bool       `json:"isMe,omitempty"`
}

// AttendanceResponse is who travels on one trip. A trip not yet written ahead by the
// scheduler is shown from the shift as it stands, everybody present.
type AttendanceResponse struct {
	ShiftId    string            `json:"shiftId"`
	ShiftName  string            `json:"shiftName"`
	Date       string            `json:"date"`
	StartTime  string            `json:"startTime"`
	EndTime    string            `json:"endTime"`
	Status     string            `json:"status"`
	Attendance []AttendanceEntry `json:"attendance"`
}

// MarkAttendanceRequest marks the calling passenger absent, or present again, for one
// date the shift runs on.
type MarkAttendanceRequest struct {
	ShiftId string `json:"shiftId"`
	Date    string `json:"date"`
	Status  string `json:"status"`
}

type MarkAttendanceResponse struct {
	ShiftId string `json:"shiftId"`
	Date    string `json:"date"`
	Status  string `json:"status"`
	Changed bool   `json:"changed"`
}

// DriverLocationRequest is what a driver tells the passengers of today's trip, a
// message, a place, or both.
type DriverLocationRequest struct {
	ShiftId  string  `json:"shiftId"`
	Message  string  `json:"message"`
	Location string  `json:"location"`
	Lat      float64 `json:"lat"`
	Lng      float64 `json:"lng"`
}

type DriverLocationResponse struct {
	UpdateId   string `json:"updateId"`
	Date       string `json:"date"`
	Recipients int    `json:"recipients"`
}

type TravelHistoryItem struct {
	OccurrenceId  string                `json:"occurrenceId"`
	ShiftId       string                `json:"shiftId"`
	ShiftName     string                `json:"shiftName"`
	ServiceName   string                `json:"serviceName"`
	Date          string                `json:"date"`
	StartTime     string                `json:"startTime"`
	EndTime       string                `json:"endTime"`
	DriverName    string                `json:"driverName"`
	DriverMobile  string                `json:"driverMobile"`
	VehicleNumber string                `json:"vehicleNumber"`
	Route         []ShiftLocationDetail `json:"route"`
	MyStop        ShiftPassengerDetail  `json:"myStop"`
	Attendance    string                `json:"attendance"`
}

type TravelHistoryResponse struct {
	TotalPages int                 `json:"totalPages"`
	Trips      []TravelHistoryItem `json:"trips"`
}

// ShiftRequestInput is a passenger's recurring requirement put out for owners to find.
// serviceId is optional, a request addressed to one service is shown only to its owner.
type ShiftRequestInput struct {
	ServiceId     string                `json:"serviceId"`
	DaysOfWeek    []int                 `json:"daysOfWeek"`
	ContactNumber string                `json:"contactNumber"`
	Note          string                `json:"note"`
	Locations     []utils.RouteLocation `json:"locations"`
}

type ShiftRequestResponse struct {
	RequestId string `json:"requestId"`
}

type ShiftRequestItem struct {
	ID            string                `json:"id"`
	ServiceId     string                `json:"serviceId"`
	ServiceName   string                `json:"serviceName"`
	PassengerName string                `json:"passengerName"`
	ContactNumber string                `json:"contactNumber"`
	Note          string                `json:"note"`
	DaysOfWeek    []int                 `json:"daysOfWeek"`
	StartTime     string                `json:"startTime"`
	EndTime       string                `json:"endTime"`
	Locations     []utils.RouteLocation `json:"locations"`
	CreatedAt     time.Time             `json:"createdAt"`
}

type ShiftRequestsResponse struct {
	TotalPages int                `json:"totalPages"`
	Requests   []ShiftRequestItem `json:"requests"`
}

// notice is a notification worked out inside a transaction, sent only once it commits.
type notice struct {
	UserId   string
	UserType int
	Type     string
	Title    string
	Message  string
}

// shiftContext is a shift together with everything its notifications need.
type shiftContext struct {
	Shift      postgress.Shift
	Service    postgress.PickDropService
	Driver     postgress.Driver
	Vehicle    postgress.Vehicle
	Locations  []postgress.ShiftLocation
	Passengers []postgress.ShiftPassenger
	Names      map[string]string
}

type passengerChange struct {
	shiftContext
	Added   []postgress.ShiftPassenger
	Moved   []postgress.ShiftPassenger
	Removed []postgress.ShiftPassenger
}

type attendanceChange struct {
	Changed       bool
	Occurrence    postgress.ShiftOccurrence
	Attendance    postgress.ShiftAttendance
	OwnerId       string
	PassengerName string
}

type driverUpdateResult struct {
	Update     postgress.ShiftDriverUpdate
	Occurrence postgress.ShiftOccurrence
	Recipients []string
}

// shiftListFilter narrows a shift list the way the ride search does: search matches the
// vehicle number or any place on the route, start_time keeps trips that start at or
// after it and end_time trips that are over by then.
type shiftListFilter struct {
	Status    string
	DayOfWeek int
	Search    string
	StartTime string
	EndTime   string
}

type shiftRequestSearch struct {
	Search    string
	DayOfWeek int
	StartTime string
	EndTime   string
}
