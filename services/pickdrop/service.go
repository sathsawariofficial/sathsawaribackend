package pickdrop

import (
	"errors"
	"fmt"
	"rideshare/pkgs/constants"
	"rideshare/pkgs/database"
	"rideshare/pkgs/database/postgress"
	"rideshare/pkgs/logger"
	"rideshare/pkgs/utils"

	"github.com/gin-gonic/gin"
)

// EnableService lets an already registered driver start Pick & Drop. The driver
// becomes the one owner of the new service. A driver who owns a service or belongs
// to one, or is still waiting on a request, cannot start another.
func EnableService(ctx *gin.Context, sessionId, driverId string, request EnableServiceRequest) (serviceId string, err error) {
	logger.LogInfo("Request received in EnableService", sessionId)

	serviceId, err = createService(ctx, sessionId, driverId, request)
	if err != nil {
		err = utils.ClientError(sessionId, err, fmt.Sprintf(constants.Creation_Failed, "Pick & Drop service"))
		return
	}

	logger.LogInfo("Response returned from EnableService", sessionId)
	logger.LogDebug2("Response returned from EnableService", sessionId, serviceId)

	return
}

func GetMyService(ctx *gin.Context, sessionId, driverId string) (data myServiceData, err error) {
	logger.LogInfo("Request received in GetMyService", sessionId)

	data, err = getMyService(ctx, driverId)
	if err != nil {
		logger.LogError(sessionId, "failed to get the service error: "+err.Error())
		err = fmt.Errorf(constants.Failed_To_Do_Job, "get the Pick & Drop service")
		return
	}

	logger.LogInfo("Response returned from GetMyService", sessionId)

	return
}

// DisableService switches the owner's service off once no shift depends on it.
func DisableService(ctx *gin.Context, sessionId, driverId string) (err error) {
	logger.LogInfo("Request received in DisableService", sessionId)

	service, notices, err := disableService(ctx, sessionId, driverId)
	if err != nil {
		err = utils.ClientError(sessionId, err, fmt.Sprintf(constants.Update_Failed, "Pick & Drop service"))
		return
	}

	sendNotices(ctx, sessionId, service.ID, notices)

	logger.LogInfo("Response returned from DisableService", sessionId)

	return
}

func SearchServices(ctx *gin.Context, sessionId, userId, search string, page int) (services []postgress.PickDropServiceDetails, totalRows int64, err error) {
	logger.LogInfo("Request received in SearchServices", sessionId)

	services, totalRows, err = searchServices(ctx, userId, search, page)
	if err != nil {
		logger.LogError(sessionId, "failed to search the services error: "+err.Error())
		err = fmt.Errorf(constants.Failed_To_Do_Job, "search the Pick & Drop services")
		return
	}

	logger.LogInfo("Response returned from SearchServices", sessionId)

	return
}

// AddServiceVehicles puts more of the owner's own vehicles into their service.
func AddServiceVehicles(ctx *gin.Context, sessionId, driverId string, request VehicleIdsRequest) (err error) {
	logger.LogInfo("Request received in AddServiceVehicles", sessionId)

	if err = addOwnerVehicles(ctx, sessionId, driverId, request.VehicleIds); err != nil {
		err = utils.ClientError(sessionId, err, fmt.Sprintf(constants.Failed_To_Do_Job, "add the vehicles"))
		return
	}

	logger.LogInfo("Response returned from AddServiceVehicles", sessionId)

	return
}

func RemoveServiceVehicle(ctx *gin.Context, sessionId, driverId, vehicleId string) (err error) {
	logger.LogInfo("Request received in RemoveServiceVehicle", sessionId)

	service, notices, err := removeServiceVehicle(ctx, sessionId, driverId, vehicleId)
	if err != nil {
		err = utils.ClientError(sessionId, err, fmt.Sprintf(constants.Failed_To_Do_Job, "remove the vehicle"))
		return
	}

	sendNotices(ctx, sessionId, service.ID, notices)

	logger.LogInfo("Response returned from RemoveServiceVehicle", sessionId)

	return
}

