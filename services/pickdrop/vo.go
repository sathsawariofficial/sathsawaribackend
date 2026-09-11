package pickdrop

import (
	"fmt"
	"net/http"
	"rideshare/pkgs/constants"
	"rideshare/pkgs/database"
	"rideshare/pkgs/database/postgress"
	"rideshare/pkgs/utils"
)

func enableServiceResp(serviceId string) utils.APIResponse {
	return utils.APIResponse{
		Code:    http.StatusOK,
		Message: fmt.Sprintf(constants.Created_Successfully, "Pick & Drop service"),
		Data: EnableServiceResponse{
			ServiceId: serviceId,
		},
	}
}

func joinRequestResp(requestId string) utils.APIResponse {
	return utils.APIResponse{
		Code:    http.StatusOK,
		Message: constants.Request_Received_Successfully,
		Data: JoinRequestResponse{
			RequestId: requestId,
		},
	}
}

func mapServiceSummary(details *postgress.PickDropServiceDetails) *ServiceSummary {
	if details == nil {
		return nil
	}

	return &ServiceSummary{
		ID:            details.ID,
		Name:          details.Name,
		Description:   details.Description,
		OwnerDriverId: details.OwnerDriverID,
		OwnerName:     details.OwnerName,
		OwnerMobile:   details.OwnerMobile,
		Status:        details.Status,
		CreatedAt:     details.CreatedAt,
	}
}

func mapMemberVehicles(rows []postgress.PickDropVehicleDetails) []MemberVehicle {
	vehicles := []MemberVehicle{}

	for _, row := range rows {
		vehicles = append(vehicles, MemberVehicle{
			RequestId:     row.ID,
			VehicleId:     row.VehicleID,
			VehicleNumber: row.VehicleNumber,
			VehicleInfo:   row.VehicleInfo,
			NumberOfSeats: row.NumberOfSeats,
			HasValidSeats: row.NumberOfSeats >= constants.Number_Of_Seats_Min_Value,
			Status:        row.Status,
			RequestedAt:   row.RequestedAt,
			DecidedAt:     row.DecidedAt,
		})
	}

	return vehicles
}

func myServiceResp(data myServiceData) utils.APIResponse {
	resp := MyServiceResponse{
		Role:     data.Role,
		Service:  mapServiceSummary(data.Service),
		Vehicles: mapMemberVehicles(data.Vehicles),
	}

	if data.Counts != nil {
		resp.Counts = &ServiceCounts{
			ApprovedDrivers:       data.Counts.ApprovedDrivers,
			ApprovedVehicleOwners: data.Counts.ApprovedVehicleOwners,
			ApprovedVehicles:      data.Counts.ApprovedVehicles,
			ApprovedPassengers:    data.Counts.ApprovedPassengers,
			PendingRequests:       data.Counts.PendingRequests,
			ActiveShifts:          data.Counts.ActiveShifts,
			Advertisements:        data.Counts.Advertisements,
			AdvertisementLimit:    data.Counts.ApprovedVehicles * int64(constants.Advertisements_Per_Vehicle),
		}
	}

	if data.Membership != nil {
		resp.Membership = &MembershipInfo{
			RequestId:   data.Membership.ID,
			JoinType:    data.Membership.JoinType,
			Status:      data.Membership.Status,
			RequestedAt: data.Membership.RequestedAt,
			DecidedAt:   data.Membership.DecidedAt,
		}
	}

	return utils.APIResponse{
		Code:    http.StatusOK,
		Message: constants.Success,
		Data:    resp,
	}
}

func passengerMembershipResp(data passengerMembershipData) utils.APIResponse {
	resp := PassengerMembershipResponse{
		Role:    data.Role,
		Service: mapServiceSummary(data.Service),
	}

	if data.Membership != nil {
		resp.Membership = &MembershipInfo{
			RequestId:   data.Membership.ID,
			Status:      data.Membership.Status,
			RequestedAt: data.Membership.RequestedAt,
			DecidedAt:   data.Membership.DecidedAt,
		}
	}

	return utils.APIResponse{
		Code:    http.StatusOK,
		Message: constants.Success,
		Data:    resp,
	}
}

