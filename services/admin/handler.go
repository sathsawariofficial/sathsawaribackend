package admin

import (
	"fmt"
	"net/http"
	"rideshare/pkgs/constants"
	"rideshare/pkgs/logger"
	"rideshare/pkgs/utils"

	"github.com/gin-gonic/gin"
	"github.com/rs/xid"
)

// create a session for admin
func LoginAdminHandler(ctx *gin.Context) {
	sessionId := xid.New().String()
	logger.LogInfo("Request received in LoginAdminHandler", sessionId)

	var request AdminLoginRequest
	if err := ctx.ShouldBindJSON(&request); err != nil {
		logger.LogError(sessionId, "binding error: "+err.Error())
		ctx.JSON(http.StatusBadRequest, utils.APIResponse{
			Code:    http.StatusBadRequest,
			Message: constants.General_Error,
		})
		return
	}

	logger.LogDebug2("Request received in LoginAdminHandler", sessionId, request)

	err := ValidateAdminLogin(&request)
	if err != nil {
		logger.LogError(sessionId, "validation error: "+err.Error())
		ctx.JSON(http.StatusBadRequest, utils.APIResponse{
			Code:    http.StatusBadRequest,
			Message: err.Error(),
		})
		return
	}

	adminSessionId, admin, err := LoginAdmin(ctx, sessionId, request)
	if err != nil {
		logger.LogError(sessionId, "login error: "+err.Error())
		ctx.JSON(http.StatusBadRequest, utils.APIResponse{
			Code:    http.StatusBadRequest,
			Message: err.Error(),
		})
		return
	}

	adminLoginResp := loginAdminResp(adminSessionId, admin)

	logger.LogInfo("Response received in LoginAdminHandler", sessionId)
	logger.LogDebug2("Response received in LoginAdminHandler", sessionId, adminLoginResp)

	ctx.JSON(http.StatusOK, adminLoginResp)
}

func GetDriverDetailsHandler(ctx *gin.Context) {
	sessionId := xid.New().String()
	logger.LogInfo("Request received in GetDriverDetailsHandler", sessionId)

	page, err := utils.GetPageNumber(ctx)
	if err != nil {
		logger.LogError(sessionId, "binding error: "+err.Error())
		err = fmt.Errorf(constants.Invalid_Data, "page")
		ctx.JSON(http.StatusBadRequest, utils.APIResponse{
			Code:    http.StatusBadRequest,
			Message: err.Error(),
			Data: DriverDetailsResponse{
				Details: []DriverWithVehicle{},
			},
		})
		return
	}

	err = ValidatePage(page)
	if err != nil {
		logger.LogError(sessionId, "validation error: "+err.Error())
		err = fmt.Errorf(constants.Invalid_Data, "page")
		ctx.JSON(http.StatusBadRequest, utils.APIResponse{
			Code:    http.StatusBadRequest,
			Message: err.Error(),
			Data: DriverDetailsResponse{
				Details: []DriverWithVehicle{},
			},
		})
		return
	}

	driverDetails, totalRows, err := GetDrivers(ctx, sessionId, page)
	if err != nil {
		logger.LogError(sessionId, "get driver details error: "+err.Error())
		ctx.JSON(http.StatusBadRequest, utils.APIResponse{
			Code:    http.StatusBadRequest,
			Message: err.Error(),
			Data: DriverDetailsResponse{
				Details: []DriverWithVehicle{},
			},
		})
		return
	}

	if totalRows == 0 {
		logger.LogError(sessionId, "no drivers found error")
		ctx.JSON(http.StatusBadRequest, utils.APIResponse{
			Code:    http.StatusNoContent,
			Message: fmt.Sprintf(constants.Not_Found, "Driver"),
			Data: DriverDetailsResponse{
				Details: []DriverWithVehicle{},
			},
		})
		return
	}

	driverDetailsResp := driverDetailsResp(driverDetails, totalRows)

	logger.LogInfo("Response returned from GetDriverDetailsHandler", sessionId)
	logger.LogDebug2("Response returned from GetDriverDetailsHandler", sessionId, driverDetailsResp)

	ctx.JSON(http.StatusOK, driverDetailsResp)
}

