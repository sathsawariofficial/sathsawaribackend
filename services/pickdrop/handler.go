package pickdrop

import (
	"fmt"
	"net/http"
	"rideshare/pkgs/constants"
	"rideshare/pkgs/logger"
	"rideshare/pkgs/utils"

	"github.com/gin-gonic/gin"
	"github.com/rs/xid"
)

// an already registered driver enables Pick & Drop and becomes the owner of it
func EnableServiceHandler(ctx *gin.Context) {
	sessionId := xid.New().String()
	logger.LogInfo("Request received in EnableServiceHandler", sessionId)

	driverId := ctx.GetString(constants.User_KEY)

	var request EnableServiceRequest
	if err := ctx.ShouldBindJSON(&request); err != nil {
		logger.LogError(sessionId, "binding error: "+err.Error())
		ctx.JSON(http.StatusBadRequest, utils.APIResponse{
			Code:    http.StatusBadRequest,
			Message: fmt.Sprintf(constants.Creation_Failed, "Pick & Drop service"),
		})
		return
	}

	logger.LogDebug2("Request received in EnableServiceHandler", sessionId, request)

	if err := ValidateEnableService(&request); err != nil {
		logger.LogError(sessionId, "validation error: "+err.Error())
		ctx.JSON(http.StatusBadRequest, utils.APIResponse{
			Code:    http.StatusBadRequest,
			Message: err.Error(),
		})
		return
	}

	serviceId, err := EnableService(ctx, sessionId, driverId, request)
	if err != nil {
		logger.LogError(sessionId, "enable service error: "+err.Error())
		ctx.JSON(http.StatusBadRequest, utils.APIResponse{
			Code:    http.StatusBadRequest,
			Message: err.Error(),
		})
		return
	}

	logger.LogInfo("Response returned from EnableServiceHandler", sessionId)

	ctx.JSON(http.StatusOK, enableServiceResp(serviceId))
}

// the service the calling driver owns, belongs to or is waiting on
func GetMyServiceHandler(ctx *gin.Context) {
	sessionId := xid.New().String()
	logger.LogInfo("Request received in GetMyServiceHandler", sessionId)

	driverId := ctx.GetString(constants.User_KEY)

	data, err := GetMyService(ctx, sessionId, driverId)
	if err != nil {
		logger.LogError(sessionId, "get service error: "+err.Error())
		ctx.JSON(http.StatusBadRequest, utils.APIResponse{
			Code:    http.StatusBadRequest,
			Message: err.Error(),
		})
		return
	}

	logger.LogInfo("Response returned from GetMyServiceHandler", sessionId)

	ctx.JSON(http.StatusOK, myServiceResp(data))
}

// switches the owner's service off once no shift depends on it
func DisableServiceHandler(ctx *gin.Context) {
	sessionId := xid.New().String()
	logger.LogInfo("Request received in DisableServiceHandler", sessionId)

	driverId := ctx.GetString(constants.User_KEY)

	if err := DisableService(ctx, sessionId, driverId); err != nil {
		logger.LogError(sessionId, "disable service error: "+err.Error())
		ctx.JSON(http.StatusBadRequest, utils.APIResponse{
			Code:    http.StatusBadRequest,
			Message: err.Error(),
		})
		return
	}

	logger.LogInfo("Response returned from DisableServiceHandler", sessionId)

	ctx.JSON(http.StatusOK, utils.GeneralSuccessResp(fmt.Sprintf(constants.Success_Info, "Pick & Drop service disabled")))
}

