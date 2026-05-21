package admin

import (
	"fmt"
	"net/http"
	"rideshare/pkgs/constants"
	"rideshare/pkgs/database/postgress"
	"rideshare/pkgs/utils"
)

func loginAdminResp(adminSessionId string, admin postgress.Admin) utils.APIResponse {
	adminDetails := AdminLoginResponse{
		Token: adminSessionId,
		Admin: AdminLogin{
			ID: admin.ID,
		},
	}

	code := http.StatusOK
	message := fmt.Sprintf(constants.Loggedin_Successfully, "Admin")

	return utils.APIResponse{
		Code:    code,
		Message: message,
		Data:    adminDetails,
	}
}

func driverDetailsResp(drivers []postgress.DriverWithVehicle, totalRows int64) utils.APIResponse {
	var driverDetails []DriverWithVehicle

	for _, driver := range drivers {

		var vehicles []Vehicle
		for _, v := range driver.Vehicles {
			vehicles = append(vehicles, Vehicle{
				ID:            v.ID,
				DriverId:      v.DriverId,
				VehicleNumber: v.VehicleNumber,
				VehicleInfo:   v.VehicleInfo,
				Status:        v.Status,
			})
		}

		driverDetails = append(driverDetails, DriverWithVehicle{
			ID:            driver.ID,
			DriverName:    driver.DriverName,
			DriverMobile:  driver.DriverMobile,
			Rating:        driver.Rating,
			NumberOfVotes: driver.NumberOfVotes,
			Status:        driver.Status,
			Vehicles:      vehicles,
		})
	}

	driverDetailsResp := utils.APIResponse{
		Code:    http.StatusOK,
		Message: constants.Success,
		Data: DriverDetailsResponse{
			TotalPages: utils.CalculatePagesize(totalRows),
			Details:    driverDetails,
		},
	}

	return driverDetailsResp
}

func vechileDetailsResp(vehicles []postgress.Vehicle, totalRows int64) utils.APIResponse {
	var vehicleDetails []Vehicle

	for _, v := range vehicles {
		vehicleDetails = append(vehicleDetails, Vehicle{
			ID:            v.ID,
			DriverId:      v.DriverId,
			VehicleNumber: v.VehicleNumber,
			VehicleInfo:   v.VehicleInfo,
			Status:        v.Status,
		})
	}

	vehicleDetailsResp := utils.APIResponse{
		Code:    http.StatusOK,
		Message: constants.Success,
		Data: VehicleDetailsResponse{
			TotalPages: utils.CalculatePagesize(totalRows),
			Vehicles:   vehicleDetails,
		},
	}

	return vehicleDetailsResp
}
