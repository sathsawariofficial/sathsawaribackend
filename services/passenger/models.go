package passenger

import "rideshare/pkgs/utils"

type BookSeatRequest struct {
	RideId       string `json:"rideId"`
	Name         string `json:"name"`
	MobileNumber string `json:"mobileNumber"`
	Seats        int    `json:"seats"`
	Code         string `json:"code"`
}

type BookSeatResponse struct {
	BookingId string `json:"bookingId"`
}

type RideRequest struct {
	StartDatetime        string `json:"startDatetime"`
	EstimatedEndDatetime string `json:"estimatedEndDatetime"`
	NumberOfSeats        int    `json:"numberOfSeats"`
	StartLocation        string `json:"startLocation"`
	EndLocation          string `json:"endLocation"`
	RouteDetails         string `json:"routeDetails"`
	ContactNumber        string `json:"contactNumber"`
}

type RideRequestResponse struct {
	OpenURL   string `json:"openUrl"`
	RequestId string `json:"requestId"`
}

type GetRideRequest struct {
	StartDatetime        string `json:"startDatetime"`
	EstimatedEndDatetime string `json:"estimatedEndDatetime"`
	StartLocation        string `json:"startLocation"`
	EndLocation          string `json:"endLocation"`
}

type RideRequestDetails struct {
	ID                   string `json:"id"`
	ContactNumber        string `json:"contactNumber"`
	StartDatetime        string `json:"startDatetime"`
	EstimatedEndDatetime string `json:"estimatedEndDatetime"`
	NumberOfSeats        int    `json:"numberOfSeats"`
	StartLocation        string `json:"startLocation"`
	EndLocation          string `json:"endLocation"`
	RouteDetails         string `json:"routeDetails"`
	OpenURL              string `json:"openUrl"`
}

type RidesRequestsDetailsResponse struct {
	TotalPages int                  `json:"totalPages"`
	Rides      []RideRequestDetails `json:"rides"`
}

type PassengerRegistrationRequest struct {
	DeviceId     string       `json:"deviceId"`
	MobileNumber string       `json:"mobile"`
	Name         string       `json:"name"`
	Gender       utils.Gender `json:"gender"`
	Password     string       `json:"password"`
	EXTPassword  string       `json:"-"`
}

type PassengerRegistrationResponse struct {
	PassengerId string `json:"passengerId"`
	OTP         string `json:"tempOTP"`
}

type PassengerLoginRequest struct {
	DeviceId     string `json:"deviceId"`
	MobileNumber string `json:"mobile"`
	Password     string `json:"password"`
	FCM          string `json:"fcm"`
}

type PassengerLoginResponse struct {
	Token     string        `json:"sessionId"`
	OTP       string        `json:"tempOTP"`
	Passenger PassengerInfo `json:"passenger"`
}

type PassengerInfo struct {
	ID              string `json:"id"`
	PassengerMobile string `json:"passengerMobile"`
	PassengerName   string `json:"passengerName"`
	Gender          string `json:"gender"`
	IsVerified      bool   `json:"isVerified"`
}

type PassengerProfileInfoResponse struct {
	PassengerDetails PassengerInfo `json:"passengerDetails"`
}

type PassengerChangePasswordRequest struct {
	OldPassword string `json:"oldPassword"`
	NewPassword string `json:"newPassword"`
}

type PassengerOTPResponse struct {
	OTP string `json:"tempOTP"`
}

// SchedulePreference is one leg of the standing weekly travel form. A passenger
// holds at most one of these per day per direction, which is what lets them travel
// in the morning but not in the evening, or be picked from one address and dropped
// at another.
type SchedulePreference struct {
	DayOfWeek     int     `json:"dayOfWeek"`
	Direction     string  `json:"direction"`
	IsEnabled     bool    `json:"isEnabled"`
	Location      string  `json:"location"`
	Lat           float64 `json:"lat"`
	Lng           float64 `json:"lng"`
	ScheduledTime string  `json:"scheduledTime"`
}

// PassengerScheduleRequest carries the whole weekly form in one call, the api
// replaces every leg it is given rather than making the app send one call per day.
type PassengerScheduleRequest struct {
	Preferences []SchedulePreference `json:"preferences"`
}

type PassengerScheduleResponse struct {
	Preferences []SchedulePreference `json:"preferences"`
}