// finds services to ask to join, served to drivers and passengers alike
func SearchServicesHandler(ctx *gin.Context) {
	sessionId := xid.New().String()
	logger.LogInfo("Request received in SearchServicesHandler", sessionId)

	userId := ctx.GetString(constants.User_KEY)
	search := ctx.DefaultQuery(constants.Search_Loc_Key, "")

	if len(search) > constants.General_Max_Len {
		err := fmt.Errorf("length of the search should not be more than %v characters", constants.General_Max_Len)
		logger.LogError(sessionId, "validation error: "+err.Error())
		ctx.JSON(http.StatusBadRequest, utils.APIResponse{
			Code:    http.StatusBadRequest,
			Message: err.Error(),
		})
		return
	}

	page, err := utils.GetPageNumber(ctx)
	if err != nil {
		logger.LogError(sessionId, "failed to get page number error: "+err.Error())
		ctx.JSON(http.StatusBadRequest, utils.APIResponse{
			Code:    http.StatusBadRequest,
			Message: fmt.Sprintf(constants.Invalid_Data, "page"),
		})
		return
	}

	services, totalRows, err := SearchServices(ctx, sessionId, userId, search, page)
	if err != nil {
		logger.LogError(sessionId, "search services error: "+err.Error())
		ctx.JSON(http.StatusBadRequest, utils.APIResponse{
			Code:    http.StatusBadRequest,
			Message: err.Error(),
		})
		return
	}

	logger.LogInfo("Response returned from SearchServicesHandler", sessionId)

	ctx.JSON(http.StatusOK, serviceSearchResp(services, totalRows))
}

// the owner puts more of their own vehicles into the service
func AddServiceVehiclesHandler(ctx *gin.Context) {
	sessionId := xid.New().String()
	logger.LogInfo("Request received in AddServiceVehiclesHandler", sessionId)

	driverId := ctx.GetString(constants.User_KEY)

	var request VehicleIdsRequest
	if err := ctx.ShouldBindJSON(&request); err != nil {
		logger.LogError(sessionId, "binding error: "+err.Error())
		ctx.JSON(http.StatusBadRequest, utils.APIResponse{
			Code:    http.StatusBadRequest,
			Message: fmt.Sprintf(constants.Failed_To_Do_Job, "add the vehicles"),
		})
		return
	}

	if err := ValidateVehicleIds(&request); err != nil {
		logger.LogError(sessionId, "validation error: "+err.Error())
		ctx.JSON(http.StatusBadRequest, utils.APIResponse{
			Code:    http.StatusBadRequest,
			Message: err.Error(),
		})
		return
	}

	if err := AddServiceVehicles(ctx, sessionId, driverId, request); err != nil {
		logger.LogError(sessionId, "add vehicles error: "+err.Error())
		ctx.JSON(http.StatusBadRequest, utils.APIResponse{
			Code:    http.StatusBadRequest,
			Message: err.Error(),
		})
		return
	}

	logger.LogInfo("Response returned from AddServiceVehiclesHandler", sessionId)

	ctx.JSON(http.StatusOK, utils.GeneralSuccessResp(fmt.Sprintf(constants.Added_Successfully, "Vehicles")))
}

// the owner takes a vehicle out of the service
func RemoveServiceVehicleHandler(ctx *gin.Context) {
	sessionId := xid.New().String()
	logger.LogInfo("Request received in RemoveServiceVehicleHandler", sessionId)

	driverId := ctx.GetString(constants.User_KEY)
	vehicleId := ctx.Query(constants.Vehicle_KEY)

	if err := ValidateId(vehicleId, "vehicle id"); err != nil {
		logger.LogError(sessionId, "validation error: "+err.Error())
		ctx.JSON(http.StatusBadRequest, utils.APIResponse{
			Code:    http.StatusBadRequest,
			Message: err.Error(),
		})
		return
	}

	if err := RemoveServiceVehicle(ctx, sessionId, driverId, vehicleId); err != nil {
		logger.LogError(sessionId, "remove vehicle error: "+err.Error())
		ctx.JSON(http.StatusBadRequest, utils.APIResponse{
			Code:    http.StatusBadRequest,
			Message: err.Error(),
		})
		return
	}

	logger.LogInfo("Response returned from RemoveServiceVehicleHandler", sessionId)

	ctx.JSON(http.StatusOK, utils.GeneralSuccessResp(fmt.Sprintf(constants.Success_Info, "Vehicle removed")))
}