func GetVehiclesHandler(ctx *gin.Context) {
	sessionId := xid.New().String()
	logger.LogInfo("Request received in GetVehiclesHandler", sessionId)

	page, err := utils.GetPageNumber(ctx)
	if err != nil {
		logger.LogError(sessionId, "binding error: "+err.Error())
		err = fmt.Errorf(constants.Invalid_Data, "page")
		ctx.JSON(http.StatusBadRequest, utils.APIResponse{
			Code:    http.StatusBadRequest,
			Message: err.Error(),
			Data: VehicleDetailsResponse{
				Vehicles: []Vehicle{},
			},
		})
		return
	}

	err = ValidatePage(page)
	if err != nil {
		logger.LogError(sessionId, "validation error: "+err.Error())
		err = fmt.Errorf(constants.Invalid_Data, "page")
		ctx.JSON(http.StatusBadRequest, utils.APIResponse{
			Code:    http.StatusBadRequest,
			Message: err.Error(),
			Data: VehicleDetailsResponse{
				Vehicles: []Vehicle{},
			},
		})
		return
	}

	vehicles, totalRows, err := GetVehicles(ctx, sessionId, page)
	if err != nil {
		logger.LogError(sessionId, "get vechile details error: "+err.Error())
		ctx.JSON(http.StatusBadRequest, utils.APIResponse{
			Code:    http.StatusBadRequest,
			Message: err.Error(),
			Data: VehicleDetailsResponse{
				Vehicles: []Vehicle{},
			},
		})
		return
	}

	if totalRows == 0 {
		logger.LogError(sessionId, "no vehicles found error")
		ctx.JSON(http.StatusBadRequest, utils.APIResponse{
			Code:    http.StatusNoContent,
			Message: fmt.Sprintf(constants.Not_Found, "Vehicles"),
			Data: VehicleDetailsResponse{
				Vehicles: []Vehicle{},
			},
		})
		return
	}

	vehicleDetailsResp := vechileDetailsResp(vehicles, totalRows)

	logger.LogInfo("Response returned from GetVehiclesHandler", sessionId)
	logger.LogDebug2("Response returned from GetVehiclesHandler", sessionId, vehicleDetailsResp)

	ctx.JSON(http.StatusOK, vehicleDetailsResp)
}

func GetRidesHandler(ctx *gin.Context) {
	sessionId := xid.New().String()
	logger.LogInfo("Request received in GetRidesHandler", sessionId)

	page, err := utils.GetPageNumber(ctx)
	if err != nil {
		logger.LogError(sessionId, "binding error: "+err.Error())
		err = fmt.Errorf(constants.Invalid_Data, "page")
		ctx.JSON(http.StatusBadRequest, utils.APIResponse{
			Code:    http.StatusBadRequest,
			Message: err.Error(),
			Data: RideDetailsResponse{
				Rides: []RideDetail{},
			},
		})
		return
	}

	err = ValidatePage(page)
	if err != nil {
		logger.LogError(sessionId, "validation error: "+err.Error())
		err = fmt.Errorf(constants.Invalid_Data, "page")
		ctx.JSON(http.StatusBadRequest, utils.APIResponse{
			Code:    http.StatusBadRequest,
			Message: err.Error(),
			Data: RideDetailsResponse{
				Rides: []RideDetail{},
			},
		})
		return
	}

	rides, totalRows, err := GetRides(ctx, sessionId, page)
	if err != nil {
		logger.LogError(sessionId, "get ride details error: "+err.Error())
		ctx.JSON(http.StatusBadRequest, utils.APIResponse{
			Code:    http.StatusBadRequest,
			Message: err.Error(),
			Data: RideDetailsResponse{
				Rides: []RideDetail{},
			},
		})
		return
	}

	if totalRows == 0 {
		logger.LogError(sessionId, "no rides found error")
		ctx.JSON(http.StatusBadRequest, utils.APIResponse{
			Code:    http.StatusNoContent,
			Message: fmt.Sprintf(constants.Not_Found, "Rides"),
			Data: RideDetailsResponse{
				Rides: []RideDetail{},
			},
		})
		return
	}

	rideDetailsResp := rideDetailsResp(rides, totalRows)

	logger.LogInfo("Response returned from GetRidesHandler", sessionId)
	logger.LogDebug2("Response returned from GetRidesHandler", sessionId, rideDetailsResp)

	ctx.JSON(http.StatusOK, rideDetailsResp)
}

