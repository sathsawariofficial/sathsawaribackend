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

func driverDetailsResp(drivers []postgress.Driver, totalRows int64) utils.APIResponse {
	driverDetails := []DriverWithVehicle{}

	for _, driver := range drivers {

		vehicles := []Vehicle{}
		for _, vehicle := range driver.Vehicles {
			vehicles = append(vehicles, Vehicle{
				ID:            vehicle.ID,
				DriverId:      vehicle.DriverId,
				VehicleNumber: vehicle.VehicleNumber,
				VehicleInfo:   vehicle.VehicleInfo,
				Status:        vehicle.Status,
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
	vehicleDetails := []Vehicle{}

	for _, vehicle := range vehicles {
		vehicleDetails = append(vehicleDetails, Vehicle{
			ID:            vehicle.ID,
			DriverId:      vehicle.DriverId,
			VehicleNumber: vehicle.VehicleNumber,
			VehicleInfo:   vehicle.VehicleInfo,
			Status:        vehicle.Status,
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

func rideDetailsResp(rides []postgress.RideDetails, totalRows int64) utils.APIResponse {
	rideDetails := []RideDetail{}

	for _, ride := range rides {
		rideDetails = append(rideDetails, RideDetail{
			ID:                   ride.ID,
			DriverID:             ride.DriverID,
			DriverName:           ride.DriverName,
			DriverMobile:         ride.DriverMobile,
			Rating:               ride.Rating,
			VehicleNumber:        ride.VehicleNumber,
			VehicleInfo:          ride.VehicleInfo,
			StartDatetime:        ride.StartDatetime,
			EstimatedEndDatetime: ride.EstimatedEndDatetime,
			NumberOfSeats:        ride.NumberOfSeats,
			SeatsTaken:           ride.SeatsTaken,
			StartLocation:        ride.StartLocation,
			EndLocation:          ride.EndLocation,
			Fare:                 ride.Fare,
			RouteDetails:         ride.RouteDetails,
			IsActive:             ride.IsActive,
		})
	}

	rideDetailsResp := utils.APIResponse{
		Code:    http.StatusOK,
		Message: constants.Success,
		Data: RideDetailsResponse{
			TotalPages: utils.CalculatePagesize(totalRows),
			Rides:      rideDetails,
		},
	}

	return rideDetailsResp
}

func createApprochInfoResp(approches []postgress.ApprochInfo, totalRows int64) utils.APIResponse {
	var approcheDetails []ApprochInfo

	for _, approch := range approches {
		approcheDetails = append(approcheDetails, ApprochInfo{
			Name:      approch.Name,
			Number:    approch.Number,
			Email:     approch.Email,
			Message:   approch.Message,
			Type:      approch.Type,
			CreatedAt: approch.CreatedAt,
		})
	}

	approchInfoResp := utils.APIResponse{
		Code:    http.StatusOK,
		Message: constants.Success,
		Data: ApprochInfoResponse{
			TotalPages: utils.CalculatePagesize(totalRows),
			Approches:  approcheDetails,
		},
	}

	return approchInfoResp
}

func createdRoleResp(roleId string) utils.APIResponse {
	return utils.APIResponse{
		Code:    http.StatusOK,
		Message: fmt.Sprintf(constants.Created_Successfully, "Role"),
		Data: AdminRoleCreatedResponse{
			Id: roleId,
		},
	}
}

func rolesResp(roles []postgress.Role, permissions map[string][]adminRolePermissionRow, totalRows int64) utils.APIResponse {
	roleDetails := []AdminRoleDetail{}

	for _, role := range roles {
		permissionDetails := []AdminPermissionDetail{}
		for _, permission := range permissions[role.ID] {
			permissionDetails = append(permissionDetails, AdminPermissionDetail{
				ID:          permission.ID,
				Code:        permission.Code,
				Description: permission.Description,
				IsSystem:    permission.IsSystem,
				CreatedAt:   permission.CreatedAt,
			})
		}

		roleDetails = append(roleDetails, AdminRoleDetail{
			ID:          role.ID,
			Name:        role.Name,
			Description: role.Description,
			IsSystem:    role.IsSystem,
			CreatedAt:   role.CreatedAt,
			Permissions: permissionDetails,
		})
	}

	return utils.APIResponse{
		Code:    http.StatusOK,
		Message: constants.Success,
		Data: AdminRoleListResponse{
			TotalPages: utils.CalculatePagesize(totalRows),
			Roles:      roleDetails,
		},
	}
}

func createdPermissionResp(permissionId string) utils.APIResponse {
	return utils.APIResponse{
		Code:    http.StatusOK,
		Message: fmt.Sprintf(constants.Created_Successfully, "Permission"),
		Data: AdminPermissionCreatedResponse{
			Id: permissionId,
		},
	}
}

func permissionsResp(permissions []postgress.Permission, totalRows int64) utils.APIResponse {
	permissionDetails := []AdminPermissionDetail{}

	for _, permission := range permissions {
		permissionDetails = append(permissionDetails, AdminPermissionDetail{
			ID:          permission.ID,
			Code:        permission.Code,
			Description: permission.Description,
			IsSystem:    permission.IsSystem,
			CreatedAt:   permission.CreatedAt,
		})
	}

	return utils.APIResponse{
		Code:    http.StatusOK,
		Message: constants.Success,
		Data: AdminPermissionListResponse{
			TotalPages:  utils.CalculatePagesize(totalRows),
			Permissions: permissionDetails,
		},
	}
}