// a driver asks to join a service, optionally proposing vehicles
func RequestJoinAsDriverHandler(ctx *gin.Context) {
	sessionId := xid.New().String()
	logger.LogInfo("Request received in RequestJoinAsDriverHandler", sessionId)

	driverId := ctx.GetString(constants.User_KEY)

	var request DriverJoinRequest
	if err := ctx.ShouldBindJSON(&request); err != nil {
		logger.LogError(sessionId, "binding error: "+err.Error())
		ctx.JSON(http.StatusBadRequest, utils.APIResponse{
			Code:    http.StatusBadRequest,
			Message: fmt.Sprintf(constants.Failed_To_Do_Job, "send the request"),
		})
		return
	}

	logger.LogDebug2("Request received in RequestJoinAsDriverHandler", sessionId, request)

	if err := ValidateDriverJoin(&request); err != nil {
		logger.LogError(sessionId, "validation error: "+err.Error())
		ctx.JSON(http.StatusBadRequest, utils.APIResponse{
			Code:    http.StatusBadRequest,
			Message: err.Error(),
		})
		return
	}

	requestId, err := RequestJoinAsDriver(ctx, sessionId, driverId, request)
	if err != nil {
		logger.LogError(sessionId, "join request error: "+err.Error())
		ctx.JSON(http.StatusBadRequest, utils.APIResponse{
			Code:    http.StatusBadRequest,
			Message: err.Error(),
		})
		return
	}

	logger.LogInfo("Response returned from RequestJoinAsDriverHandler", sessionId)

	ctx.JSON(http.StatusOK, joinRequestResp(requestId))
}

// a member driver offers more vehicles, each waiting for the owner
func OfferVehiclesHandler(ctx *gin.Context) {
	sessionId := xid.New().String()
	logger.LogInfo("Request received in OfferVehiclesHandler", sessionId)

	driverId := ctx.GetString(constants.User_KEY)

	var request VehicleIdsRequest
	if err := ctx.ShouldBindJSON(&request); err != nil {
		logger.LogError(sessionId, "binding error: "+err.Error())
		ctx.JSON(http.StatusBadRequest, utils.APIResponse{
			Code:    http.StatusBadRequest,
			Message: fmt.Sprintf(constants.Failed_To_Do_Job, "offer the vehicles"),
		})
		return
	}

	if err := ValidateVehicleIds(&request); err != nil {
		logger.LogError(sessionId, "validation error: "+err.Error())
		ctx.JSON(http.StatusBadRequest, utils.APIResponse{
			Code:    http.StatusBadRequest,
			Message: err.Error(),
		})
		return
	}

	if err := OfferVehicles(ctx, sessionId, driverId, request); err != nil {
		logger.LogError(sessionId, "offer vehicles error: "+err.Error())
		ctx.JSON(http.StatusBadRequest, utils.APIResponse{
			Code:    http.StatusBadRequest,
			Message: err.Error(),
		})
		return
	}

	logger.LogInfo("Response returned from OfferVehiclesHandler", sessionId)

	ctx.JSON(http.StatusOK, utils.GeneralSuccessResp(constants.Request_Received_Successfully))
}

// a driver takes back a vehicle offer, or an approved vehicle no shift depends on
func WithdrawVehicleHandler(ctx *gin.Context) {
	sessionId := xid.New().String()
	logger.LogInfo("Request received in WithdrawVehicleHandler", sessionId)

	driverId := ctx.GetString(constants.User_KEY)
	vehicleId := ctx.Query(constants.Vehicle_KEY)

	if err := ValidateId(vehicleId, "vehicle id"); err != nil {
		logger.LogError(sessionId, "validation error: "+err.Error())
		ctx.JSON(http.StatusBadRequest, utils.APIResponse{
			Code:    http.StatusBadRequest,
			Message: err.Error(),
		})
		return
	}

	if err := WithdrawVehicle(ctx, sessionId, driverId, vehicleId); err != nil {
		logger.LogError(sessionId, "withdraw vehicle error: "+err.Error())
		ctx.JSON(http.StatusBadRequest, utils.APIResponse{
			Code:    http.StatusBadRequest,
			Message: err.Error(),
		})
		return
	}

	logger.LogInfo("Response returned from WithdrawVehicleHandler", sessionId)

	ctx.JSON(http.StatusOK, utils.GeneralSuccessResp(fmt.Sprintf(constants.Success_Info, "Vehicle withdrawn")))
}