func DeleteDriverHandler(ctx *gin.Context) {
	sessionId := xid.New().String()
	logger.LogInfo("Request received in DeleteDriverHandler", sessionId)

	err := DeleteDriver(ctx, sessionId)
	if err != nil {
		logger.LogError(sessionId, err)
		ctx.JSON(http.StatusBadRequest, utils.APIResponse{
			Code:    http.StatusBadRequest,
			Message: err.Error(),
			Data: DriverDetailsResponse{
				Details: []DriverWithVehicle{},
			},
		})
		return
	}

	logger.LogInfo("Response returned from DeleteDriverHandler", sessionId)
	logger.LogDebug2("Response returned from DeleteDriverHandler", sessionId, driverDetailsResp)

	ctx.JSON(http.StatusOK, utils.GeneralSuccessResp(constants.Success))
}

func AdminBroadcastHandler(ctx *gin.Context) {
	sessionId := xid.New().String()
	logger.LogInfo("Request received in AdminBroadcastHandler", sessionId)

	var request AdminBroadcastRequest
	if err := ctx.ShouldBindJSON(&request); err != nil {
		logger.LogError(sessionId, "binding error: "+err.Error())
		ctx.JSON(http.StatusBadRequest, utils.APIResponse{
			Code:    http.StatusBadRequest,
			Message: constants.General_Error,
		})
		return
	}

	logger.LogDebug2("Request received in AdminBroadcastHandler", sessionId, request)

	err := ValidateBoardcastRequest(&request)
	if err != nil {
		logger.LogError(sessionId, "validation error: "+err.Error())
		ctx.JSON(http.StatusBadRequest, utils.APIResponse{
			Code:    http.StatusBadRequest,
			Message: err.Error(),
		})
		return
	}

	err = CreateAdminBroadcastRequest(ctx, sessionId, request)
	if err != nil {
		logger.LogError(sessionId, "broadcast error: "+err.Error())
		ctx.JSON(http.StatusBadRequest, utils.APIResponse{
			Code:    http.StatusBadRequest,
			Message: err.Error(),
		})
		return
	}

	logger.LogInfo("Response returned from AdminBroadcastHandler", sessionId)

	ctx.JSON(http.StatusOK, utils.GeneralSuccessResp(constants.Success))
}

func AnnouncementHandler(ctx *gin.Context) {
	sessionId := xid.New().String()
	logger.LogInfo("Request received in AnnouncementHandler", sessionId)

	var request AnnouncementRequest
	if err := ctx.ShouldBindJSON(&request); err != nil {
		logger.LogError(sessionId, "binding error: "+err.Error())
		ctx.JSON(http.StatusBadRequest, utils.APIResponse{
			Code:    http.StatusBadRequest,
			Message: constants.General_Error,
		})
		return
	}

	logger.LogDebug2("Request received in AnnouncementHandler", sessionId, request)

	err := ValidateAnnouncementRequest(&request)
	if err != nil {
		logger.LogError(sessionId, "validation error: "+err.Error())
		ctx.JSON(http.StatusBadRequest, utils.APIResponse{
			Code:    http.StatusBadRequest,
			Message: err.Error(),
		})
		return
	}

	err = CreateAnnouncement(ctx, sessionId, request)
	if err != nil {
		logger.LogError(sessionId, "announcement creation error: "+err.Error())
		ctx.JSON(http.StatusBadRequest, utils.APIResponse{
			Code:    http.StatusBadRequest,
			Message: err.Error(),
		})
		return
	}

	logger.LogInfo("Response returned from AnnouncementHandler", sessionId)

	ctx.JSON(http.StatusOK, utils.GeneralSuccessResp(constants.Success))
}

func GetApprochRequestsHandler(ctx *gin.Context) {
	sessionId := xid.New().String()
	logger.LogInfo("Request received in GetApprochRequestsHandler", sessionId)

	approchType := ctx.Query(constants.Type_Key)
	page := ctx.Query(constants.Page_Key)

	err := ValidateApprochInfoReq(approchType, page)
	if err != nil {
		logger.LogError(sessionId, "validation error: "+err.Error())
		ctx.JSON(http.StatusBadRequest, utils.APIResponse{
			Code:    http.StatusBadRequest,
			Message: err.Error(),
		})
		return
	}

	approches, totalRows, err := GetApprochRequest(ctx, sessionId, approchType, utils.ToInt(page))
	if err != nil {
		logger.LogError(sessionId, err)
		ctx.JSON(http.StatusBadRequest, utils.APIResponse{
			Code:    http.StatusBadRequest,
			Message: err.Error(),
			Data: DriverDetailsResponse{
				Details: []DriverWithVehicle{},
			},
		})
		return
	}

	resp := createApprochInfoResp(approches, totalRows)

	logger.LogInfo("Response returned from GetApprochRequestsHandler", sessionId)
	logger.LogDebug2("Response returned from GetApprochRequestsHandler", sessionId, resp)

	ctx.JSON(http.StatusOK, resp)
}