func serviceSearchResp(rows []postgress.PickDropServiceDetails, totalRows int64) utils.APIResponse {
	services := []ServiceSearchItem{}

	for _, row := range rows {
		services = append(services, ServiceSearchItem{
			ID:             row.ID,
			Name:           row.Name,
			Description:    row.Description,
			OwnerName:      row.OwnerName,
			DriverCount:    row.DriverCount,
			VehicleCount:   row.VehicleCount,
			PassengerCount: row.PassengerCount,
			MyStatus:       row.MyStatus,
			CreatedAt:      row.CreatedAt,
		})
	}

	return utils.APIResponse{
		Code:    http.StatusOK,
		Message: constants.Success,
		Data: ServiceSearchResponse{
			TotalPages: utils.CalculatePagesize(totalRows),
			Services:   services,
		},
	}
}

// weeklyDemand stitches the availability days and their locations back together,
// keyed by passenger.
func weeklyDemand(days []postgress.PassengerAvailability, locations []postgress.PassengerAvailabilityLocation) map[string][]AvailabilityDay {
	byDay := map[string][]utils.RouteLocation{}
	for _, location := range locations {
		key := fmt.Sprintf("%s:%d", location.PassengerID, location.DayOfWeek)
		byDay[key] = append(byDay[key], utils.RouteLocation{
			Location: location.Location,
			Lat:      location.Lat,
			Lng:      location.Lng,
			Time:     location.Time,
		})
	}

	demand := map[string][]AvailabilityDay{}
	for _, day := range days {
		dayLocations := byDay[fmt.Sprintf("%s:%d", day.PassengerID, day.DayOfWeek)]
		if dayLocations == nil {
			dayLocations = []utils.RouteLocation{}
		}

		demand[day.PassengerID] = append(demand[day.PassengerID], AvailabilityDay{
			DayOfWeek:  day.DayOfWeek,
			IsRequired: day.IsRequired,
			Locations:  dayLocations,
		})
	}

	return demand
}

func mapDriverRequest(row postgress.PickDropDriverDetails, vehicles []postgress.PickDropVehicleDetails) DriverRequestItem {
	return DriverRequestItem{
		RequestId:    row.ID,
		DriverId:     row.DriverID,
		DriverName:   row.DriverName,
		DriverMobile: row.DriverMobile,
		Rating:       row.Rating,
		JoinType:     row.JoinType,
		Status:       row.Status,
		RequestedAt:  row.RequestedAt,
		DecidedAt:    row.DecidedAt,
		Vehicles:     mapMemberVehicles(vehicles),
	}
}

func mapVehicleRequest(row postgress.PickDropVehicleDetails) VehicleRequestItem {
	return VehicleRequestItem{
		RequestId:     row.ID,
		VehicleId:     row.VehicleID,
		VehicleNumber: row.VehicleNumber,
		VehicleInfo:   row.VehicleInfo,
		NumberOfSeats: row.NumberOfSeats,
		HasValidSeats: row.NumberOfSeats >= constants.Number_Of_Seats_Min_Value,
		HasAC:         row.HasAC,
		HasHeating:    row.HasHeating,
		DriverId:      row.DriverID,
		DriverName:    row.DriverName,
		DriverMobile:  row.DriverMobile,
		OwnerJoinType: row.OwnerJoinType,
		Status:        row.Status,
		RequestedAt:   row.RequestedAt,
		DecidedAt:     row.DecidedAt,
	}
}

func mapPassengerRequest(row postgress.PickDropPassengerDetails, demand []AvailabilityDay) PassengerRequestItem {
	return PassengerRequestItem{
		RequestId:       row.ID,
		PassengerId:     row.PassengerID,
		PassengerName:   row.PassengerName,
		PassengerMobile: row.PassengerMobile,
		Gender:          row.Gender,
		Status:          row.Status,
		RequestedAt:     row.RequestedAt,
		DecidedAt:       row.DecidedAt,
		WeeklyDemand:    demand,
	}
}

func driverRequestsPage(rows []postgress.PickDropDriverDetails, vehicles []postgress.PickDropVehicleDetails, totalRows int64) JoinRequestsResponse {
	byDriver := map[string][]postgress.PickDropVehicleDetails{}
	for _, vehicle := range vehicles {
		byDriver[vehicle.DriverID] = append(byDriver[vehicle.DriverID], vehicle)
	}

	requests := []DriverRequestItem{}
	for _, row := range rows {
		requests = append(requests, mapDriverRequest(row, byDriver[row.DriverID]))
	}

	return JoinRequestsResponse{
		TotalPages: utils.CalculatePagesize(totalRows),
		Type:       constants.Request_Type_Driver,
		Requests:   requests,
	}
}

