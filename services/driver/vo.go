package driver

import (
	"fmt"
	"net/http"
	"rideshare/pkgs/constants"
	"rideshare/pkgs/database/postgress"
	"rideshare/pkgs/utils"
)

func registerDriverResp(driverId, otp string) utils.APIResponse {
	return utils.APIResponse{
		Code:    http.StatusOK,
		Message: "OTP sent to your registered number. Please verify.",
		Data: DriverRegistrationResponse{
			DriverId: driverId,
			OTP:      otp,
		},
	}
}

func registerVehicleResp(vehicleId, otp string) utils.APIResponse {
	return utils.APIResponse{
		Code:    http.StatusOK,
		Message: fmt.Sprintf(constants.Registered_Successfully, "Vehicle"),
		Data: VehicleRegistrationResponse{
			VehicleId: vehicleId,
			OTP:       otp,
		},
	}
}

func updateVehicleResp(vehicleId, status string) utils.APIResponse {
	msg := fmt.Sprintf(constants.Updated_Successfully, "Vehicle")
	if status == constants.Status_InActive {
		msg = fmt.Sprintf(constants.Deleted_Successfully, "Vehicle")
	}

	return utils.APIResponse{
		Code:    http.StatusOK,
		Message: msg,
		Data: VehicleUpdateResponse{
			VehicleId: vehicleId,
		},
	}
}

func loginDriverResp(driverSessionId, otp string, driver postgress.Driver) utils.APIResponse {
	driverDetails := DriverLoginResponse{
		Token: driverSessionId,
		OTP:   otp,
		Driver: DriverLogin{
			ID:           driver.ID,
			DriverMobile: driver.DriverMobile,
			DriverName:   driver.DriverName,
			Rating:       driver.Rating,
			HasPin:       !utils.IsStringEmpty(driver.Pin),
		},
	}

	code := http.StatusOK
	message := fmt.Sprintf(constants.Loggedin_Successfully, "Driver")
	if !utils.IsStringEmpty(otp) {
		code = http.StatusAccepted
		message = constants.Unverified_Driver
	}

	return utils.APIResponse{
		Code:    code,
		Message: message,
		Data:    driverDetails,
	}
}

func logoutResp() utils.APIResponse {
	return utils.APIResponse{
		Code:    http.StatusOK,
		Message: fmt.Sprintf(constants.Loggedout_Successfully, "Driver"),
	}
}

func rateDriverResp() utils.APIResponse {
	return utils.APIResponse{
		Code:    http.StatusOK,
		Message: fmt.Sprintf(constants.Added_Successfully, "Rating"),
	}
}

func dirverProfileResp(driverDetails postgress.DriverDetails) utils.APIResponse {
	return utils.APIResponse{
		Code:    http.StatusOK,
		Message: constants.Success,
		Data: DriverProfileInfoResponse{
			DriverDetails: DriverDetails{
				ID:           driverDetails.ID,
				DriverMobile: driverDetails.DriverMobile,
				DriverName:   driverDetails.DriverName,
				Rating:       driverDetails.Rating,
				HasPin:       driverDetails.HasPin,
			},
		},
	}
}

func vehicleInfoResp(vehicles []postgress.Vehicle) utils.APIResponse {
	var vehicleDetails []Vehicles

	for _, val := range vehicles {
		vehicleDetails = append(vehicleDetails, Vehicles{
			ID:            val.ID,
			DriverId:      val.DriverId,
			VehicleNumber: val.VehicleNumber,
			VehicleInfo:   val.VehicleInfo,
			Status:        val.Status,
		})
	}

	return utils.APIResponse{
		Code:    http.StatusOK,
		Message: constants.Success,
		Data: &VehiclesResponse{
			Vehicles: vehicleDetails,
		},
	}
}

func changePasswordResp(otp string) utils.APIResponse {
	return utils.APIResponse{
		Code:    http.StatusOK,
		Message: constants.SENT_OTP_Successfully,
		Data: ChangePasswordResponse{
			OTP: otp,
		},
	}
}

func forgotPasswordResp(otp string) utils.APIResponse {
	return utils.APIResponse{
		Code:    http.StatusOK,
		Message: constants.SENT_OTP_Successfully,
		Data: ForgotPasswordResponse{
			OTP: otp,
		},
	}
}

func changePinResp(otp string) utils.APIResponse {
	return utils.APIResponse{
		Code:    http.StatusOK,
		Message: constants.SENT_OTP_Successfully,
		Data: ChangePinResponse{
			OTP: otp,
		},
	}
}

func forgotPinResp(otp string) utils.APIResponse {
	return utils.APIResponse{
		Code:    http.StatusOK,
		Message: constants.SENT_OTP_Successfully,
		Data: ForgotPinResponse{
			OTP: otp,
		},
	}
}

func getBookedSeatsResp(bookings []postgress.RideBooking) utils.APIResponse {
	var bookedSeats []BookedSeat

	for _, val := range bookings {
		bookedSeats = append(bookedSeats, BookedSeat{
			ID:           val.ID,
			BookingID:    val.BookingID,
			MobileNumber: val.MobileNumber,
			Name:         val.Name,
			Seats:        val.Seats,
		})
	}

	msg := constants.Success
	if len(bookedSeats) == 0 {
		msg = "No Seat has been booked yet."
	}

	return utils.APIResponse{
		Code:    http.StatusOK,
		Message: msg,
		Data: &BookedSeatsResponse{
			Bookings: bookedSeats,
		},
	}
}

func bookSeatResponse(msg, uuid string) utils.APIResponse {
	return utils.APIResponse{
		Code:    http.StatusOK,
		Message: msg,
		Data: BookSeatResponse{
			BookingId: uuid,
		},
	}
}
