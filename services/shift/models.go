package shift

import "time"

// SeatAssignment places one passenger in one numbered seat. Gender is held on the
// seat itself, a seat kept for one gender can never be handed to the other.
type SeatAssignment struct {
	SeatNumber  int    `json:"seatNumber"`
	Gender      string `json:"gender"`
	PassengerId string `json:"passengerId"`
}

// StopRequest is one waypoint of the route, in the order the manager arranged it.
// The seats listed under it are the people picked up or dropped at that stop.
type StopRequest struct {
	Location      string           `json:"location"`
	Lat           float64          `json:"lat"`
	Lng           float64          `json:"lng"`
	ScheduledTime string           `json:"scheduledTime"`
	Seats         []SeatAssignment `json:"seats"`
}

// CreateShiftRequest builds one directional trip of one vehicle on one day. A
// pickup and a drop off are always two separate shifts, even for the same people.
type CreateShiftRequest struct {
	GroupId              string        `json:"groupId"`
	VehicleId            string        `json:"vehicleId"`
	DriverId             string        `json:"driverId"`
	Direction            string        `json:"direction"`
	StartDatetime        string        `json:"startDatetime"`
	EstimatedEndDatetime string        `json:"estimatedEndDatetime"`
	StartLocation        string        `json:"startLocation"`
	EndLocation          string        `json:"endLocation"`
	RouteDetails         string        `json:"routeDetails"`
	Stops                []StopRequest `json:"stops"`
	MakeTemplate         bool          `json:"makeTemplate"`
	TemplateName         string        `json:"templateName"`
	DaysOfWeek           []int         `json:"daysOfWeek"`
}

type CreateShiftResponse struct {
	ShiftId    string `json:"shiftId"`
	TemplateId string `json:"templateId,omitempty"`
}

// SeatUpdate re-seats one place on an existing shift. An empty passengerId frees
// the seat, stopSequence says which stop of the route the passenger waits at.
type SeatUpdate struct {
	SeatNumber   int    `json:"seatNumber"`
	Gender       string `json:"gender"`
	PassengerId  string `json:"passengerId"`
	StopSequence int    `json:"stopSequence"`
}

type UpdateShiftSeatsRequest struct {
	ShiftId string       `json:"shiftId"`
	Seats   []SeatUpdate `json:"seats"`
}

type ShiftStopDetails struct {
	ID             string  `json:"id"`
	SequenceNumber int     `json:"sequenceNumber"`
	Location       string  `json:"location"`
	Lat            float64 `json:"lat"`
	Lng            float64 `json:"lng"`
	ScheduledTime  string  `json:"scheduledTime"`
}

type ShiftSeatDetails struct {
	ID              string `json:"id"`
	SeatNumber      int    `json:"seatNumber"`
	Gender          string `json:"gender"`
	Status          string `json:"status"`
	PassengerId     string `json:"passengerId,omitempty"`
	PassengerName   string `json:"passengerName,omitempty"`
	PassengerMobile string `json:"passengerMobile,omitempty"`
	StopId          string `json:"stopId,omitempty"`
	StopSequence    int    `json:"stopSequence,omitempty"`
	Location        string `json:"location,omitempty"`
	ScheduledTime   string `json:"scheduledTime,omitempty"`
}

// ShiftDetails carries the driver's number and the number of the manager who built
// the shift, so anybody with a problem on the day knows who to call.
type ShiftDetails struct {
	ID                   string    `json:"id"`
	GroupId              string    `json:"groupId"`
	GroupName            string    `json:"groupName,omitempty"`
	VehicleId            string    `json:"vehicleId"`
	VehicleNumber        string    `json:"vehicleNumber"`
	VehicleInfo          string    `json:"vehicleInfo"`
	HasAC                bool      `json:"hasAC"`
	HasHeating           bool      `json:"hasHeating"`
	DriverId             string    `json:"driverId"`
	DriverName           string    `json:"driverName"`
	DriverMobile         string    `json:"driverMobile"`
	Rating               string    `json:"rating,omitempty"`
	Direction            string    `json:"direction"`
	StartDatetime        string    `json:"startDatetime"`
	EstimatedEndDatetime string    `json:"estimatedEndDatetime"`
	StartLocation        string    `json:"startLocation"`
	EndLocation          string    `json:"endLocation"`
	NumberOfSeats        int       `json:"numberOfSeats"`
	SeatsTaken           int       `json:"seatsTaken"`
	RouteDetails         string    `json:"routeDetails,omitempty"`
	CreatedById          string    `json:"createdByDriverId"`
	CreatedByName        string    `json:"createdByName"`
	CreatedByMobile      string    `json:"createdByMobile"`
	IsActive             bool      `json:"isActive"`
	CreatedAt            time.Time `json:"createdAt"`
}

type ShiftDetailsResponse struct {
	Shift ShiftDetails       `json:"shift"`
	Stops []ShiftStopDetails `json:"stops"`
	Seats []ShiftSeatDetails `json:"seats"`
}

type ShiftsResponse struct {
	TotalPages int            `json:"totalPages"`
	Shifts     []ShiftDetails `json:"shifts"`
}

type TemplateSeatDetails struct {
	SeatNumber   int    `json:"seatNumber"`
	Gender       string `json:"gender"`
	PassengerId  string `json:"passengerId,omitempty"`
	StopSequence int    `json:"stopSequence"`
}

// TemplateDetails is what the app pulls down to prefill the create shift form, the
// same way a ride template is reused, there is no separate build from template api.
type TemplateDetails struct {
	ID                   string                `json:"id"`
	GroupId              string                `json:"groupId"`
	Name                 string                `json:"name"`
	VehicleId            string                `json:"vehicleId"`
	VehicleNumber        string                `json:"vehicleNumber"`
	VehicleInfo          string                `json:"vehicleInfo"`
	NumberOfSeats        int                   `json:"numberOfSeats"`
	DriverId             string                `json:"driverId"`
	Direction            string                `json:"direction"`
	StartDatetime        string                `json:"startDatetime"`
	EstimatedEndDatetime string                `json:"estimatedEndDatetime"`
	StartLocation        string                `json:"startLocation"`
	EndLocation          string                `json:"endLocation"`
	RouteDetails         string                `json:"routeDetails,omitempty"`
	DaysOfWeek           []int                 `json:"daysOfWeek"`
	Stops                []ShiftStopDetails    `json:"stops"`
	Seats                []TemplateSeatDetails `json:"seats"`
	CreatedAt            time.Time             `json:"createdAt"`
}

type TemplatesResponse struct {
	Templates []TemplateDetails `json:"templates"`
}

// seatPlan is the flattened seat list worked out from the nested stop payload, it
// pairs a seat with the stop index its passenger waits at.
type seatPlan struct {
	SeatNumber  int
	Gender      string
	PassengerId string
	StopIndex   int
}