// RequestJoinAsDriver puts a driver in the queue of a service with the vehicles they
// propose. The owner decides on the driver and on each vehicle.
func RequestJoinAsDriver(ctx *gin.Context, sessionId, driverId string, request DriverJoinRequest) (requestId string, err error) {
	logger.LogInfo("Request received in RequestJoinAsDriver", sessionId)

	service, requestId, notices, err := requestJoinAsDriver(ctx, sessionId, driverId, request)
	if err != nil {
		err = utils.ClientError(sessionId, err, fmt.Sprintf(constants.Failed_To_Do_Job, "send the request"))
		return
	}

	sendNotices(ctx, sessionId, service.ID, notices)

	logger.LogInfo("Response returned from RequestJoinAsDriver", sessionId)

	return
}

func OfferVehicles(ctx *gin.Context, sessionId, driverId string, request VehicleIdsRequest) (err error) {
	logger.LogInfo("Request received in OfferVehicles", sessionId)

	service, notices, err := offerVehicles(ctx, sessionId, driverId, request.VehicleIds)
	if err != nil {
		err = utils.ClientError(sessionId, err, fmt.Sprintf(constants.Failed_To_Do_Job, "offer the vehicles"))
		return
	}

	sendNotices(ctx, sessionId, service.ID, notices)

	logger.LogInfo("Response returned from OfferVehicles", sessionId)

	return
}

func WithdrawVehicle(ctx *gin.Context, sessionId, driverId, vehicleId string) (err error) {
	logger.LogInfo("Request received in WithdrawVehicle", sessionId)

	service, notices, err := withdrawVehicle(ctx, sessionId, driverId, vehicleId)
	if err != nil {
		err = utils.ClientError(sessionId, err, fmt.Sprintf(constants.Failed_To_Do_Job, "withdraw the vehicle"))
		return
	}

	sendNotices(ctx, sessionId, service.ID, notices)

	logger.LogInfo("Response returned from WithdrawVehicle", sessionId)

	return
}

func LeaveAsDriver(ctx *gin.Context, sessionId, driverId string) (err error) {
	logger.LogInfo("Request received in LeaveAsDriver", sessionId)

	service, notices, err := leaveAsDriver(ctx, sessionId, driverId)
	if err != nil {
		err = utils.ClientError(sessionId, err, fmt.Sprintf(constants.Failed_To_Do_Job, "leave the Pick & Drop service"))
		return
	}

	sendNotices(ctx, sessionId, service.ID, notices)

	logger.LogInfo("Response returned from LeaveAsDriver", sessionId)

	return
}

func RequestJoinAsPassenger(ctx *gin.Context, sessionId, passengerId string, request PassengerJoinRequest) (requestId string, err error) {
	logger.LogInfo("Request received in RequestJoinAsPassenger", sessionId)

	service, requestId, notices, err := requestJoinAsPassenger(ctx, sessionId, passengerId, request)
	if err != nil {
		err = utils.ClientError(sessionId, err, fmt.Sprintf(constants.Failed_To_Do_Job, "send the request"))
		return
	}

	sendNotices(ctx, sessionId, service.ID, notices)

	logger.LogInfo("Response returned from RequestJoinAsPassenger", sessionId)

	return
}

func GetPassengerMembership(ctx *gin.Context, sessionId, passengerId string) (data passengerMembershipData, err error) {
	logger.LogInfo("Request received in GetPassengerMembership", sessionId)

	data, err = getPassengerMembership(ctx, passengerId)
	if err != nil {
		logger.LogError(sessionId, "failed to get the membership error: "+err.Error())
		err = fmt.Errorf(constants.Failed_To_Do_Job, "get the membership")
		return
	}

	logger.LogInfo("Response returned from GetPassengerMembership", sessionId)

	return
}

func LeaveAsPassenger(ctx *gin.Context, sessionId, passengerId string) (err error) {
	logger.LogInfo("Request received in LeaveAsPassenger", sessionId)

	service, notices, err := leaveAsPassenger(ctx, sessionId, passengerId)
	if err != nil {
		err = utils.ClientError(sessionId, err, fmt.Sprintf(constants.Failed_To_Do_Job, "leave the Pick & Drop service"))
		return
	}

	sendNotices(ctx, sessionId, service.ID, notices)

	logger.LogInfo("Response returned from LeaveAsPassenger", sessionId)

	return
}