// a driver leaves their service, or withdraws the request still waiting
func LeaveServiceAsDriverHandler(ctx *gin.Context) {
	sessionId := xid.New().String()
	logger.LogInfo("Request received in LeaveServiceAsDriverHandler", sessionId)

	driverId := ctx.GetString(constants.User_KEY)

	if err := LeaveAsDriver(ctx, sessionId, driverId); err != nil {
		logger.LogError(sessionId, "leave service error: "+err.Error())
		ctx.JSON(http.StatusBadRequest, utils.APIResponse{
			Code:    http.StatusBadRequest,
			Message: err.Error(),
		})
		return
	}

	logger.LogInfo("Response returned from LeaveServiceAsDriverHandler", sessionId)

	ctx.JSON(http.StatusOK, utils.GeneralSuccessResp(fmt.Sprintf(constants.Success_Info, "Left the Pick & Drop service")))
}

// a passenger asks to join a service
func RequestJoinAsPassengerHandler(ctx *gin.Context) {
	sessionId := xid.New().String()
	logger.LogInfo("Request received in RequestJoinAsPassengerHandler", sessionId)

	passengerId := ctx.GetString(constants.User_KEY)

	var request PassengerJoinRequest
	if err := ctx.ShouldBindJSON(&request); err != nil {
		logger.LogError(sessionId, "binding error: "+err.Error())
		ctx.JSON(http.StatusBadRequest, utils.APIResponse{
			Code:    http.StatusBadRequest,
			Message: fmt.Sprintf(constants.Failed_To_Do_Job, "send the request"),
		})
		return
	}

	if err := ValidatePassengerJoin(&request); err != nil {
		logger.LogError(sessionId, "validation error: "+err.Error())
		ctx.JSON(http.StatusBadRequest, utils.APIResponse{
			Code:    http.StatusBadRequest,
			Message: err.Error(),
		})
		return
	}

	requestId, err := RequestJoinAsPassenger(ctx, sessionId, passengerId, request)
	if err != nil {
		logger.LogError(sessionId, "join request error: "+err.Error())
		ctx.JSON(http.StatusBadRequest, utils.APIResponse{
			Code:    http.StatusBadRequest,
			Message: err.Error(),
		})
		return
	}

	logger.LogInfo("Response returned from RequestJoinAsPassengerHandler", sessionId)

	ctx.JSON(http.StatusOK, joinRequestResp(requestId))
}

// where a passenger's membership stands
func GetPassengerMembershipHandler(ctx *gin.Context) {
	sessionId := xid.New().String()
	logger.LogInfo("Request received in GetPassengerMembershipHandler", sessionId)

	passengerId := ctx.GetString(constants.User_KEY)

	data, err := GetPassengerMembership(ctx, sessionId, passengerId)
	if err != nil {
		logger.LogError(sessionId, "get membership error: "+err.Error())
		ctx.JSON(http.StatusBadRequest, utils.APIResponse{
			Code:    http.StatusBadRequest,
			Message: err.Error(),
		})
		return
	}

	logger.LogInfo("Response returned from GetPassengerMembershipHandler", sessionId)

	ctx.JSON(http.StatusOK, passengerMembershipResp(data))
}

// a passenger leaves their service, or withdraws the request still waiting
func LeaveServiceAsPassengerHandler(ctx *gin.Context) {
	sessionId := xid.New().String()
	logger.LogInfo("Request received in LeaveServiceAsPassengerHandler", sessionId)

	passengerId := ctx.GetString(constants.User_KEY)

	if err := LeaveAsPassenger(ctx, sessionId, passengerId); err != nil {
		logger.LogError(sessionId, "leave service error: "+err.Error())
		ctx.JSON(http.StatusBadRequest, utils.APIResponse{
			Code:    http.StatusBadRequest,
			Message: err.Error(),
		})
		return
	}

	logger.LogInfo("Response returned from LeaveServiceAsPassengerHandler", sessionId)

	ctx.JSON(http.StatusOK, utils.GeneralSuccessResp(fmt.Sprintf(constants.Success_Info, "Left the Pick & Drop service")))
}