func vehicleRequestsPage(rows []postgress.PickDropVehicleDetails, totalRows int64) JoinRequestsResponse {
	requests := []VehicleRequestItem{}
	for _, row := range rows {
		requests = append(requests, mapVehicleRequest(row))
	}

	return JoinRequestsResponse{
		TotalPages: utils.CalculatePagesize(totalRows),
		Type:       constants.Request_Type_Vehicle,
		Requests:   requests,
	}
}

func passengerRequestsPage(rows []postgress.PickDropPassengerDetails, totalRows int64) JoinRequestsResponse {
	requests := []PassengerRequestItem{}
	for _, row := range rows {
		requests = append(requests, mapPassengerRequest(row, nil))
	}

	return JoinRequestsResponse{
		TotalPages: utils.CalculatePagesize(totalRows),
		Type:       constants.Request_Type_Passenger,
		Requests:   requests,
	}
}

func joinRequestsResp(resp JoinRequestsResponse) utils.APIResponse {
	return utils.APIResponse{
		Code:    http.StatusOK,
		Message: constants.Success,
		Data:    resp,
	}
}

func joinRequestDetailResp(resp JoinRequestDetailResponse) utils.APIResponse {
	return utils.APIResponse{
		Code:    http.StatusOK,
		Message: constants.Success,
		Data:    resp,
	}
}

func decideRequestsResp(applied, skipped []DecisionResult) utils.APIResponse {
	if applied == nil {
		applied = []DecisionResult{}
	}
	if skipped == nil {
		skipped = []DecisionResult{}
	}

	return utils.APIResponse{
		Code:    http.StatusOK,
		Message: fmt.Sprintf(constants.Updated_Successfully, "Requests"),
		Data: DecideRequestsResponse{
			Applied: applied,
			Skipped: skipped,
		},
	}
}

// availability works out the isAvailable flag of one listed entry. It is only filled
// when the owner sent a schedule, and an entry that is not eligible at all, like a
// vehicle without seats, is never available whatever its schedule.
func availability(enabled bool, conflicts map[string]database.ShiftClash, id string, eligible bool, reason string) (*bool, string) {
	if !enabled {
		return nil, ""
	}

	available := eligible
	if clash, found := conflicts[id]; found {
		available = false
		reason = clash.Message()
	}

	if available {
		reason = ""
	}

	return &available, reason
}

func availableDriversPage(rows []postgress.AvailableDriverDetails, conflicts map[string]database.ShiftClash, enabled bool, totalRows int64) AvailableDriversResponse {
	drivers := []AvailableDriverItem{}

	for _, row := range rows {
		isAvailable, conflict := availability(enabled, conflicts, row.DriverID, true, "")

		drivers = append(drivers, AvailableDriverItem{
			DriverId:     row.DriverID,
			DriverName:   row.DriverName,
			DriverMobile: row.DriverMobile,
			Rating:       row.Rating,
			IsOwner:      row.IsOwner,
			JoinedAt:     row.JoinedAt,
			ActiveShifts: row.ActiveShifts,
			IsAvailable:  isAvailable,
			Conflict:     conflict,
		})
	}

	return AvailableDriversResponse{
		TotalPages: utils.CalculatePagesize(totalRows),
		Drivers:    drivers,
	}
}

func availableVehiclesPage(rows []postgress.PickDropVehicleDetails, conflicts map[string]database.ShiftClash, enabled bool, totalRows int64) AvailableVehiclesResponse {
	vehicles := []AvailableVehicleItem{}

	for _, row := range rows {
		hasValidSeats := row.NumberOfSeats >= constants.Number_Of_Seats_Min_Value
		isAvailable, conflict := availability(enabled, conflicts, row.VehicleID, hasValidSeats, fmt.Sprintf(constants.Vehicle_Seats_Missing, row.VehicleNumber))

		vehicles = append(vehicles, AvailableVehicleItem{
			VehicleId:     row.VehicleID,
			VehicleNumber: row.VehicleNumber,
			VehicleInfo:   row.VehicleInfo,
			NumberOfSeats: row.NumberOfSeats,
			HasValidSeats: hasValidSeats,
			HasAC:         row.HasAC,
			HasHeating:    row.HasHeating,
			DriverId:      row.DriverID,
			DriverName:    row.DriverName,
			OwnerJoinType: row.OwnerJoinType,
			JoinedAt:      row.DecidedAt,
			ActiveShifts:  row.ActiveShifts,
			IsAvailable:   isAvailable,
			Conflict:      conflict,
		})
	}

	return AvailableVehiclesResponse{
		TotalPages: utils.CalculatePagesize(totalRows),
		Vehicles:   vehicles,
	}
}