// GetJoinRequests lists one kind of request of the owner's service, the waiting ones
// unless another status is asked for.
func GetJoinRequests(ctx *gin.Context, sessionId, driverId, requestType, status string, page int) (resp JoinRequestsResponse, err error) {
	logger.LogInfo("Request received in GetJoinRequests", sessionId)

	service, err := requireOwnedService(ctx, sessionId, driverId)
	if err != nil {
		return
	}

	switch requestType {
	case constants.Request_Type_Driver:
		rows, vehicles, totalRows, e := getDriverRequests(ctx, service.ID, status, page)
		err = e
		resp = driverRequestsPage(rows, vehicles, totalRows)
	case constants.Request_Type_Vehicle:
		rows, totalRows, e := getVehicleRequests(ctx, service.ID, status, page)
		err = e
		resp = vehicleRequestsPage(rows, totalRows)
	default:
		rows, totalRows, e := getPassengerRequests(ctx, service.ID, status, page)
		err = e
		resp = passengerRequestsPage(rows, totalRows)
	}

	if err != nil {
		logger.LogError(sessionId, "failed to get the requests error: "+err.Error())
		err = fmt.Errorf(constants.Failed_To_Do_Job, "get the requests")
		return
	}

	logger.LogInfo("Response returned from GetJoinRequests", sessionId)

	return
}

func GetJoinRequest(ctx *gin.Context, sessionId, driverId, requestType, requestId string) (resp JoinRequestDetailResponse, err error) {
	logger.LogInfo("Request received in GetJoinRequest", sessionId)

	service, err := requireOwnedService(ctx, sessionId, driverId)
	if err != nil {
		return
	}

	resp.Type = requestType

	switch requestType {
	case constants.Request_Type_Driver:
		row, vehicles, e := getDriverRequest(ctx, service.ID, requestId)
		err = e
		resp.Request = mapDriverRequest(row, vehicles)
	case constants.Request_Type_Vehicle:
		row, e := getVehicleRequest(ctx, service.ID, requestId)
		err = e
		resp.Request = mapVehicleRequest(row)
	default:
		row, days, locations, e := getPassengerRequest(ctx, service.ID, requestId)
		err = e
		resp.Request = mapPassengerRequest(row, weeklyDemand(days, locations)[row.PassengerID])
	}

	if err != nil {
		if isNotFound(err) {
			err = errors.New(constants.Request_Not_Found)
		} else {
			logger.LogError(sessionId, "failed to get the request error: "+err.Error())
			err = fmt.Errorf(constants.Failed_To_Do_Job, "get the request")
		}
		logger.LogError(sessionId, err)
		return
	}

	logger.LogInfo("Response returned from GetJoinRequest", sessionId)

	return
}

// DecideJoinRequests settles a batch of requests. Drivers are decided before vehicles
// so a driver and their vehicle can be approved in one call, and every line runs in
// its own transaction, so one refusal never undoes the rest of the batch.
func DecideJoinRequests(ctx *gin.Context, sessionId, driverId string, request DecideRequestsRequest) (applied, skipped []DecisionResult, err error) {
	logger.LogInfo("Request received in DecideJoinRequests", sessionId)

	service, err := requireOwnedService(ctx, sessionId, driverId)
	if err != nil {
		return
	}

	order := map[string]int{
		constants.Request_Type_Driver:    0,
		constants.Request_Type_Vehicle:   1,
		constants.Request_Type_Passenger: 2,
	}

	ordered := make([]RequestDecision, 0, len(request.Decisions))
	for rank := 0; rank < 3; rank++ {
		for _, decision := range request.Decisions {
			if order[decision.Type] == rank {
				ordered = append(ordered, decision)
			}
		}
	}

	var notices []notice

	for _, decision := range ordered {
		result := DecisionResult{
			Type:      decision.Type,
			RequestId: decision.RequestId,
			Action:    decision.Action,
		}

		var decisionNotices []notice
		var e error

		switch decision.Type {
		case constants.Request_Type_Driver:
			decisionNotices, e = decideDriverRequest(ctx, service, driverId, decision.RequestId, decision.Action)
		case constants.Request_Type_Vehicle:
			decisionNotices, e = decideVehicleRequest(ctx, service, driverId, decision.RequestId, decision.Action)
		default:
			decisionNotices, e = decidePassengerRequest(ctx, service, driverId, decision.RequestId, decision.Action)
		}

		if e != nil {
			result.Reason = utils.ClientError(sessionId, e, constants.General_Error).Error()
			skipped = append(skipped, result)
			continue
		}

		result.Applied = true
		applied = append(applied, result)
		notices = append(notices, decisionNotices...)
	}

	sendNotices(ctx, sessionId, service.ID, notices)

	logger.LogInfo("Response returned from DecideJoinRequests", sessionId)
	logger.LogDebug2("Response returned from DecideJoinRequests", sessionId, fmt.Sprintf("applied: %d, skipped: %d", len(applied), len(skipped)))

	return
}