// the whole platform counted in one call, the admin's first screen
func GetOverviewHandler(ctx *gin.Context) {
	sessionId := xid.New().String()
	logger.LogInfo("Request received in GetOverviewHandler", sessionId)

	overview, err := GetPlatformOverview(ctx, sessionId)
	if err != nil {
		logger.LogError(sessionId, "get overview error: "+err.Error())
		ctx.JSON(http.StatusBadRequest, utils.APIResponse{
			Code:    http.StatusBadRequest,
			Message: err.Error(),
		})
		return
	}

	overviewResp := platformOverviewResp(overview)

	logger.LogInfo("Response returned from GetOverviewHandler", sessionId)

	ctx.JSON(http.StatusOK, overviewResp)
}

// lists passenger accounts, searchable by name or number
func GetPassengersHandler(ctx *gin.Context) {
	sessionId := xid.New().String()
	logger.LogInfo("Request received in GetPassengersHandler", sessionId)

	search := ctx.DefaultQuery(constants.Search_Loc_Key, "")
	status := ctx.DefaultQuery(constants.Status_Key, "")

	if err := ValidateSearchFilters(status); err != nil {
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

	passengers, totalRows, err := GetPassengers(ctx, sessionId, search, status, page)
	if err != nil {
		logger.LogError(sessionId, "get passengers error: "+err.Error())
		ctx.JSON(http.StatusBadRequest, utils.APIResponse{
			Code:    http.StatusBadRequest,
			Message: err.Error(),
		})
		return
	}

	passengersResp := passengersResp(passengers, totalRows)

	logger.LogInfo("Response returned from GetPassengersHandler", sessionId)

	ctx.JSON(http.StatusOK, passengersResp)
}

// one passenger account
func GetPassengerProfileHandler(ctx *gin.Context) {
	sessionId := xid.New().String()
	logger.LogInfo("Request received in GetPassengerProfileHandler", sessionId)

	passengerId := ctx.Query(constants.Passenger_Key)

	if err := utils.ValidateId(passengerId); err != nil {
		logger.LogError(sessionId, "validation error: "+err.Error())
		ctx.JSON(http.StatusBadRequest, utils.APIResponse{
			Code:    http.StatusBadRequest,
			Message: fmt.Sprintf(constants.Invalid_Data, "passenger id"),
		})
		return
	}

	passenger, err := GetPassengerProfile(ctx, sessionId, passengerId)
	if err != nil {
		logger.LogError(sessionId, "get passenger error: "+err.Error())
		ctx.JSON(http.StatusBadRequest, utils.APIResponse{
			Code:    http.StatusBadRequest,
			Message: err.Error(),
		})
		return
	}

	profileResp := passengerProfileResp(passenger)

	logger.LogInfo("Response returned from GetPassengerProfileHandler", sessionId)

	ctx.JSON(http.StatusOK, profileResp)
}

// switches a passenger account on or off without destroying anything
func UpdatePassengerStatusHandler(ctx *gin.Context) {
	sessionId := xid.New().String()
	logger.LogInfo("Request received in UpdatePassengerStatusHandler", sessionId)

	adminId := ctx.GetString(constants.User_KEY)
	passengerId := ctx.Query(constants.Passenger_Key)

	var request AdminStatusRequest
	if err := ctx.ShouldBindJSON(&request); err != nil {
		logger.LogError(sessionId, "binding error: "+err.Error())
		ctx.JSON(http.StatusBadRequest, utils.APIResponse{
			Code:    http.StatusBadRequest,
			Message: fmt.Sprintf(constants.Update_Failed, "passenger"),
		})
		return
	}

	if err := ValidateAccountStatus(passengerId, &request); err != nil {
		logger.LogError(sessionId, "validation error: "+err.Error())
		ctx.JSON(http.StatusBadRequest, utils.APIResponse{
			Code:    http.StatusBadRequest,
			Message: err.Error(),
		})
		return
	}

	if err := UpdatePassengerStatus(ctx, sessionId, adminId, passengerId, request.Status); err != nil {
		logger.LogError(sessionId, "update passenger status error: "+err.Error())
		ctx.JSON(http.StatusBadRequest, utils.APIResponse{
			Code:    http.StatusBadRequest,
			Message: err.Error(),
		})
		return
	}

	logger.LogInfo("Response returned from UpdatePassengerStatusHandler", sessionId)

	ctx.JSON(http.StatusOK, utils.GeneralSuccessResp(fmt.Sprintf(constants.Updated_Successfully, "Passenger")))
}

// removes a passenger account and everything still pointing at it
func DeletePassengerHandler(ctx *gin.Context) {
	sessionId := xid.New().String()
	logger.LogInfo("Request received in DeletePassengerHandler", sessionId)

	adminId := ctx.GetString(constants.User_KEY)
	passengerId := ctx.Query(constants.Passenger_Key)

	if err := utils.ValidateId(passengerId); err != nil {
		logger.LogError(sessionId, "validation error: "+err.Error())
		ctx.JSON(http.StatusBadRequest, utils.APIResponse{
			Code:    http.StatusBadRequest,
			Message: fmt.Sprintf(constants.Invalid_Data, "passenger id"),
		})
		return
	}

	if err := DeletePassenger(ctx, sessionId, adminId, passengerId); err != nil {
		logger.LogError(sessionId, "delete passenger error: "+err.Error())
		ctx.JSON(http.StatusBadRequest, utils.APIResponse{
			Code:    http.StatusBadRequest,
			Message: err.Error(),
		})
		return
	}

	logger.LogInfo("Response returned from DeletePassengerHandler", sessionId)

	ctx.JSON(http.StatusOK, utils.GeneralSuccessResp(fmt.Sprintf(constants.Deleted_Successfully, "Passenger")))
}

// switches a driver account on or off, the softer alternative to deleting them
func UpdateDriverStatusHandler(ctx *gin.Context) {
	sessionId := xid.New().String()
	logger.LogInfo("Request received in UpdateDriverStatusHandler", sessionId)

	adminId := ctx.GetString(constants.User_KEY)
	driverId := ctx.Query(constants.Driver_Key)

	var request AdminStatusRequest
	if err := ctx.ShouldBindJSON(&request); err != nil {
		logger.LogError(sessionId, "binding error: "+err.Error())
		ctx.JSON(http.StatusBadRequest, utils.APIResponse{
			Code:    http.StatusBadRequest,
			Message: fmt.Sprintf(constants.Update_Failed, "driver"),
		})
		return
	}

	if err := ValidateAccountStatus(driverId, &request); err != nil {
		logger.LogError(sessionId, "validation error: "+err.Error())
		ctx.JSON(http.StatusBadRequest, utils.APIResponse{
			Code:    http.StatusBadRequest,
			Message: err.Error(),
		})
		return
	}

	if err := UpdateDriverStatus(ctx, sessionId, adminId, driverId, request.Status); err != nil {
		logger.LogError(sessionId, "update driver status error: "+err.Error())
		ctx.JSON(http.StatusBadRequest, utils.APIResponse{
			Code:    http.StatusBadRequest,
			Message: err.Error(),
		})
		return
	}

	logger.LogInfo("Response returned from UpdateDriverStatusHandler", sessionId)

	ctx.JSON(http.StatusOK, utils.GeneralSuccessResp(fmt.Sprintf(constants.Updated_Successfully, "Driver")))
}

////////////////////////////// PICK & DROP OVERSIGHT //////////////////////////////
// Read only: Pick & Drop and its shifts stay run by each service's own owner, the
// admin console only ever looks.

// every Pick & Drop service on the platform, active or disabled, with its roster
func GetPickDropServicesHandler(ctx *gin.Context) {
	sessionId := xid.New().String()
	logger.LogInfo("Request received in GetPickDropServicesHandler", sessionId)

	search := ctx.DefaultQuery(constants.Search_Loc_Key, "")
	status := ctx.DefaultQuery(constants.Status_Key, "")

	page, err := utils.GetPageNumber(ctx)
	if err != nil {
		logger.LogError(sessionId, "failed to get page number error: "+err.Error())
		ctx.JSON(http.StatusBadRequest, utils.APIResponse{
			Code:    http.StatusBadRequest,
			Message: fmt.Sprintf(constants.Invalid_Data, "page"),
		})
		return
	}

	services, totalRows, err := GetPickDropServices(ctx, sessionId, search, status, page)
	if err != nil {
		logger.LogError(sessionId, "get services error: "+err.Error())
		ctx.JSON(http.StatusBadRequest, utils.APIResponse{
			Code:    http.StatusBadRequest,
			Message: err.Error(),
		})
		return
	}

	logger.LogInfo("Response returned from GetPickDropServicesHandler", sessionId)

	ctx.JSON(http.StatusOK, servicesResp(services, totalRows))
}

// one service, its roster counts and every active shift it has
func GetPickDropServiceDetailHandler(ctx *gin.Context) {
	sessionId := xid.New().String()
	logger.LogInfo("Request received in GetPickDropServiceDetailHandler", sessionId)

	serviceId := ctx.Query(constants.Service_Key)
	if err := utils.ValidateId(serviceId); err != nil {
		logger.LogError(sessionId, "validation error: "+err.Error())
		ctx.JSON(http.StatusBadRequest, utils.APIResponse{
			Code:    http.StatusBadRequest,
			Message: err.Error(),
		})
		return
	}

	service, counts, shifts, err := GetPickDropServiceDetail(ctx, sessionId, serviceId)
	if err != nil {
		logger.LogError(sessionId, "get service error: "+err.Error())
		ctx.JSON(http.StatusBadRequest, utils.APIResponse{
			Code:    http.StatusBadRequest,
			Message: err.Error(),
		})
		return
	}

	logger.LogInfo("Response returned from GetPickDropServiceDetailHandler", sessionId)

	ctx.JSON(http.StatusOK, serviceDetailResp(service, counts, shifts))
}

// every shift on the platform, in any service, searched the same way an owner
// searches their own
func GetShiftsHandler(ctx *gin.Context) {
	sessionId := xid.New().String()
	logger.LogInfo("Request received in GetShiftsHandler", sessionId)

	filter, err := ParseAdminShiftFilter(
		ctx.DefaultQuery(constants.Status_Key, ""),
		ctx.DefaultQuery(constants.Day_Of_Week_Key, ""),
		ctx.DefaultQuery(constants.Search_Loc_Key, ""),
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

	shifts, totalRows, err := GetShifts(ctx, sessionId, filter, page)
	if err != nil {
		logger.LogError(sessionId, "get shifts error: "+err.Error())
		ctx.JSON(http.StatusBadRequest, utils.APIResponse{
			Code:    http.StatusBadRequest,
			Message: err.Error(),
		})
		return
	}

	logger.LogInfo("Response returned from GetShiftsHandler", sessionId)

	ctx.JSON(http.StatusOK, shiftsResp(shifts, totalRows))
}

// every open shift request on the platform, addressed to a service or still open to
// any of them
func GetShiftRequestsHandler(ctx *gin.Context) {
	sessionId := xid.New().String()
	logger.LogInfo("Request received in GetShiftRequestsHandler", sessionId)

	filter, err := ParseAdminShiftFilter(
		constants.Shift_Status_All,
		ctx.DefaultQuery(constants.Day_Of_Week_Key, ""),
		ctx.DefaultQuery(constants.Search_Loc_Key, ""),
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

	requests, totalRows, err := GetShiftRequests(ctx, sessionId, filter, page)
	if err != nil {
		logger.LogError(sessionId, "get shift requests error: "+err.Error())
		ctx.JSON(http.StatusBadRequest, utils.APIResponse{
			Code:    http.StatusBadRequest,
			Message: err.Error(),
		})
		return
	}

	logger.LogInfo("Response returned from GetShiftRequestsHandler", sessionId)

	ctx.JSON(http.StatusOK, shiftRequestsResp(requests, totalRows))
}

// every advertisement on the platform, whatever the status of the service that
// posted it
func GetAdvertisementsHandler(ctx *gin.Context) {
	sessionId := xid.New().String()
	logger.LogInfo("Request received in GetAdvertisementsHandler", sessionId)

	search := ctx.DefaultQuery(constants.Search_Loc_Key, "")

	page, err := utils.GetPageNumber(ctx)
	if err != nil {
		logger.LogError(sessionId, "failed to get page number error: "+err.Error())
		ctx.JSON(http.StatusBadRequest, utils.APIResponse{
			Code:    http.StatusBadRequest,
			Message: fmt.Sprintf(constants.Invalid_Data, "page"),
		})
		return
	}

	ads, totalRows, err := GetAdvertisements(ctx, sessionId, search, page)
	if err != nil {
		logger.LogError(sessionId, "get advertisements error: "+err.Error())
		ctx.JSON(http.StatusBadRequest, utils.APIResponse{
			Code:    http.StatusBadRequest,
			Message: err.Error(),
		})
		return
	}

	logger.LogInfo("Response returned from GetAdvertisementsHandler", sessionId)

	ctx.JSON(http.StatusOK, advertisementsResp(ads, totalRows))
}
