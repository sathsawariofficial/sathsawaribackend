package ride

// recurring ride period
const (
	DAILY   = 1
	WEEKLY  = 2
	MONTHLY = 3
)

const (
	MONDAY    = 1
	TUESDAY   = 2
	WEDNESDAY = 3
	THRUSDAY  = 4
	FRIDAY    = 5
	SATURDAY  = 6
	SUNDAY    = 7
)

// ride-series cancellation skip reason codes
const (
	Ride_Cancel_Skip_Reason_Booked   = "SEATS_BOOKED"
	Ride_Cancel_Skip_Reason_Imminent = "STARTS_TOO_SOON"
	Ride_Cancel_Skip_Reason_Invalid  = "INVALID_START_DATE"
)

// default human-readable convenience messages per skip reason
const (
	Ride_Cancel_Skip_Message_Booked   = "This ride already has %d seat(s) booked and cannot be cancelled"
	Ride_Cancel_Skip_Message_Imminent = "This ride starts in less than %d hour(s) and cannot be cancelled"
	Ride_Cancel_Skip_Message_Invalid  = "Unable to verify this ride's start time"
)