func availablePassengersPage(rows []postgress.PickDropPassengerDetails, demand map[string][]AvailabilityDay, conflicts map[string]database.ShiftClash, enabled bool, totalRows int64) AvailablePassengersResponse {
	passengers := []AvailablePassengerItem{}

	for _, row := range rows {
		isAvailable, conflict := availability(enabled, conflicts, row.PassengerID, true, "")

		days := demand[row.PassengerID]
		if days == nil {
			days = []AvailabilityDay{}
		}

		passengers = append(passengers, AvailablePassengerItem{
			PassengerId:     row.PassengerID,
			PassengerName:   row.PassengerName,
			PassengerMobile: row.PassengerMobile,
			Gender:          row.Gender,
			JoinedAt:        row.DecidedAt,
			ActiveShifts:    row.ActiveShifts,
			WeeklyDemand:    days,
			IsAvailable:     isAvailable,
			Conflict:        conflict,
		})
	}

	return AvailablePassengersResponse{
		TotalPages: utils.CalculatePagesize(totalRows),
		Passengers: passengers,
	}
}

func availableDriversResp(resp AvailableDriversResponse) utils.APIResponse {
	return utils.APIResponse{Code: http.StatusOK, Message: constants.Success, Data: resp}
}

func availableVehiclesResp(resp AvailableVehiclesResponse) utils.APIResponse {
	return utils.APIResponse{Code: http.StatusOK, Message: constants.Success, Data: resp}
}

func availablePassengersResp(resp AvailablePassengersResponse) utils.APIResponse {
	return utils.APIResponse{Code: http.StatusOK, Message: constants.Success, Data: resp}
}

func advertisementResp(adId string) utils.APIResponse {
	return utils.APIResponse{
		Code:    http.StatusOK,
		Message: fmt.Sprintf(constants.Created_Successfully, "Advertisement"),
		Data: AdvertisementResponse{
			AdvertisementId: adId,
		},
	}
}

func mapAdvertisements(ads []postgress.AdvertisementDetails, locations []postgress.PickDropAdvertisementLocation) []AdvertisementItem {
	byAd := map[string][]utils.RouteLocation{}
	for _, location := range locations {
		byAd[location.AdvertisementID] = append(byAd[location.AdvertisementID], utils.RouteLocation{
			Location: location.Location,
			Lat:      location.Lat,
			Lng:      location.Lng,
			Time:     location.Time,
		})
	}

	items := []AdvertisementItem{}
	for _, ad := range ads {
		adLocations := byAd[ad.ID]
		if adLocations == nil {
			adLocations = []utils.RouteLocation{}
		}

		items = append(items, AdvertisementItem{
			ID:           ad.ID,
			ServiceId:    ad.ServiceID,
			ServiceName:  ad.ServiceName,
			OwnerName:    ad.OwnerName,
			OwnerMobile:  ad.OwnerMobile,
			Title:        ad.Title,
			Description:  ad.Description,
			Fare:         ad.Fare,
			DaysOfWeek:   database.IntsFromArray(ad.DaysOfWeek),
			StartTime:    ad.StartTime,
			EndTime:      ad.EndTime,
			VehicleCount: ad.VehicleCount,
			Locations:    adLocations,
			CreatedAt:    ad.CreatedAt,
		})
	}

	return items
}

func myAdvertisementsResp(resp MyAdvertisementsResponse) utils.APIResponse {
	return utils.APIResponse{Code: http.StatusOK, Message: constants.Success, Data: resp}
}

func advertisementSearchResp(ads []postgress.AdvertisementDetails, locations []postgress.PickDropAdvertisementLocation, totalRows int64) utils.APIResponse {
	return utils.APIResponse{
		Code:    http.StatusOK,
		Message: constants.Success,
		Data: AdvertisementSearchResponse{
			TotalPages:     utils.CalculatePagesize(totalRows),
			Advertisements: mapAdvertisements(ads, locations),
		},
	}
}