// the owner's join requests of one kind, the waiting ones by default
func GetJoinRequestsHandler(ctx *gin.Context) {
	sessionId := xid.New().String()
	logger.LogInfo("Request received in GetJoinRequestsHandler", sessionId)

	driverId := ctx.GetString(constants.User_KEY)
	requestType := ctx.Query(constants.Type_Key)
	status := ctx.DefaultQuery(constants.Status_Key, "")

	if err := ValidateRequestType(requestType); err != nil {
		logger.LogError(sessionId, "validation error: "+err.Error())
		ctx.JSON(http.StatusBadRequest, utils.APIResponse{
			Code:    http.StatusBadRequest,
			Message: err.Error(),
		})
		return
	}

	if err := ValidateRequestStatus(status); err != nil {
		logger.LogError(sessionId, "validation error: "+err.Error())
		ctx.JSON(http.StatusBadRequest, utils.APIResponse{
			Code:    http.StatusBadRequest,
			Message: err.Error(),
		})
		return
	}

	page, err := utils.GetPageNumber(ctx)
	if err != nil {
		logger.LogError(sessionId, "failed to get page number error: "+err.Error())
		ctx.JSON(http.StatusBadRequest, utils.APIResponse{
			Code:    http.StatusBadRequest,
			Message: fmt.Sprintf(constants.Invalid_Data, "page"),
		})
		return
	}

	resp, err := GetJoinRequests(ctx, sessionId, driverId, requestType, status, page)
	if err != nil {
		logger.LogError(sessionId, "get requests error: "+err.Error())
		ctx.JSON(http.StatusBadRequest, utils.APIResponse{
			Code:    http.StatusBadRequest,
			Message: err.Error(),
		})
		return
	}

	logger.LogInfo("Response returned from GetJoinRequestsHandler", sessionId)

	ctx.JSON(http.StatusOK, joinRequestsResp(resp))
}

// one join request with what the owner needs to decide on it
func GetJoinRequestHandler(ctx *gin.Context) {
	sessionId := xid.New().String()
	logger.LogInfo("Request received in GetJoinRequestHandler", sessionId)

	driverId := ctx.GetString(constants.User_KEY)
	requestType := ctx.Query(constants.Type_Key)
	requestId := ctx.Query(constants.Request_Key)

	if err := ValidateRequestType(requestType); err != nil {
		logger.LogError(sessionId, "validation error: "+err.Error())
		ctx.JSON(http.StatusBadRequest, utils.APIResponse{
			Code:    http.StatusBadRequest,
			Message: err.Error(),
		})
		return
	}

	if err := ValidateId(requestId, "request id"); err != nil {
		logger.LogError(sessionId, "validation error: "+err.Error())
		ctx.JSON(http.StatusBadRequest, utils.APIResponse{
			Code:    http.StatusBadRequest,
			Message: err.Error(),
		})
		return
	}

	resp, err := GetJoinRequest(ctx, sessionId, driverId, requestType, requestId)
	if err != nil {
		logger.LogError(sessionId, "get request error: "+err.Error())
		ctx.JSON(http.StatusBadRequest, utils.APIResponse{
			Code:    http.StatusBadRequest,
			Message: err.Error(),
		})
		return
	}

	logger.LogInfo("Response returned from GetJoinRequestHandler", sessionId)

	ctx.JSON(http.StatusOK, joinRequestDetailResp(resp))
}