// GetAvailableDrivers lists who the owner can put behind the wheel, themselves first.
// With a schedule each driver also says whether they are free for it.
func GetAvailableDrivers(ctx *gin.Context, sessionId, driverId string, filter AvailabilityFilter, page int) (resp AvailableDriversResponse, err error) {
	logger.LogInfo("Request received in GetAvailableDrivers", sessionId)

	service, err := requireOwnedService(ctx, sessionId, driverId)
	if err != nil {
		return
	}

	rows, totalRows, err := getAvailableDrivers(ctx, service, page)
	if err != nil {
		logger.LogError(sessionId, "failed to get the drivers error: "+err.Error())
		err = fmt.Errorf(constants.Failed_To_Do_Job, "get the available drivers")
		return
	}

	driverIds := make([]string, 0, len(rows))
	for _, row := range rows {
		driverIds = append(driverIds, row.DriverID)
	}

	conflicts, err := findConflicts(ctx, filter, driverIds, nil, nil)
	if err != nil {
		logger.LogError(sessionId, "failed to check the drivers error: "+err.Error())
		err = fmt.Errorf(constants.Failed_To_Do_Job, "check the available drivers")
		return
	}

	resp = availableDriversPage(rows, conflicts, filter.Enabled(), totalRows)

	logger.LogInfo("Response returned from GetAvailableDrivers", sessionId)

	return
}

// GetAvailableVehicles lists the vehicles the owner can use. A vehicle without a
// valid seat count is listed but never available, it cannot be put on a shift.
func GetAvailableVehicles(ctx *gin.Context, sessionId, driverId string, filter AvailabilityFilter, page int) (resp AvailableVehiclesResponse, err error) {
	logger.LogInfo("Request received in GetAvailableVehicles", sessionId)

	service, err := requireOwnedService(ctx, sessionId, driverId)
	if err != nil {
		return
	}

	rows, totalRows, err := getAvailableVehicles(ctx, service.ID, page)
	if err != nil {
		logger.LogError(sessionId, "failed to get the vehicles error: "+err.Error())
		err = fmt.Errorf(constants.Failed_To_Do_Job, "get the available vehicles")
		return
	}

	vehicleIds := make([]string, 0, len(rows))
	for _, row := range rows {
		vehicleIds = append(vehicleIds, row.VehicleID)
	}

	conflicts, err := findConflicts(ctx, filter, nil, vehicleIds, nil)
	if err != nil {
		logger.LogError(sessionId, "failed to check the vehicles error: "+err.Error())
		err = fmt.Errorf(constants.Failed_To_Do_Job, "check the available vehicles")
		return
	}

	resp = availableVehiclesPage(rows, conflicts, filter.Enabled(), totalRows)

	logger.LogInfo("Response returned from GetAvailableVehicles", sessionId)

	return
}

