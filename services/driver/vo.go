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
		Message: fmt.Sprintf(constants.Registered_Successfully, "Driver"),
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

func updateVehicleResp(vehicleId string) utils.APIResponse {
	return utils.APIResponse{
		Code:    http.StatusOK,
		Message: fmt.Sprintf(constants.Updated_Successfully, "Vehicle"),
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
				DriverName:    driverDetails.DriverName,
				DriverMobile:  driverDetails.DriverMobile,
				VehicleNumber: driverDetails.VehicleNumber,
				VehicleInfo:   driverDetails.VehicleInfo,
				TotalRides:    driverDetails.TotalRides,
				Rating:        driverDetails.Rating,
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

func sendOTPResp(otp string) utils.APIResponse {
	return utils.APIResponse{
		Code:    http.StatusOK,
		Message: constants.Success,
		Data: ForgotPasswordResponse{
			OTP: otp,
		},
	}
}