// the owner approves, rejects or removes in one call
func DecideJoinRequestsHandler(ctx *gin.Context) {
	sessionId := xid.New().String()
	logger.LogInfo("Request received in DecideJoinRequestsHandler", sessionId)

	driverId := ctx.GetString(constants.User_KEY)

	var request DecideRequestsRequest
	if err := ctx.ShouldBindJSON(&request); err != nil {
		logger.LogError(sessionId, "binding error: "+err.Error())
		ctx.JSON(http.StatusBadRequest, utils.APIResponse{
			Code:    http.StatusBadRequest,
			Message: fmt.Sprintf(constants.Failed_To_Do_Job, "apply the decisions"),
		})
		return
	}

	logger.LogDebug2("Request received in DecideJoinRequestsHandler", sessionId, request)

	if err := ValidateDecideRequests(&request); err != nil {
		logger.LogError(sessionId, "validation error: "+err.Error())
		ctx.JSON(http.StatusBadRequest, utils.APIResponse{
			Code:    http.StatusBadRequest,
			Message: err.Error(),
		})
		return
	}

	applied, skipped, err := DecideJoinRequests(ctx, sessionId, driverId, request)
	if err != nil {
		logger.LogError(sessionId, "decide requests error: "+err.Error())
		ctx.JSON(http.StatusBadRequest, utils.APIResponse{
			Code:    http.StatusBadRequest,
			Message: err.Error(),
		})
		return
	}

	logger.LogInfo("Response returned from DecideJoinRequestsHandler", sessionId)

	ctx.JSON(http.StatusOK, decideRequestsResp(applied, skipped))
}

func readAvailabilityFilter(ctx *gin.Context) (AvailabilityFilter, error) {
	return ParseAvailabilityFilter(
		ctx.DefaultQuery(constants.Days_Of_Week_Key, ""),
		ctx.DefaultQuery(constants.Start_Date_Key, ""),
		ctx.DefaultQuery(constants.End_Date_Key, ""),
		ctx.DefaultQuery(constants.Start_Time_Key, ""),
		ctx.DefaultQuery(constants.End_Time_Key, ""),
		ctx.DefaultQuery(constants.Exclude_Shift_Key, ""),
	)
}

// the drivers the owner can put on a shift, themselves first
func GetAvailableDriversHandler(ctx *gin.Context) {
	sessionId := xid.New().String()
	logger.LogInfo("Request received in GetAvailableDriversHandler", sessionId)

	driverId := ctx.GetString(constants.User_KEY)

	filter, err := readAvailabilityFilter(ctx)
	if err != nil {
		logger.LogError(sessionId, "validation error: "+err.Error())
		ctx.JSON(http.StatusBadRequest, utils.APIResponse{
			Code:    http.StatusBadRequest,
			Message: err.Error(),
		})
		return
	}

	page, err := utils.GetPageNumber(ctx)
	if err != nil {
		logger.LogError(sessionId, "failed to get page number error: "+err.Error())
		ctx.JSON(http.StatusBadRequest, utils.APIResponse{
			Code:    http.StatusBadRequest,
			Message: fmt.Sprintf(constants.Invalid_Data, "page"),
		})
		return
	}

	resp, err := GetAvailableDrivers(ctx, sessionId, driverId, filter, page)
	if err != nil {
		logger.LogError(sessionId, "get drivers error: "+err.Error())
		ctx.JSON(http.StatusBadRequest, utils.APIResponse{
			Code:    http.StatusBadRequest,
			Message: err.Error(),
		})
		return
	}

	logger.LogInfo("Response returned from GetAvailableDriversHandler", sessionId)

	ctx.JSON(http.StatusOK, availableDriversResp(resp))
}

// the vehicles the owner can put on a shift
func GetAvailableVehiclesHandler(ctx *gin.Context) {
	sessionId := xid.New().String()
	logger.LogInfo("Request received in GetAvailableVehiclesHandler", sessionId)

	driverId := ctx.GetString(constants.User_KEY)

	filter, err := readAvailabilityFilter(ctx)
	if err != nil {
		logger.LogError(sessionId, "validation error: "+err.Error())
		ctx.JSON(http.StatusBadRequest, utils.APIResponse{
			Code:    http.StatusBadRequest,
			Message: err.Error(),
		})
		return
	}

	page, err := utils.GetPageNumber(ctx)
	if err != nil {
		logger.LogError(sessionId, "failed to get page number error: "+err.Error())
		ctx.JSON(http.StatusBadRequest, utils.APIResponse{
			Code:    http.StatusBadRequest,
			Message: fmt.Sprintf(constants.Invalid_Data, "page"),
		})
		return
	}

	resp, err := GetAvailableVehicles(ctx, sessionId, driverId, filter, page)
	if err != nil {
		logger.LogError(sessionId, "get vehicles error: "+err.Error())
		ctx.JSON(http.StatusBadRequest, utils.APIResponse{
			Code:    http.StatusBadRequest,
			Message: err.Error(),
		})
		return
	}

	logger.LogInfo("Response returned from GetAvailableVehiclesHandler", sessionId)

	ctx.JSON(http.StatusOK, availableVehiclesResp(resp))
}