// GetAvailablePassengers lists the approved passengers with their weekly demand, the
// sheet an owner builds shifts from.
func GetAvailablePassengers(ctx *gin.Context, sessionId, driverId string, filter AvailabilityFilter, dayOfWeek, page int) (resp AvailablePassengersResponse, err error) {
	logger.LogInfo("Request received in GetAvailablePassengers", sessionId)

	service, err := requireOwnedService(ctx, sessionId, driverId)
	if err != nil {
		return
	}

	rows, days, locations, totalRows, err := getAvailablePassengers(ctx, service.ID, dayOfWeek, page)
	if err != nil {
		logger.LogError(sessionId, "failed to get the passengers error: "+err.Error())
		err = fmt.Errorf(constants.Failed_To_Do_Job, "get the available passengers")
		return
	}

	passengerIds := make([]string, 0, len(rows))
	for _, row := range rows {
		passengerIds = append(passengerIds, row.PassengerID)
	}

	conflicts, err := findConflicts(ctx, filter, nil, nil, passengerIds)
	if err != nil {
		logger.LogError(sessionId, "failed to check the passengers error: "+err.Error())
		err = fmt.Errorf(constants.Failed_To_Do_Job, "check the available passengers")
		return
	}

	resp = availablePassengersPage(rows, weeklyDemand(days, locations), conflicts, filter.Enabled(), totalRows)

	logger.LogInfo("Response returned from GetAvailablePassengers", sessionId)

	return
}

func CreateAdvertisement(ctx *gin.Context, sessionId, driverId string, request AdvertisementRequest) (adId string, err error) {
	logger.LogInfo("Request received in CreateAdvertisement", sessionId)

	adId, err = createAdvertisement(ctx, sessionId, driverId, request)
	if err != nil {
		err = utils.ClientError(sessionId, err, fmt.Sprintf(constants.Creation_Failed, "advertisement"))
		return
	}

	logger.LogInfo("Response returned from CreateAdvertisement", sessionId)
	logger.LogDebug2("Response returned from CreateAdvertisement", sessionId, adId)

	return
}

func GetMyAdvertisements(ctx *gin.Context, sessionId, driverId string, page int) (resp MyAdvertisementsResponse, err error) {
	logger.LogInfo("Request received in GetMyAdvertisements", sessionId)

	service, err := requireOwnedService(ctx, sessionId, driverId)
	if err != nil {
		return
	}

	ads, locations, totalRows, err := queryAdvertisements(ctx, service.ID, AdvertisementSearch{}, page)
	if err != nil {
		logger.LogError(sessionId, "failed to get the advertisements error: "+err.Error())
		err = fmt.Errorf(constants.Failed_To_Do_Job, "get the advertisements")
		return
	}

	ctxDb, cancel := withTimeout(ctx)
	defer cancel()

	approvedVehicles, err := countApprovedVehicles(database.DatabaseConn.Postgres.WithContext(ctxDb), service.ID)
	if err != nil {
		logger.LogError(sessionId, "failed to count the vehicles error: "+err.Error())
		err = fmt.Errorf(constants.Failed_To_Do_Job, "get the advertisements")
		return
	}

	resp = MyAdvertisementsResponse{
		TotalPages:         utils.CalculatePagesize(totalRows),
		AdvertisementCount: totalRows,
		AdvertisementLimit: approvedVehicles * int64(constants.Advertisements_Per_Vehicle),
		Advertisements:     mapAdvertisements(ads, locations),
	}

	logger.LogInfo("Response returned from GetMyAdvertisements", sessionId)

	return
}

func DeleteAdvertisement(ctx *gin.Context, sessionId, driverId, adId string) (err error) {
	logger.LogInfo("Request received in DeleteAdvertisement", sessionId)

	if err = deleteAdvertisement(ctx, sessionId, driverId, adId); err != nil {
		err = utils.ClientError(sessionId, err, fmt.Sprintf(constants.DELETE_Failed, "advertisement"))
		return
	}

	logger.LogInfo("Response returned from DeleteAdvertisement", sessionId)

	return
}

// SearchAdvertisements is the public search passengers find a service through, it
// matches places along the whole advertised route the way the ride search does.
func SearchAdvertisements(ctx *gin.Context, sessionId string, filter AdvertisementSearch, page int) (ads []postgress.AdvertisementDetails, locations []postgress.PickDropAdvertisementLocation, totalRows int64, err error) {
	logger.LogInfo("Request received in SearchAdvertisements", sessionId)

	ads, locations, totalRows, err = queryAdvertisements(ctx, "", filter, page)
	if err != nil {
		logger.LogError(sessionId, "failed to search the advertisements error: "+err.Error())
		err = fmt.Errorf(constants.Failed_To_Do_Job, "search the advertisements")
		return
	}

	logger.LogInfo("Response returned from SearchAdvertisements", sessionId)

	return
}
