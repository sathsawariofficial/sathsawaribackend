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

// AvailabilityDay is one day of a Pick & Drop passenger's weekly requirement: whether
// they need the service that day and, when they do, up to six places each with the
// time they need to be there.
type AvailabilityDay struct {
	DayOfWeek  int                   `json:"dayOfWeek"`
	IsRequired bool                  `json:"isRequired"`
	Locations  []utils.RouteLocation `json:"locations"`
}

// AvailabilityRequest replaces the days it carries in one call, a day left out stays
// as it was.
type AvailabilityRequest struct {
	Days []AvailabilityDay `json:"days"`
}

// AvailabilityResponse always carries the whole week, Monday to Sunday.
type AvailabilityResponse struct {
	ServiceId string            `json:"serviceId"`
	Days      []AvailabilityDay `json:"days"`
}

// NotificationSettingsRequest is the whole place notification setting of a passenger's
// device: while enabled, rides taking in one of the places are pushed to the fcm. What
// is sent replaces what was saved before.
type NotificationSettingsRequest struct {
	DeviceId string   `json:"deviceId"`
	FCM      string   `json:"fcm"`
	Enabled  bool     `json:"enabled"`
	Places   []string `json:"places"`
}

type NotificationSettingsResponse struct {
	DeviceId string   `json:"deviceId"`
	Enabled  bool     `json:"enabled"`
	Places   []string `json:"places"`
}
