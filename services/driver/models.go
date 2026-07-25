package driver

import (
	"rideshare/pkgs/utils"
)

type DriverRegistrationRequest struct {
	DeviceId     string       `json:"deviceId"`
	MobileNumber string       `json:"mobile"`
	Name         string       `json:"name"`
	Gender       utils.Gender `json:"gender"`
	Password     string       `json:"password"`
	EXTPassword  string       `json:"-"`
}

type VehicleRegistrationRequest struct {
	VehicleNumber string `json:"vehicleNumber"`
	VehicleInfo   string `json:"vehicleInfo"`
	Pin           string `json:"pin"`
}

type VehicleUpdateRequest struct {
	VehicleId     string `json:"vehicleId"`
	VehicleNumber string `json:"vehicleNumber"`
	VehicleInfo   string `json:"vehicleInfo"`
	Status        string `json:"status"`
	Pin           string `json:"pin"`
}

type DriverLoginRequest struct {
	DeviceId     string `json:"deviceId"`
	MobileNumber string `json:"mobile"`
	Password     string `json:"password"`
	FCM          string `json:"fcm"`
}

type DriverLoginResponse struct {
	Token  string      `json:"sessionId"`
	OTP    string      `json:"tempOTP"`
	Driver DriverLogin `json:"driver"`
}

type DriverRegistrationResponse struct {
	DriverId string `json:"driverId"`
	OTP      string `json:"tempOTP"`
}

type VehicleRegistrationResponse struct {
	VehicleId string `json:"vehicleId"`
	OTP       string `json:"tempOTP"`
}

type VehicleUpdateResponse struct {
	VehicleId string `json:"vehicleId"`
}

type RateDriverRequest struct {
	DriverId     string `json:"driverId"`
	RideId       string `json:"rideId"`
	MobileNumber string `json:"mobileNumber"`
	Rating       int    `json:"rating"`
}

type UpdateDriverRequest struct {
	DriverId     string `json:"driverId"`
	MobileNumber string `json:"mobile"`
	Password     string `json:"password"`
}

type DriverProfileInfoResponse struct {
	DriverDetails DriverDetails `json:"driverDetails"`
}

type DriverDetails struct {
	ID           string `json:"id"`
	VehicleID    string `json:"vehicleId"`
	DriverMobile string `json:"driverMobile"`
	DriverName   string `json:"driverName"`
	Rating       string `json:"rating"`
	HasPin       bool   `json:"hasPin"`
}

type DriverLogin struct {
	ID           string `json:"id"`
	VehicleID    string `json:"vehicleId"`
	DriverMobile string `json:"driverMobile"`
	DriverName   string `json:"driverName"`
	Rating       string `json:"rating"`
	HasPin       bool   `json:"hasPin"`
}

type VehiclesResponse struct {
	Vehicles []Vehicles `json:"vehicles"`
}

type Vehicles struct {
	ID            string `json:"id"`
	DriverId      string `json:"driverId"`
	VehicleNumber string `json:"vehicleNumber"`
	VehicleInfo   string `json:"vehicleInfo"`
	Status        string `json:"status"`
}

type ChangePasswordRequest struct {
	OldPassword string `json:"oldPassword"`
	NewPassword string `json:"newPassword"`
}

type ChangePasswordResponse struct {
	OTP string `json:"tempOTP"`
}

type ForgotPasswordResponse struct {
	OTP string `json:"tempOTP"`
}

type ChangePinRequest struct {
	OldPin string `json:"oldPin"`
	NewPin string `json:"newPin"`
}

type ChangePinResponse struct {
	OTP string `json:"tempOTP"`
}

type ForgotPinResponse struct {
	OTP string `json:"tempOTP"`
}

type BookSeatRequest struct {
	RideId       string `json:"rideId"`
	Name         string `json:"name"`
	MobileNumber string `json:"mobileNumber"`
	Code         string `json:"code"`
	Seats        int    `json:"seats"`
	IsBook       bool   `json:"isBook"`
}

type BookSeatResponse struct {
	BookingId string `json:"bookingId"`
}
type BookedSeat struct {
	ID           string `json:"id"`
	BookingID    string `json:"bookingId"`
	MobileNumber string `json:"mobileNumber"`
	Name         string `json:"name"`
	Seats        int    `json:"seats"`
	Reserved     bool   `json:"reserved"`
}

type BookedSeatsResponse struct {
	Bookings []BookedSeat `json:"bookings"`
}
