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

func platformOverviewResp(o postgress.PlatformOverview) utils.APIResponse {
	return utils.APIResponse{
		Code:    http.StatusOK,
		Message: constants.Success,
		Data: AdminOverviewResponse{
			Drivers:              AdminCountPair{Total: o.TotalDrivers, Live: o.ActiveDrivers, Label: "active"},
			Passengers:           AdminCountPair{Total: o.TotalPassengers, Live: o.ActivePassengers, Label: "active"},
			Vehicles:             AdminCountPair{Total: o.TotalVehicles, Live: o.SeatedVehicles, Label: "fleet ready"},
			Groups:               AdminCountPair{Total: o.TotalGroups, Live: o.ActiveGroups, Label: "active"},
			Shifts:               AdminCountPair{Total: o.TotalShifts, Live: o.UpcomingShifts, Label: "upcoming"},
			Rides:                AdminCountPair{Total: o.TotalRides, Live: o.ActiveRides, Label: "active"},
			PendingGroupRequests: o.PendingRequests,
		},
	}
}

func mapPassengerDetail(passenger postgress.Passenger) AdminPassengerDetail {
	return AdminPassengerDetail{
		ID:              passenger.ID,
		PassengerName:   passenger.PassengerName,
		PassengerMobile: passenger.PassengerMobile,
		Gender:          passenger.Gender,
		Status:          passenger.Status,
		CreatedAt:       passenger.CreatedAt,
	}
}

func passengersResp(passengers []postgress.Passenger, totalRows int64) utils.APIResponse {
	details := []AdminPassengerDetail{}
	for _, passenger := range passengers {
		details = append(details, mapPassengerDetail(passenger))
	}

	return utils.APIResponse{
		Code:    http.StatusOK,
		Message: constants.Success,
		Data: AdminPassengerListResponse{
			TotalPages: utils.CalculatePagesize(totalRows),
			Passengers: details,
		},
	}
}

func passengerProfileResp(
	passenger postgress.Passenger,
	groups []postgress.GroupPassengerDetails,
	preferences []postgress.PassengerLocationPreference,
) utils.APIResponse {
	groupDetails := []AdminPassengerGroup{}
	for _, group := range groups {
		groupDetails = append(groupDetails, AdminPassengerGroup{
			GroupId: group.GroupID,
			// the joined query aliases the group name into this column
			GroupName: group.PassengerName,
			Status:    group.Status,
			JoinedAt:  group.CreatedAt,
		})
	}

	form := []AdminSchedulePreference{}
	for _, preference := range preferences {
		form = append(form, AdminSchedulePreference{
			DayOfWeek:     preference.DayOfWeek,
			Direction:     preference.Direction,
			IsEnabled:     preference.IsEnabled,
			Location:      preference.Location,
			Lat:           preference.Lat,
			Lng:           preference.Lng,
			ScheduledTime: preference.ScheduledTime,
		})
	}

	return utils.APIResponse{
		Code:    http.StatusOK,
		Message: constants.Success,
		Data: AdminPassengerProfileResponse{
			Passenger:   mapPassengerDetail(passenger),
			Groups:      groupDetails,
			TravelForm:  form,
			TotalGroups: len(groupDetails),
		},
	}
}

func mapGroupOverview(group postgress.AdminGroupOverview) AdminGroupDetail {
	return AdminGroupDetail{
		ID:             group.ID,
		Name:           group.Name,
		Description:    group.Description,
		Status:         group.Status,
		OwnerId:        group.OwnerDriverID,
		OwnerName:      group.OwnerName,
		OwnerMobile:    group.OwnerMobile,
		MemberCount:    group.MemberCount,
		VehicleCount:   group.VehicleCount,
		PassengerCount: group.PassengerCount,
		ShiftCount:     group.ShiftCount,
		CreatedAt:      group.CreatedAt,
	}
}

func adminGroupsResp(groups []postgress.AdminGroupOverview, totalRows int64) utils.APIResponse {
	details := []AdminGroupDetail{}
	for _, group := range groups {
		details = append(details, mapGroupOverview(group))
	}

	return utils.APIResponse{
		Code:    http.StatusOK,
		Message: constants.Success,
		Data: AdminGroupListResponse{
			TotalPages: utils.CalculatePagesize(totalRows),
			Groups:     details,
		},
	}
}

