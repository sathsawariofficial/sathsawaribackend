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

func platformOverviewResp(o postgress.PlatformOverview) utils.APIResponse {
	return utils.APIResponse{
		Code:    http.StatusOK,
		Message: constants.Success,
		Data: AdminOverviewResponse{
			Drivers:             AdminCountPair{Total: o.TotalDrivers, Live: o.ActiveDrivers, Label: "active"},
			Passengers:          AdminCountPair{Total: o.TotalPassengers, Live: o.ActivePassengers, Label: "active"},
			Vehicles:            AdminCountPair{Total: o.TotalVehicles, Live: o.SeatedVehicles, Label: "seats recorded"},
			Rides:               AdminCountPair{Total: o.TotalRides, Live: o.ActiveRides, Label: "active"},
			Services:            AdminCountPair{Total: o.TotalServices, Live: o.ActiveServices, Label: "active"},
			Shifts:              AdminCountPair{Total: o.TotalShifts, Live: o.ActiveShifts, Label: "active"},
			Advertisements:      o.TotalAdvertisements,
			OpenShiftRequests:   o.OpenShiftRequests,
			PendingJoinRequests: o.PendingJoinRequests,
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

func passengerProfileResp(passenger postgress.Passenger) utils.APIResponse {
	return utils.APIResponse{
		Code:    http.StatusOK,
		Message: constants.Success,
		Data: AdminPassengerProfileResponse{
			Passenger: mapPassengerDetail(passenger),
		},
	}
}

////////////////////////////// PICK & DROP OVERSIGHT //////////////////////////////

func mapServiceSummary(row postgress.PickDropServiceDetails) AdminServiceSummary {
	return AdminServiceSummary{
		ID:             row.ID,
		Name:           row.Name,
		Description:    row.Description,
		OwnerDriverId:  row.OwnerDriverID,
		OwnerName:      row.OwnerName,
		OwnerMobile:    row.OwnerMobile,
		Status:         row.Status,
		DriverCount:    row.DriverCount,
		VehicleCount:   row.VehicleCount,
		PassengerCount: row.PassengerCount,
		ShiftCount:     row.ShiftCount,
		CreatedAt:      row.CreatedAt,
	}
}

func servicesResp(services []postgress.PickDropServiceDetails, totalRows int64) utils.APIResponse {
	details := []AdminServiceSummary{}
	for _, row := range services {
		details = append(details, mapServiceSummary(row))
	}

	return utils.APIResponse{
		Code:    http.StatusOK,
		Message: constants.Success,
		Data: AdminServicesResponse{
			TotalPages: utils.CalculatePagesize(totalRows),
			Services:   details,
		},
	}
}

func mapShiftSummary(row postgress.ShiftDetails) AdminShiftSummary {
	days := make([]int, 0, len(row.DaysOfWeek))
	for _, d := range row.DaysOfWeek {
		days = append(days, int(d))
	}

	return AdminShiftSummary{
		ID:               row.ID,
		Name:             row.Name,
		ServiceId:        row.ServiceID,
		ServiceName:      row.ServiceName,
		ServiceOwnerId:   row.OwnerDriverID,
		DriverId:         row.DriverID,
		DriverName:       row.DriverName,
		DriverMobile:     row.DriverMobile,
		VehicleId:        row.VehicleID,
		VehicleNumber:    row.VehicleNumber,
		VehicleOwnerId:   row.VehicleOwnerID,
		VehicleOwnerName: row.VehicleOwnerName,
		DaysOfWeek:       days,
		StartDate:        row.StartDate,
		EndDate:          row.EndDate,
		StartTime:        row.StartTime,
		EndTime:          row.EndTime,
		SeatCapacity:     row.SeatCapacity,
		OccupiedSeats:    row.OccupiedSeats,
		Status:           row.Status,
		CreatedAt:        row.CreatedAt,
	}
}

func serviceDetailResp(service postgress.PickDropServiceDetails, counts pickDropServiceCountsRow, shifts []postgress.ShiftDetails) utils.APIResponse {
	summaries := []AdminShiftSummary{}
	for _, row := range shifts {
		summaries = append(summaries, mapShiftSummary(row))
	}

	return utils.APIResponse{
		Code:    http.StatusOK,
		Message: constants.Success,
		Data: AdminServiceDetailResponse{
			Service: mapServiceSummary(service),
			Counts: AdminServiceCounts{
				ApprovedDrivers:       counts.ApprovedDrivers,
				ApprovedVehicleOwners: counts.ApprovedVehicleOwners,
				ApprovedVehicles:      counts.ApprovedVehicles,
				ApprovedPassengers:    counts.ApprovedPassengers,
				PendingRequests:       counts.PendingRequests,
				ActiveShifts:          counts.ActiveShifts,
			},
			Shifts: summaries,
		},
	}
}

func shiftsResp(shifts []postgress.ShiftDetails, totalRows int64) utils.APIResponse {
	details := []AdminShiftSummary{}
	for _, row := range shifts {
		details = append(details, mapShiftSummary(row))
	}

	return utils.APIResponse{
		Code:    http.StatusOK,
		Message: constants.Success,
		Data: AdminShiftsResponse{
			TotalPages: utils.CalculatePagesize(totalRows),
			Shifts:     details,
		},
	}
}

func shiftRequestsResp(requests []postgress.ShiftRequestDetails, totalRows int64) utils.APIResponse {
	details := []AdminShiftRequestItem{}
	for _, row := range requests {
		days := make([]int, 0, len(row.DaysOfWeek))
		for _, d := range row.DaysOfWeek {
			days = append(days, int(d))
		}

		details = append(details, AdminShiftRequestItem{
			ID:            row.ID,
			PassengerId:   row.PassengerID,
			PassengerName: row.PassengerName,
			ServiceId:     row.ServiceID,
			ServiceName:   row.ServiceName,
			ContactNumber: row.ContactNumber,
			Note:          row.Note,
			DaysOfWeek:    days,
			StartTime:     row.StartTime,
			EndTime:       row.EndTime,
			CreatedAt:     row.CreatedAt,
		})
	}

	return utils.APIResponse{
		Code:    http.StatusOK,
		Message: constants.Success,
		Data: AdminShiftRequestsResponse{
			TotalPages: utils.CalculatePagesize(totalRows),
			Requests:   details,
		},
	}
}

func advertisementsResp(ads []postgress.AdvertisementDetails, totalRows int64) utils.APIResponse {
	details := []AdminAdvertisementItem{}
	for _, row := range ads {
		days := make([]int, 0, len(row.DaysOfWeek))
		for _, d := range row.DaysOfWeek {
			days = append(days, int(d))
		}

		details = append(details, AdminAdvertisementItem{
			ID:            row.ID,
			ServiceId:     row.ServiceID,
			ServiceName:   row.ServiceName,
			OwnerDriverId: row.OwnerDriverID,
			OwnerName:     row.OwnerName,
			OwnerMobile:   row.OwnerMobile,
			Title:         row.Title,
			Description:   row.Description,
			Fare:          row.Fare,
			DaysOfWeek:    days,
			StartTime:     row.StartTime,
			EndTime:       row.EndTime,
			VehicleCount:  row.VehicleCount,
			CreatedAt:     row.CreatedAt,
		})
	}

	return utils.APIResponse{
		Code:    http.StatusOK,
		Message: constants.Success,
		Data: AdminAdvertisementsResponse{
			TotalPages:     utils.CalculatePagesize(totalRows),
			Advertisements: details,
		},
	}
}