// the approved passengers with their weekly demand
func GetAvailablePassengersHandler(ctx *gin.Context) {
	sessionId := xid.New().String()
	logger.LogInfo("Request received in GetAvailablePassengersHandler", sessionId)

	driverId := ctx.GetString(constants.User_KEY)
	dayOfWeek := utils.ToInt(ctx.DefaultQuery(constants.Day_Of_Week_Key, ""))

	if dayOfWeek != 0 && (dayOfWeek < constants.Day_Of_Week_Min_Value || dayOfWeek > constants.Day_Of_Week_Max_Value) {
		err := fmt.Errorf(constants.Invalid_Data, "day of week, use 1 for Monday to 7 for Sunday")
		logger.LogError(sessionId, "validation error: "+err.Error())
		ctx.JSON(http.StatusBadRequest, utils.APIResponse{
			Code:    http.StatusBadRequest,
			Message: err.Error(),
		})
		return
	}

	filter, err := readAvailabilityFilter(ctx)
	if err != nil {
		logger.LogError(sessionId, "validation error: "+err.Error())
		ctx.JSON(http.StatusBadRequest, utils.APIResponse{
			Code:    http.StatusBadRequest,
			Message: err.Error(),
		})
		return
	}

	page, err := utils.GetPageNumber(ctx)
	if err != nil {
		logger.LogError(sessionId, "failed to get page number error: "+err.Error())
		ctx.JSON(http.StatusBadRequest, utils.APIResponse{
			Code:    http.StatusBadRequest,
			Message: fmt.Sprintf(constants.Invalid_Data, "page"),
		})
		return
	}

	resp, err := GetAvailablePassengers(ctx, sessionId, driverId, filter, dayOfWeek, page)
	if err != nil {
		logger.LogError(sessionId, "get passengers error: "+err.Error())
		ctx.JSON(http.StatusBadRequest, utils.APIResponse{
			Code:    http.StatusBadRequest,
			Message: err.Error(),
		})
		return
	}

	logger.LogInfo("Response returned from GetAvailablePassengersHandler", sessionId)

	ctx.JSON(http.StatusOK, availablePassengersResp(resp))
}

// the owner advertises a route, two per approved vehicle at most
func CreateAdvertisementHandler(ctx *gin.Context) {
	sessionId := xid.New().String()
	logger.LogInfo("Request received in CreateAdvertisementHandler", sessionId)

	driverId := ctx.GetString(constants.User_KEY)

	var request AdvertisementRequest
	if err := ctx.ShouldBindJSON(&request); err != nil {
		logger.LogError(sessionId, "binding error: "+err.Error())
		ctx.JSON(http.StatusBadRequest, utils.APIResponse{
			Code:    http.StatusBadRequest,
			Message: fmt.Sprintf(constants.Creation_Failed, "advertisement"),
		})
		return
	}

	logger.LogDebug2("Request received in CreateAdvertisementHandler", sessionId, request)

	if err := ValidateAdvertisement(&request); err != nil {
		logger.LogError(sessionId, "validation error: "+err.Error())
		ctx.JSON(http.StatusBadRequest, utils.APIResponse{
			Code:    http.StatusBadRequest,
			Message: err.Error(),
		})
		return
	}

	adId, err := CreateAdvertisement(ctx, sessionId, driverId, request)
	if err != nil {
		logger.LogError(sessionId, "create advertisement error: "+err.Error())
		ctx.JSON(http.StatusBadRequest, utils.APIResponse{
			Code:    http.StatusBadRequest,
			Message: err.Error(),
		})
		return
	}

	logger.LogInfo("Response returned from CreateAdvertisementHandler", sessionId)

	ctx.JSON(http.StatusOK, advertisementResp(adId))
}

