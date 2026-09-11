package database

import (
	"fmt"
	"rideshare/pkgs/constants"
	"time"

	"github.com/go-redis/redis"
	"gorm.io/gorm"
)

type DatabaseConnections struct {
	Postgres  *gorm.DB
	RedisConn *redis.Client
}

// ShiftWindow is the slice of the calendar a shift occupies: the days of the week it
// runs on, between a start date and an optional end date, from the time of its first
// location to the time of its last. Dates are YYYY-MM-DD and times HH:MM, both in
// the business time zone. An empty EndDate means the shift runs until changed.
type ShiftWindow struct {
	DaysOfWeek []int
	StartDate  string
	EndDate    string
	StartTime  string
	EndTime    string
}

// the three things a shift can clash over
const (
	Clash_Kind_Driver    = "driver"
	Clash_Kind_Vehicle   = "vehicle"
	Clash_Kind_Passenger = "passenger"
)

// ShiftClash describes the first thing standing in the way of an assignment: which
// driver, vehicle or passenger, and what they are already committed to on which
// date. It reads as the error message the caller gets back.
type ShiftClash struct {
	Kind      string
	SubjectID string
	Subject   string
	ShiftName string
	Date      string
	StartTime string
	EndTime   string
	IsRide    bool
}

func (clash ShiftClash) Message() string {
	if clash.IsRide {
		return fmt.Sprintf(constants.Resource_Ride_Clash, clash.Subject, clash.StartTime, clash.EndTime)
	}

	name := clash.ShiftName
	if name == "" {
		name = "unnamed shift"
	}

	return fmt.Sprintf(constants.Resource_Shift_Clash, clash.Subject, name, clash.Date, clash.StartTime, clash.EndTime)
}

// RideMessage reads the clash from the side of a ride being put on the calendar.
func (clash ShiftClash) RideMessage() string {
	if clash.IsRide {
		return fmt.Sprintf(constants.Ride_Clash, clash.Subject, clash.StartTime, clash.EndTime)
	}

	return clash.Message()
}

// RideSlot is the time one ride takes on the calendar. A recurring request lays down
// one slot for the ride itself and one for every repeat.
type RideSlot struct {
	Start time.Time
	End   time.Time
}