func adminGroupDetailsResp(
	overview postgress.AdminGroupOverview,
	members []postgress.GroupMemberDetails,
	vehicles []postgress.GroupVehicleDetails,
	passengers []postgress.GroupPassengerDetails,
) utils.APIResponse {
	memberDetails := []AdminGroupMember{}
	for _, member := range members {
		memberDetails = append(memberDetails, AdminGroupMember{
			ID:           member.ID,
			DriverId:     member.DriverID,
			DriverName:   member.DriverName,
			DriverMobile: member.DriverMobile,
			RoleName:     member.RoleName,
			JoinType:     member.JoinType,
			Status:       member.Status,
			CreatedAt:    member.CreatedAt,
		})
	}

	vehicleDetails := []AdminGroupVehicle{}
	for _, vehicle := range vehicles {
		vehicleDetails = append(vehicleDetails, AdminGroupVehicle{
			ID:            vehicle.ID,
			VehicleId:     vehicle.VehicleID,
			VehicleNumber: vehicle.VehicleNumber,
			VehicleInfo:   vehicle.VehicleInfo,
			NumberOfSeats: vehicle.NumberOfSeats,
			HasAC:         vehicle.HasAC,
			HasHeating:    vehicle.HasHeating,
			DriverName:    vehicle.DriverName,
			Status:        vehicle.Status,
			CreatedAt:     vehicle.CreatedAt,
		})
	}

	passengerDetails := []AdminGroupPassenger{}
	for _, passenger := range passengers {
		passengerDetails = append(passengerDetails, AdminGroupPassenger{
			ID:              passenger.ID,
			PassengerId:     passenger.PassengerID,
			PassengerName:   passenger.PassengerName,
			PassengerMobile: passenger.PassengerMobile,
			Gender:          passenger.Gender,
			Status:          passenger.Status,
			CreatedAt:       passenger.CreatedAt,
		})
	}

	return utils.APIResponse{
		Code:    http.StatusOK,
		Message: constants.Success,
		Data: AdminGroupDetailsResponse{
			Group:      mapGroupOverview(overview),
			Members:    memberDetails,
			Vehicles:   vehicleDetails,
			Passengers: passengerDetails,
		},
	}
}

func mapShiftDetail(shift postgress.ShiftDetails) AdminShiftDetail {
	return AdminShiftDetail{
		ID:                   shift.ID,
		GroupId:              shift.GroupID,
		GroupName:            shift.GroupName,
		VehicleNumber:        shift.VehicleNumber,
		VehicleInfo:          shift.VehicleInfo,
		DriverId:             shift.DriverID,
		DriverName:           shift.DriverName,
		DriverMobile:         shift.DriverMobile,
		Direction:            shift.Direction,
		StartDatetime:        shift.StartDatetime,
		EstimatedEndDatetime: shift.EstimatedEndDatetime,
		StartLocation:        shift.StartLocation,
		EndLocation:          shift.EndLocation,
		NumberOfSeats:        shift.NumberOfSeats,
		SeatsTaken:           shift.SeatsTaken,
		CreatedByName:        shift.CreatedByName,
		CreatedByMobile:      shift.CreatedByMobile,
		IsActive:             shift.IsActive,
		CreatedAt:            shift.CreatedAt,
	}
}

func adminShiftsResp(shifts []postgress.ShiftDetails, totalRows int64) utils.APIResponse {
	details := []AdminShiftDetail{}
	for _, shift := range shifts {
		details = append(details, mapShiftDetail(shift))
	}

	return utils.APIResponse{
		Code:    http.StatusOK,
		Message: constants.Success,
		Data: AdminShiftListResponse{
			TotalPages: utils.CalculatePagesize(totalRows),
			Shifts:     details,
		},
	}
}

func adminShiftDetailsResp(shift postgress.ShiftDetails, stops []postgress.ShiftStop, seats []postgress.ShiftSeatDetails) utils.APIResponse {
	stopDetails := []AdminShiftStop{}
	for _, stop := range stops {
		stopDetails = append(stopDetails, AdminShiftStop{
			SequenceNumber: stop.SequenceNumber,
			Location:       stop.Location,
			Lat:            stop.Lat,
			Lng:            stop.Lng,
			ScheduledTime:  stop.ScheduledTime,
		})
	}

	seatDetails := []AdminShiftSeat{}
	for _, seat := range seats {
		seatDetails = append(seatDetails, AdminShiftSeat{
			SeatNumber:      seat.SeatNumber,
			Gender:          seat.Gender,
			Status:          seat.Status,
			PassengerName:   seat.PassengerName,
			PassengerMobile: seat.PassengerMobile,
			StopSequence:    seat.SequenceNumber,
			Location:        seat.Location,
			ScheduledTime:   seat.ScheduledTime,
		})
	}

	return utils.APIResponse{
		Code:    http.StatusOK,
		Message: constants.Success,
		Data: AdminShiftDetailsResponse{
			Shift: mapShiftDetail(shift),
			Stops: stopDetails,
			Seats: seatDetails,
		},
	}
}