// the owner's advertisements with how many more are allowed
func GetMyAdvertisementsHandler(ctx *gin.Context) {
	sessionId := xid.New().String()
	logger.LogInfo("Request received in GetMyAdvertisementsHandler", sessionId)

	driverId := ctx.GetString(constants.User_KEY)

	page, err := utils.GetPageNumber(ctx)
	if err != nil {
		logger.LogError(sessionId, "failed to get page number error: "+err.Error())
		ctx.JSON(http.StatusBadRequest, utils.APIResponse{
			Code:    http.StatusBadRequest,
			Message: fmt.Sprintf(constants.Invalid_Data, "page"),
		})
		return
	}

	resp, err := GetMyAdvertisements(ctx, sessionId, driverId, page)
	if err != nil {
		logger.LogError(sessionId, "get advertisements error: "+err.Error())
		ctx.JSON(http.StatusBadRequest, utils.APIResponse{
			Code:    http.StatusBadRequest,
			Message: err.Error(),
		})
		return
	}

	logger.LogInfo("Response returned from GetMyAdvertisementsHandler", sessionId)

	ctx.JSON(http.StatusOK, myAdvertisementsResp(resp))
}

// the owner deletes an advertisement, it is archived first
func DeleteAdvertisementHandler(ctx *gin.Context) {
	sessionId := xid.New().String()
	logger.LogInfo("Request received in DeleteAdvertisementHandler", sessionId)

	driverId := ctx.GetString(constants.User_KEY)
	adId := ctx.Query(constants.Advertisement_Key)

	if err := ValidateId(adId, "advertisement id"); err != nil {
		logger.LogError(sessionId, "validation error: "+err.Error())
		ctx.JSON(http.StatusBadRequest, utils.APIResponse{
			Code:    http.StatusBadRequest,
			Message: err.Error(),
		})
		return
	}

	if err := DeleteAdvertisement(ctx, sessionId, driverId, adId); err != nil {
		logger.LogError(sessionId, "delete advertisement error: "+err.Error())
		ctx.JSON(http.StatusBadRequest, utils.APIResponse{
			Code:    http.StatusBadRequest,
			Message: err.Error(),
		})
		return
	}

	logger.LogInfo("Response returned from DeleteAdvertisementHandler", sessionId)

	ctx.JSON(http.StatusOK, utils.GeneralSuccessResp(fmt.Sprintf(constants.Success_Info, "Advertisement deleted")))
}

// the public search passengers find Pick & Drop services through
func SearchAdvertisementsHandler(ctx *gin.Context) {
	sessionId := xid.New().String()
	logger.LogInfo("Request received in SearchAdvertisementsHandler", sessionId)

	filter, err := ParseAdvertisementSearch(
		ctx.DefaultQuery(constants.Search_Loc_Key, ""),
		ctx.DefaultQuery(constants.Day_Of_Week_Key, ""),
		ctx.DefaultQuery(constants.Start_Time_Key, ""),
		ctx.DefaultQuery(constants.End_Time_Key, ""),
	)
	if err != nil {
		logger.LogError(sessionId, "validation error: "+err.Error())
		ctx.JSON(http.StatusBadRequest, utils.APIResponse{
			Code:    http.StatusBadRequest,
			Message: err.Error(),
		})
		return
	}

	page, err := utils.GetPageNumber(ctx)
	if err != nil {
		logger.LogError(sessionId, "failed to get page number error: "+err.Error())
		ctx.JSON(http.StatusBadRequest, utils.APIResponse{
			Code:    http.StatusBadRequest,
			Message: fmt.Sprintf(constants.Invalid_Data, "page"),
		})
		return
	}

	ads, locations, totalRows, err := SearchAdvertisements(ctx, sessionId, filter, page)
	if err != nil {
		logger.LogError(sessionId, "search advertisements error: "+err.Error())
		ctx.JSON(http.StatusBadRequest, utils.APIResponse{
			Code:    http.StatusBadRequest,
			Message: err.Error(),
		})
		return
	}

	logger.LogInfo("Response returned from SearchAdvertisementsHandler", sessionId)

	ctx.JSON(http.StatusOK, advertisementSearchResp(ads, locations, totalRows))
}
